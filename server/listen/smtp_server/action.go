package smtp_server

import (
	"database/sql"
	"errors"
	"net"
	"strings"

	"github.com/ffyuhf/pmail/db"
	"github.com/ffyuhf/pmail/models"
	"github.com/ffyuhf/pmail/utils/context"
	"github.com/ffyuhf/pmail/utils/id"
	"github.com/ffyuhf/pmail/utils/password"
	"github.com/ffyuhf/pmail/utils/ratelimit"
	"github.com/emersion/go-sasl"
	"github.com/emersion/go-smtp"
	log "github.com/sirupsen/logrus"
)

// Backend 实现 SMTP 服务器方法。
type Backend struct{}

func (bkd *Backend) NewSession(conn *smtp.Conn) (smtp.Session, error) {

	remoteAddress := conn.Conn().RemoteAddr()
	ctx := &context.Context{}
	ctx.SetValue(context.LogID, id.GenLogID())
	log.WithContext(ctx).Debugf("新SMTP连接")

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
	log.WithContext(s.Ctx).Debugf("Auth :%s", mech)
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

func (s *Session) AuthPlain(username, pwd string) error {
	log.WithContext(s.Ctx).Debugf("Auth %s %s", username, pwd)

	s.User = username

	infos := strings.Split(username, "@")
	if len(infos) > 1 {
		username = infos[0]
	}

	// 暴力破解防护：提取客户端 IP，检查速率限制
	clientIP := ratelimit.ExtractIP(s.RemoteAddress.String())
	if lockErr := ratelimit.Check(clientIP, username); lockErr != nil {
		log.WithContext(s.Ctx).WithField("ip", clientIP).Warnf("SMTP auth rate limited: %v", lockErr)
		return errors.New("too many failed attempts, try again later")
	}

	// 指数退避延迟：根据历史失败次数增加等待时间
	ratelimit.WaitDelay(clientIP, username)

	var user models.User

	// 仅用账号查询，不将密码作为查询条件（支持双算法验证）
	_, err := db.Instance.Where("account =? and disabled=0", username).Get(&user)
	if err != nil && err != sql.ErrNoRows {
		log.Errorf("%+v", err)
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

			log.WithContext(s.Ctx).Debugf("Auth Success %+v", user)
			return nil
		}
	}

	// 认证失败，记录失败
	ratelimit.RecordFailure(clientIP, username)
	log.WithContext(s.Ctx).Debugf("登陆错误%s %s", username, pwd)
	return errors.New("password error")
}

func (s *Session) Mail(from string, opts *smtp.MailOptions) error {
	log.WithContext(s.Ctx).Debugf("Mail Success %+v %+v", from, opts)
	s.From = from
	return nil
}

func (s *Session) Rcpt(to string, opts *smtp.RcptOptions) error {
	log.WithContext(s.Ctx).Debugf("Rcpt Success %+v", to)

	s.To = append(s.To, to)
	return nil
}

func (s *Session) Reset() {}

func (s *Session) Logout() error {
	return nil
}
