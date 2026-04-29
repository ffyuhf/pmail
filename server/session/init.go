package session

import (
	"net/http"

	"github.com/alexedwards/scs/mysqlstore"
	"github.com/alexedwards/scs/postgresstore"
	"github.com/alexedwards/scs/sqlite3store"
	"github.com/alexedwards/scs/v2"
	"github.com/ffyuhf/pmail/config"
	"github.com/ffyuhf/pmail/db"

	"time"
)

var Instance *scs.SessionManager

func Init() {
	Instance = scs.New()
	Instance.Lifetime = 7 * 24 * time.Hour

	// CSRF 防护：设置 SameSite=Lax，阻止跨站 POST 请求携带 Cookie
	// Lax 模式允许从外部链接通过 GET 导航进入站点，但阻止跨站 POST/DELETE 等写操作
	// 修改日期: 20260425
	Instance.Cookie.SameSite = http.SameSiteLaxMode

	// 会话安全：设置 HttpOnly，防止 JavaScript 通过 document.cookie 访问 session cookie
	// 有效防御 XSS 攻击导致的会话劫持
	// 修改日期: 20260430
	Instance.Cookie.HttpOnly = true

	// HTTPS 模式下设置 Secure 属性，确保 Cookie 仅通过加密通道传输
	// HttpsEnabled == 2 表示 HTTPS 已禁用，此时不设置 Secure（否则 Cookie 无法通过 HTTP 发送）
	if config.Instance.HttpsEnabled != 2 {
		Instance.Cookie.Secure = true
	}

	// 使用db存储session数据，目前为了架构简单，
	// 暂不引入redis存储，如果日后性能存在瓶颈，可以将session迁移到redis

	switch config.Instance.DbType {
	case config.DBTypeMySQL:
		Instance.Store = mysqlstore.New(db.Instance.DB().DB)
	case config.DBTypeSQLite:
		Instance.Store = sqlite3store.New(db.Instance.DB().DB)
	case config.DBTypePostgres:
		Instance.Store = postgresstore.New(db.Instance.DB().DB)
	default:
		panic("Unsupported database type: " + config.Instance.DbType)
	}

}
