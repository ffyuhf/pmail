package imap_server

import (
	"sync"
	"time"

	"github.com/emersion/go-imap/v2/imapserver"
	"github.com/ffyuhf/pmail/models"
	"github.com/ffyuhf/pmail/services/group"
	"github.com/ffyuhf/pmail/utils/context"
	pmailLog "github.com/ffyuhf/pmail/utils/log"
	"github.com/spf13/cast"
)

// idleConnection 记录单个 IDLE 连接的写入器和监听邮箱。
// 修复 BUG-2d（2026-06-09）：添加 mailbox 字段，支持按邮箱精确推送新邮件通知。
type idleConnection struct {
	writer  *imapserver.UpdateWriter
	mailbox string
}

// idleConnections 封装单个用户的所有 IDLE 连接，提供互斥保护。
// 修复 BUG-2c（2026-06-09）：使用 sync.Mutex 保护 writers map，避免并发读写 panic。
type idleConnections struct {
	mu      sync.Mutex
	writers map[string]idleConnection
}

var userConnects sync.Map

// userConnectsMu 保护 userConnects 的 Load/Store/Delete 操作。
// 修复 BUG-2c（2026-06-09）：添加全局互斥锁，确保对 sync.Map 的复合操作原子性。
var userConnectsMu sync.Mutex

// Idle 处理 IMAP IDLE 命令，保持连接并等待新邮件通知。
// 记录 IDLE 进入和结束事件，便于诊断长时间空闲连接问题。
//
// 修复历史：
//   - BUG-2a（2026-06-09）：修复错误键删除。原代码用 logId 删除 sync.Map 条目，
//     但存储键是 userId，导致空闲连接永远不会被清理（内存泄漏）。
//     修复后从内层 map 正确删除 logId，并在 map 为空时清理外层条目。
//   - BUG-2c（2026-06-09）：添加 mutex 保护并发读写。
//   - BUG-2d（2026-06-09）：记录 currentMailbox，支持按邮箱精确推送。
func (s *serverSession) Idle(w *imapserver.UpdateWriter, stop <-chan struct{}) error {
	userId := s.ctx.UserID
	logId := cast.ToString(s.ctx.GetValue(context.LogID))

	pmailLog.ImapDebugf(s.ctx, pmailLog.EventIMAPIdleEnter, "用户=%s 文件夹=%s", s.ctx.UserAccount, s.currentMailbox)

	// 注册 IDLE 连接
	userConnectsMu.Lock()
	connectsAny, ok := userConnects.Load(userId)
	if !ok {
		connectsAny = &idleConnections{writers: map[string]idleConnection{}}
		userConnects.Store(userId, connectsAny)
	}
	connects := connectsAny.(*idleConnections)

	connects.mu.Lock()
	connects.writers[logId] = idleConnection{writer: w, mailbox: s.currentMailbox}
	connects.mu.Unlock()
	userConnectsMu.Unlock()

	// IDLE 结束时清理连接
	idleStart := time.Now()
	go func() {
		<-stop
		duration := time.Since(idleStart)
		pmailLog.ImapInfof(s.ctx, pmailLog.EventIMAPIdleTimeout, "用户=%s 文件夹=%s 持续时间=%v", s.ctx.UserAccount, s.currentMailbox, duration)

		// 修复 BUG-2a：从内层 map 正确删除 logId 条目，而非错误地用 logId 删除 sync.Map
		userConnectsMu.Lock()
		connects.mu.Lock()
		delete(connects.writers, logId)
		if len(connects.writers) == 0 {
			userConnects.Delete(userId)
		}
		connects.mu.Unlock()
		userConnectsMu.Unlock()
	}()

	return nil
}

// IdleNotice 向指定用户的活跃 IMAP IDLE 连接推送新邮件通知。
//
// 修复历史：
//   - BUG-2b（2026-06-09）：修复硬编码消息数。原代码 WriteNumMessages(1) 始终推送 1，
//     导致客户端收到错误的 EXISTS 更新。修复后通过 idleNumMessages 查询实际消息数。
//   - BUG-2c（2026-06-09）：在锁保护下复制连接列表，释放锁后再推送，避免死锁。
//   - BUG-2d（2026-06-09）：每个连接携带 mailbox 信息，按邮箱查询对应消息数。
func IdleNotice(ctx *context.Context, userId int, email *models.Email) error {
	if userId == 0 || email == nil || email.Id == 0 {
		return nil
	}

	// 在锁保护下获取连接快照
	userConnectsMu.Lock()
	connectsAny, ok := userConnects.Load(userId)
	if !ok {
		userConnectsMu.Unlock()
		return nil
	}

	connects := connectsAny.(*idleConnections)
	connects.mu.Lock()
	idleConnects := make([]idleConnection, 0, len(connects.writers))
	for _, connect := range connects.writers {
		idleConnects = append(idleConnects, connect)
	}
	connects.mu.Unlock()
	userConnectsMu.Unlock()

	// 释放所有锁后再推送通知，避免死锁
	for _, connect := range idleConnects {
		connect.writer.WriteNumMessages(idleNumMessages(ctx, userId, connect.mailbox))
	}
	return nil
}

// idleNumMessages 查询指定用户在指定邮箱中的实际消息数。
// 修复 BUG-2b（2026-06-09）：替代硬编码的 WriteNumMessages(1)。
// 复用已有的 group.GetGroupStatus 函数查询 MESSAGES 计数。
func idleNumMessages(ctx *context.Context, userId int, mailbox string) uint32 {
	if mailbox == "" {
		mailbox = "INBOX"
	}
	noticeCtx := *ctx
	noticeCtx.UserID = userId
	_, data := group.GetGroupStatus(&noticeCtx, mailbox, []string{"MESSAGES"})
	return cast.ToUint32(data["MESSAGES"])
}
