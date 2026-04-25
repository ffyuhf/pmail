package imap_server

import (
	"errors"
	"strings"

	"github.com/ffyuhf/pmail/db"
	"github.com/ffyuhf/pmail/models"
	pmailpassword "github.com/ffyuhf/pmail/utils/password"
	"github.com/ffyuhf/pmail/utils/ratelimit"
	"github.com/emersion/go-imap/v2"
	log "github.com/sirupsen/logrus"
)

// Login 处理 IMAP LOGIN 命令，包含暴力破解防护。
// 根据用户数据库验证凭据并执行速率限制。
func (s *serverSession) Login(username, pwd string) error {

	infos := strings.Split(username, "@")
	if len(infos) > 1 {
		username = infos[0]
	}

	// 暴力破解防护：检查速率限制
	if lockErr := ratelimit.Check(s.remoteAddr, username); lockErr != nil {
		log.WithField("ip", s.remoteAddr).Warnf("IMAP login rate limited: %v", lockErr)
		return &imap.Error{
			Type: imap.StatusResponseTypeNo,
			Text: "too many failed attempts, try again later",
		}
	}

	// 指数退避延迟：根据历史失败次数增加等待时间
	ratelimit.WaitDelay(s.remoteAddr, username)

	var user models.User

	// 仅用账号查询
	_, err := db.Instance.Where("account =? and disabled=0", username).Get(&user)
	if err != nil {
		ratelimit.RecordFailure(s.remoteAddr, username)
		log.Errorf("%+v", err)
		return &imap.Error{
			Type: imap.StatusResponseTypeNo,
			Text: "login failed",
		}
	}

	if user.ID > 0 {
		// 使用双算法验证：先bcrypt，后旧MD5
		ok, needsUpgrade := pmailpassword.Verify(pwd, user.Password)
		if ok {
			// 登录成功，清除速率限制记录
			ratelimit.RecordSuccess(s.remoteAddr, username)

			// 旧MD5密码自动升级为bcrypt
			if needsUpgrade {
				newHash := pmailpassword.Encode(pwd)
				_, _ = db.Instance.Table("user").Where("id=?", user.ID).Update(map[string]interface{}{"password": newHash})
			}

			s.ctx.UserID = user.ID
			s.ctx.UserAccount = user.Account
			s.ctx.UserName = user.Name
			s.ctx.IsAdmin = user.IsAdmin == 1

			s.status = AUTHORIZED
			return nil
		}
	}

	// 认证失败，记录失败
	ratelimit.RecordFailure(s.remoteAddr, username)
	return errors.New("login failed")
}

func (s *serverSession) Authenticate(authType string) (interface{}, error) {
	return nil, errors.New("not implemented")
}
