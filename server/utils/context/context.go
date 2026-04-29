package context

import (
	"context"
)

// LogID 是日志上下文中用于存储会话唯一标识的 key。
const LogID = "LogID"

// Context 封装了请求级别的上下文信息，贯穿整个会话生命周期。
// 用于日志关联、用户身份传递和协议元数据记录。
type Context struct {
	context.Context `json:"-"`
	// UserID 是已认证用户的数据库 ID，未登录时为 0。
	UserID int
	// UserAccount 是已认证用户的账号名（如 "admin"）。
	UserAccount string
	// UserName 是已认证用户的显示名称。
	UserName string
	// Values 存储自定义键值对，如 LogID 等。
	Values map[string]any
	// Lang 是用户的语言偏好设置。
	Lang string
	// IsAdmin 标识当前用户是否为管理员。
	IsAdmin bool
	// Protocol 标识当前会话所属协议（SMTP/IMAP/POP3/HTTP）。
	Protocol string
	// ClientIP 是客户端的 IP 地址，用于日志和安全审计。
	ClientIP string
	// ServerPort 是服务器监听的端口号，用于区分服务类型。
	ServerPort int
	// Transport 标识传输层安全状态（TCP/TLS/STARTTLS）。
	Transport string
}

// SetValue 向上下文中存储自定义键值对。
func (c *Context) SetValue(key string, value any) {
	if c.Values == nil {
		c.Values = map[string]any{}
	}
	c.Values[key] = value
}

// GetValue 从上下文中读取自定义键值对，不存在则返回 nil。
func (c *Context) GetValue(key string) any {
	if c.Values == nil {
		return nil
	}
	return c.Values[key]
}
