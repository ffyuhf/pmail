package http_server

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	olog "log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ffyuhf/pmail/config"
	"github.com/ffyuhf/pmail/controllers"
	"github.com/ffyuhf/pmail/db"
	"github.com/ffyuhf/pmail/dto/response"
	"github.com/ffyuhf/pmail/i18n"
	"github.com/ffyuhf/pmail/models"
	"github.com/ffyuhf/pmail/session"
	"github.com/ffyuhf/pmail/utils/context"
	"github.com/ffyuhf/pmail/utils/id"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cast"
)

//go:embed dist/*
var local embed.FS

var httpsServer *http.Server

type nullWrite struct {
}

func (w *nullWrite) Write(p []byte) (int, error) {
	return len(p), nil
}

func HttpsStart() {

	mux := http.NewServeMux()

	router(mux)

	// go http server会打一堆没用的日志，写一个空的日志处理器，屏蔽掉日志输出
	nullLog := olog.New(&nullWrite{}, "", olog.Ldate)

	HttpsPort := 443
	if config.Instance.HttpsPort > 0 {
		HttpsPort = config.Instance.HttpsPort
	}

	if config.Instance.HttpsEnabled != 2 {
		log.Infof("[HTTPS] HTTPS服务启动 端口=%d", HttpsPort)
		httpsServer = &http.Server{
			Addr:         fmt.Sprintf(":%d", HttpsPort),
			Handler:      session.Instance.LoadAndSave(mux),
			ReadTimeout:  time.Second * 90,
			WriteTimeout: time.Second * 90,
			ErrorLog:     nullLog,
		}
		err := httpsServer.ListenAndServeTLS(config.Instance.SSLPublicKeyPath, config.Instance.SSLPrivateKeyPath)
		if err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				// 正常关闭（重启或停机）
				log.Infof("[HTTPS] HTTPS服务正常关闭 端口=%d", HttpsPort)
			} else {
				// 异常错误仍然打印并退出
				log.Errorf("[HTTPS] HTTPS服务异常 端口=%d 错误=%+v", HttpsPort, err)
				panic(err)
			}
		}
	}
}

func HttpsStop() {
	if httpsServer != nil {
		httpsServer.Close()
	}
}

// 新增：分类处理关闭错误

// 注入context
func contextIterceptor(h controllers.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if w.Header().Get("Content-Type") == "" {
			w.Header().Set("Content-Type", "application/json")
		}

		// HTTP 安全响应头：防止 MIME 嗅探、点击劫持、XSS 攻击
		// 修改日期: 20260425
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		// Strict-Transport-Security 仅在实际 HTTPS 连接上设置（r.TLS != nil）
		// 在 HTTP 响应中发送 HSTS 头无效且不合规，因此需要条件判断
		if r.TLS != nil {
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		// CSRF 防护：对写操作校验请求来源，阻止跨站请求伪造
		// 仅检查 Origin/Referer，同源请求自动通过，跨站请求被拒绝
		// 修改日期: 20260425
		// 修改日期: 20260504，增加 config.IsInit 前置判断：
		// setup 阶段（config.IsInit==false）域名尚未配置，allowedDomains 为空，
		// isSameOrigin 必然返回 false，导致所有 POST 请求被拦截、setup API 无法执行。
		// setup 阶段由 SetupToken 鉴权保护，跳过 CSRF 不降低安全性。
		if config.IsInit && r.Method != "GET" && r.Method != "HEAD" && r.Method != "OPTIONS" {
			if !isSameOrigin(r) {
				response.NewErrorResponse(response.ParamsError, "csrf check failed", "").FPrint(w)
				return
			}
		}

		ctx := &context.Context{}
		ctx.Context = r.Context()
		ctx.SetValue(context.LogID, id.GenLogID())
		lang := r.Header.Get("Lang")
		if lang == "" {
			lang = "en"
		}
		ctx.Lang = lang

		if config.IsInit {
			user := cast.ToString(session.Instance.Get(ctx, "user"))
			var userInfo models.User
			if user != "" {
				_ = json.Unmarshal([]byte(user), &userInfo)
			}
			if userInfo.ID > 0 {
				ctx.UserID = userInfo.ID
				ctx.UserName = userInfo.Name
				ctx.UserAccount = userInfo.Account
				ctx.IsAdmin = userInfo.IsAdmin == 1
			}

			if userInfo.ID == 0 {
				token := r.Header.Get("Token")
				if token != "" {
					user, err := getLoginInfoByToken(token, r.RemoteAddr)
					if user.ID >= 0 && err == nil {
						ctx.UserID = user.ID
						ctx.UserName = user.Name
						ctx.UserAccount = user.Account
						ctx.IsAdmin = user.IsAdmin == 1
					} else {
						response.NewErrorResponse(response.NeedLogin, err.Error(), "").FPrint(w)
						return
					}
				}
			}

			if ctx.UserID == 0 {
				if r.URL.Path != "/api/ping" && r.URL.Path != "/api/login" {
					response.NewErrorResponse(response.NeedLogin, i18n.GetText(ctx.Lang, "login_exp"), "").FPrint(w)
					return
				}
			}
		} else if r.URL.Path != "/api/setup" {
			response.NewErrorResponse(response.NeedSetup, "", "").FPrint(w)
			return
		}
		h(ctx, w, r)
	}
}

// getLoginInfoByToken 验证随机不透明 API Token，替代旧的密码派生 Token 格式
// 新格式：64字符随机hex字符串，服务端存储，支持过期检查、IP绑定、撤销
func getLoginInfoByToken(tokenString string, clientIP string) (models.User, error) {
	ret := models.User{}
	if len(tokenString) != 64 {
		return ret, errors.New("invalid token format")
	}

	// 根据 token 字符串查询数据库
	var apiToken models.ApiToken
	_, err := db.Instance.Table(&models.ApiToken{}).Where("token = ?", tokenString).Get(&apiToken)
	if err != nil {
		return ret, errors.New("token lookup failed")
	}
	if apiToken.ID == 0 {
		return ret, errors.New("token not found")
	}

	// 检查过期时间
	if !apiToken.ExpiresAt.IsZero() && time.Now().After(apiToken.ExpiresAt) {
		// 过期 Token 自动清理
		db.Instance.Table(&models.ApiToken{}).Where("id = ?", apiToken.ID).Delete(&models.ApiToken{})
		return ret, errors.New("token expired")
	}

	// 检查 IP 绑定（仅当生成时记录了 IP 才校验）
	if apiToken.ClientIP != "" && apiToken.ClientIP != clientIP {
		return ret, errors.New("token ip mismatch")
	}

	// 查询关联用户
	var user models.User
	_, err = db.Instance.Table("user").Where("id = ? and disabled=0", apiToken.UserID).Get(&user)
	if err != nil || user.ID == 0 {
		return ret, errors.New("user not found or disabled")
	}

	// 更新最后使用时间（忽略错误，不影响主流程）
	db.Instance.Table(&models.ApiToken{}).
		Where("id = ?", apiToken.ID).
		Update(map[string]interface{}{"last_used_at": time.Now()})

	return user, nil
}

// isSameOrigin 校验请求的 Origin 或 Referer Header 是否与当前服务域名匹配
// 用于 CSRF 防护的第二层防御（第一层为 SameSite Cookie）
// 优先检查 Origin Header（POST 请求浏览器必定携带），其次检查 Referer
// 修改日期: 20260425
func isSameOrigin(r *http.Request) bool {
	// 收集所有允许的域名（主域名 + 多域名配置）
	allowedDomains := config.Instance.Domains
	if len(allowedDomains) == 0 {
		allowedDomains = []string{config.Instance.Domain}
	}

	// 优先检查 Origin Header（标准 CSRF 防护方式）
	origin := r.Header.Get("Origin")
	if origin != "" {
		originURL, err := url.Parse(origin)
		if err != nil {
			return false
		}
		return isHostAllowed(originURL.Host, allowedDomains)
	}

	// Origin 为空时检查 Referer（部分隐私模式可能不发送 Origin）
	referer := r.Header.Get("Referer")
	if referer != "" {
		refererURL, err := url.Parse(referer)
		if err != nil {
			return false
		}
		return isHostAllowed(refererURL.Host, allowedDomains)
	}

	// Origin 和 Referer 均为空：允许通过
	// 场景：某些旧浏览器或隐私配置，SameSite Cookie 已提供基础防护
	return true
}

// isHostAllowed 检查请求主机是否在允许的域名列表中
// 支持带端口号的主机名（如 "example.com:8443"）
func isHostAllowed(host string, allowedDomains []string) bool {
	// 去除端口后比较（配置中的域名通常不含端口）
	hostWithoutPort := host
	if idx := strings.LastIndex(host, ":"); idx != -1 {
		hostWithoutPort = host[:idx]
	}

	for _, domain := range allowedDomains {
		if hostWithoutPort == domain || host == domain {
			return true
		}
	}
	return false
}
