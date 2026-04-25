package models

import "time"

// ApiToken 服务端存储的随机不透明 API 令牌，替代旧的密码派生 Token 格式
// 解决旧格式存在的密码哈希暴露、无撤销机制、无客户端绑定等安全隐患
type ApiToken struct {
	ID         int       `xorm:"id unsigned int not null pk autoincr"`
	UserID     int       `xorm:"user_id unsigned int not null index"`
	Token      string    `xorm:"token varchar(64) notnull unique"`
	ClientIP   string    `xorm:"client_ip varchar(45) notnull default('')" comment("生成时绑定的客户端IP，为空表示不绑定")`
	ExpiresAt  time.Time `xorm:"expires_at datetime index"`
	CreatedAt  time.Time `xorm:"created_at datetime"`
	LastUsedAt time.Time `xorm:"last_used_at datetime"`
}

func (p *ApiToken) TableName() string {
	return "api_tokens"
}
