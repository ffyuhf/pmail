// Package ratelimit 提供集中式内存速率限制器，用于暴力破解防护。
//
// 所有协议认证端点（HTTP、SMTP、IMAP、POP3）共享此限制器，
// 以执行基于 IP 和基于账户的登录频率限制。
//
// 用法：
//
//	ratelimit.Check(ip, account)    // 检查 IP 或账户是否被锁定
//	ratelimit.WaitDelay(ip, account) // 认证前的指数退避延迟
//	ratelimit.RecordFailure(ip, account) // 记录登录失败
//	ratelimit.RecordSuccess(ip, account) // 登录成功后清除失败记录
package ratelimit

import (
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

// 速率限制阈值（编译时常量，用于安全敏感配置）
const (
	// MaxIPFailures 单个 IP 在被锁定前允许的最大失败尝试次数。
	MaxIPFailures = 5

	// IPLockoutDuration IP 超过 MaxIPFailures 后保持锁定的时长。
	IPLockoutDuration = 15 * time.Minute

	// MaxAccountFailures 单个账户在被锁定前允许的最大失败尝试次数。
	MaxAccountFailures = 10

	// AccountLockoutDuration 账户超过 MaxAccountFailures 后保持锁定的时长。
	AccountLockoutDuration = 30 * time.Minute

	// MaxDelaySeconds 指数退避延迟的上限（秒）。
	MaxDelaySeconds = 8

	// CleanupInterval 后台协程清理过期条目的间隔。
	CleanupInterval = 5 * time.Minute
)

// failureRecord 跟踪单个键（IP 或账户）的认证失败记录。
type failureRecord struct {
	failures    int       // 连续失败次数
	lockedUntil time.Time // 若非零，表示该键锁定至此时间
	lastAttempt time.Time // 最近一次失败的时间戳
}

// Limiter 是集中式速率限制器实例。
type Limiter struct {
	records sync.Map // map[string]*failureRecord，键为 "ip:<addr>" 或 "account:<name>"
	stopCh  chan struct{}
}

// globalInstance 是所有协议处理程序共享的单例。
var globalInstance *Limiter

// Init 初始化全局速率限制器并启动后台清理协程。
// 应在应用启动时调用一次。
func Init() {
	globalInstance = &Limiter{
		stopCh: make(chan struct{}),
	}
	go globalInstance.cleanupLoop()
	log.Info("Rate limiter initialized")
}

// ipKey 返回 IP 地址的 sync.Map 键。
func ipKey(ip string) string {
	return "ip:" + ip
}

// accountKey 返回账户名称的 sync.Map 键（转为小写以保持一致性）。
func accountKey(account string) string {
	return "account:" + strings.ToLower(account)
}

// getOrCreate 原子地加载或创建给定键的 failureRecord。
func (l *Limiter) getOrCreate(key string) *failureRecord {
	val, _ := l.records.LoadOrStore(key, &failureRecord{})
	return val.(*failureRecord)
}

// Check 验证给定的 IP 或账户当前是否被锁定。
// 如果允许认证则返回 nil，否则返回描述锁定状态的错误。
func Check(ip, account string) error {
	if globalInstance == nil {
		return nil
	}

	now := time.Now()

	// 检查 IP 级别锁定
	if ip != "" {
		rec := globalInstance.getOrCreate(ipKey(ip))
		if !rec.lockedUntil.IsZero() && now.Before(rec.lockedUntil) {
			remaining := time.Until(rec.lockedUntil).Round(time.Second)
			return fmt.Errorf("IP %s is locked due to too many failed attempts, try again in %v", ip, remaining)
		}
	}

	// 检查账户级别锁定
	if account != "" {
		rec := globalInstance.getOrCreate(accountKey(account))
		if !rec.lockedUntil.IsZero() && now.Before(rec.lockedUntil) {
			remaining := time.Until(rec.lockedUntil).Round(time.Second)
			return fmt.Errorf("account %s is locked due to too many failed attempts, try again in %v", account, remaining)
		}
	}

	return nil
}

// WaitDelay 根据给定 IP 和账户的失败次数阻塞一段指数退避时间。
// 这会随失败次数增加时间惩罚，从而减缓暴力破解攻击。
//
// 延迟计算公式为 min(2^failures, MaxDelaySeconds) 秒。
// 例如：1s, 2s, 4s, 8s, 8s, 8s...
func WaitDelay(ip, account string) {
	if globalInstance == nil {
		return
	}

	maxFailures := 0

	// 使用 IP 和账户中较高的失败次数
	if ip != "" {
		rec := globalInstance.getOrCreate(ipKey(ip))
		if rec.failures > maxFailures {
			maxFailures = rec.failures
		}
	}
	if account != "" {
		rec := globalInstance.getOrCreate(accountKey(account))
		if rec.failures > maxFailures {
			maxFailures = rec.failures
		}
	}

	if maxFailures > 0 {
		// 计算指数退避：2^failures 秒，上限为 MaxDelaySeconds
		delaySeconds := 1 << min(maxFailures, 3) // 1, 2, 4, 8
		if delaySeconds > MaxDelaySeconds {
			delaySeconds = MaxDelaySeconds
		}
		time.Sleep(time.Duration(delaySeconds) * time.Second)
	}
}

// RecordFailure 记录给定 IP 和账户的认证失败尝试。
// 如果失败次数超过阈值，IP 或账户将被锁定指定时长。
func RecordFailure(ip, account string) {
	if globalInstance == nil {
		return
	}

	now := time.Now()

	// 记录 IP 级别失败
	if ip != "" {
		rec := globalInstance.getOrCreate(ipKey(ip))
		rec.failures++
		rec.lastAttempt = now
		if rec.failures >= MaxIPFailures {
			rec.lockedUntil = now.Add(IPLockoutDuration)
			log.Warnf("Rate limit: IP %s locked for %v after %d failed attempts", ip, IPLockoutDuration, rec.failures)
		}
	}

	// 记录账户级别失败
	if account != "" {
		rec := globalInstance.getOrCreate(accountKey(account))
		rec.failures++
		rec.lastAttempt = now
		if rec.failures >= MaxAccountFailures {
			rec.lockedUntil = now.Add(AccountLockoutDuration)
			log.Warnf("Rate limit: account %s locked for %v after %d failed attempts", account, AccountLockoutDuration, rec.failures)
		}
	}
}

// RecordSuccess 清除给定 IP 和账户的所有失败记录，
// 在成功认证后重置速率限制状态。
func RecordSuccess(ip, account string) {
	if globalInstance == nil {
		return
	}

	if ip != "" {
		globalInstance.records.Delete(ipKey(ip))
	}
	if account != "" {
		globalInstance.records.Delete(accountKey(account))
	}
}

// cleanupLoop 定期移除过期条目，防止内存无限增长。
func (l *Limiter) cleanupLoop() {
	ticker := time.NewTicker(CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			now := time.Now()
			l.records.Range(func(key, value any) bool {
				rec := value.(*failureRecord)
				// 如果锁定已过期且无近期活动，则删除
				if !rec.lockedUntil.IsZero() && now.After(rec.lockedUntil) {
					l.records.Delete(key)
				}
				// 删除无锁定且超过 1 小时无活动的过期记录
				if rec.lockedUntil.IsZero() && now.Sub(rec.lastAttempt) > time.Hour {
					l.records.Delete(key)
				}
				return true
			})
		case <-l.stopCh:
			return
		}
	}
}

// min 返回两个整数中较小的一个（Go 1.21+ 有内置 min，但此函数确保兼容性和包内清晰性）。
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ExtractIP 从 "host:port" 格式的地址字符串中提取主机部分。
// 同时适用于 TCP 地址（"192.168.1.1:25"）和 net.Addr.String() 输出。
// 对于 HTTP 请求的远程地址，请改用 ExtractIPFromRequest。
func ExtractIP(addr string) string {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		// 如果 SplitHostPort 失败，地址可能不包含端口（例如裸 IP）
		return strings.TrimSpace(addr)
	}
	return strings.TrimSpace(host)
}
