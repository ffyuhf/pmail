package smtp_server

import (
	"database/sql"
	"errors"
	"net"
	"strings"

	"github.com/emersion/go-sasl"
	"github.com/emersion/go-smtp"
	"github.com/ffyuhf/pmail/db"
	"github.com/ffyuhf/pmail/models"
	"github.com/ffyuhf/pmail/utils/context"
	"github.com/ffyuhf/pmail/utils/id"
	"github.com/ffyuhf/pmail/utils/log"
	"github.com/ffyuhf/pmail/utils/password"
	"github.com/ffyuhf/pmail/utils/ratelimit"
)

// Backend 实现 SMTP 服务器方法。
type Backend struct{}

func (bkd *Backend) NewSession(conn *smtp.Conn) (smtp.Session, error) {
	remoteAddress := conn.Conn().RemoteAddr()
	ctx := &context.Context{}
	ctx.SetValue(context.LogID, id.GenLogID())
	// 设置协议标识，用于日志格式化器输出 [SMTP] 前缀
	ctx.Protocol = log.ProtocolSMTP
	ctx.ClientIP = ratelimit.ExtractIP(remoteAddress.String())

	log.SmtpInfof(ctx, log.EventSMTPSessionNew, "客户端IP=%s 客户端端口=%s", ctx.ClientIP, remoteAddress.String())

	return &Session{
		RemoteAddress: remoteAddress,
		Ctx:           ctx,
	}, nil
}

// Session 在 EHLO 后返回。
type Session struct {
	RemoteAddress net.Addr
	User          string
	From          string
	To            []string
	Ctx           *context.Context
}

// AuthMechanisms 返回支持的认证机制列表。
func (s *Session) AuthMechanisms() []string {
	return []string{sasl.Plain, sasl.Login}
}

// Auth 是支持的认证器的处理函数。
func (s *Session) Auth(mech string) (sasl.Server, error) {
	log.SmtpDebugf(s.Ctx, log.EventSMTPAuth, "认证机制=%s", mech)
	if mech == sasl.Plain {
		return sasl.NewPlainServer(func(identity, username, password string) error {
			return s.AuthPlain(username, password)
		}), nil
	}

	if mech == sasl.Login {
		return NewLoginServer(func(username, password string) error {
			return s.AuthPlain(username, password)
		}), nil
	}

	return nil, errors.New("Auth Not Supported")
}

// AuthPlain 处理 SMTP PLAIN/LOGIN 认证，包含暴力破解防护。
func (s *Session) AuthPlain(username, pwd string) error {
	log.SmtpDebugf(s.Ctx, log.EventSMTPAuth, "用户名=%s 密码=%s", username, pwd)

	s.User = username

	infos := strings.Split(username, "@")
	if len(infos) > 1 {
		username = infos[0]
	}

	// 暴力破解防护：提取客户端 IP，检查速率限制
	clientIP := ratelimit.ExtractIP(s.RemoteAddress.String())
	if lockErr := ratelimit.Check(clientIP, username); lockErr != nil {
		log.SmtpWarnf(s.Ctx, log.EventSMTPRateLimit, "IP=%s 用户=%s 原因=%v", clientIP, username, lockErr)
		return errors.New("too many failed attempts, try again later")
	}

	// 指数退避延迟：根据历史失败次数增加等待时间
	ratelimit.WaitDelay(clientIP, username)

	var user models.User

	// 仅用账号查询，不将密码作为查询条件（支持双算法验证）
	_, err := db.Instance.Where("account =? and disabled=0", username).Get(&user)
	if err != nil && err != sql.ErrNoRows {
		log.SmtpErrorf(s.Ctx, log.EventSMTPAuth, "数据库查询失败 用户=%s 错误=%v", username, err)
	}

	if user.ID > 0 {
		// 使用双算法验证：先bcrypt，后旧MD5
		ok, needsUpgrade := password.Verify(pwd, user.Password)
		if ok {
			// 登录成功，清除速率限制记录
			ratelimit.RecordSuccess(clientIP, username)

			// 旧MD5密码自动升级为bcrypt
			if needsUpgrade {
				newHash := password.Encode(pwd)
				_, _ = db.Instance.Table("user").Where("id=?", user.ID).Update(map[string]interface{}{"password": newHash})
			}

			s.Ctx.UserAccount = user.Account
			s.Ctx.UserID = user.ID
			s.Ctx.UserName = user.Name
			s.Ctx.IsAdmin = user.IsAdmin == 1

			log.SmtpInfof(s.Ctx, log.EventSMTPAuthSuccess, "用户=%s IP=%s", user.Account, clientIP)
			return nil
		}
	}

	// 认证失败，记录失败
	ratelimit.RecordFailure(clientIP, username)
	log.SmtpWarnf(s.Ctx, log.EventSMTPAuthFail, "用户=%s IP=%s", username, clientIP)
	return errors.New("password error")
}

// Mail 处理 SMTP MAIL FROM 命令，记录发件人地址。
func (s *Session) Mail(from string, opts *smtp.MailOptions) error {
	log.SmtpInfof(s.Ctx, log.EventSMTPMailFrom, "发件人=%s", from)
	s.From = from
	return nil
}

// Rcpt 处理 SMTP RCPT TO 命令，记录收件人地址。
func (s *Session) Rcpt(to string, opts *smtp.RcptOptions) error {
	log.SmtpInfof(s.Ctx, log.EventSMTPRcptTo, "收件人=%s", to)
	s.To = append(s.To, to)
	return nil
}

func (s *Session) Reset() {}

func (s *Session) Logout() error {
	return nil
}
