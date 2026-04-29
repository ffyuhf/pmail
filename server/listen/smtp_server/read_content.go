package smtp_server

import (
	"bytes"
	"database/sql"
	"encoding/json"
	oerrors "errors"
	"io"
	"net"
	"net/netip"
	"strings"
	"time"

	"github.com/ffyuhf/pmail/config"
	"github.com/ffyuhf/pmail/db"
	"github.com/ffyuhf/pmail/dto/parsemail"
	"github.com/ffyuhf/pmail/hooks"
	"github.com/ffyuhf/pmail/hooks/framework"
	"github.com/ffyuhf/pmail/listen/imap_server"
	"github.com/ffyuhf/pmail/models"
	"github.com/ffyuhf/pmail/services/rule"
	"github.com/ffyuhf/pmail/utils/array"
	"github.com/ffyuhf/pmail/utils/async"
	pmailContext "github.com/ffyuhf/pmail/utils/context"
	pmailLog "github.com/ffyuhf/pmail/utils/log"
	"github.com/ffyuhf/pmail/utils/send"
	"github.com/mileusna/spf"
	"github.com/spf13/cast"
	. "xorm.io/builder"
)

// DropUnknownRecipientEmails 是代码级别的功能开关
// 当设置为 true 时，发给不存在用户的邮件将被直接丢弃，不会保存到数据库
// 这可以有效防止扫描器产生的垃圾邮件被存入第一个用户（管理员）的邮箱
// 注意：此开关只能通过修改代码来更改，不暴露给用户配置
const DropUnknownRecipientEmails = true

// Data 处理 SMTP DATA 命令，接收并处理完整邮件内容。
// 根据是否已认证区分转发（已登录用户发送）和收件（外部邮件接收）两种流程。
func (s *Session) Data(r io.Reader) error {

	ctx := s.Ctx

	emailData, err := io.ReadAll(r)
	if err != nil {
		pmailLog.SmtpErrorf(ctx, pmailLog.EventSMTPMailRecv, "读取失败 错误=%v", err)
		return err
	}

	pmailLog.SmtpDebugf(ctx, pmailLog.EventSMTPMailRecv, "原始数据大小=%d", len(emailData))

	// 执行 ReceiveParseBefore 插件链
	pmailLog.SmtpDebug(ctx, pmailLog.EventSMTPPlugin, "插件=接收解析前 开始")
	for _, hook := range hooks.HookList {
		if hook == nil {
			continue
		}
		hook.ReceiveParseBefore(ctx, &emailData)
	}
	pmailLog.SmtpDebug(ctx, pmailLog.EventSMTPPlugin, "插件=接收解析前 结束")

	email := parsemail.NewEmailFromReader(s.To, bytes.NewReader(emailData), len(emailData))

	if s.From != "" {
		from := parsemail.BuilderUser(s.From)
		if email.From == nil {
			email.From = from
		}
		if email.From.EmailAddress != from.EmailAddress {
			// 协议中的from和邮件内容中的from不匹配，当成垃圾邮件处理
		}
	}

	// 判断是收信还是转发，只要是登陆了，都当成转发处理
	if s.Ctx.UserID > 0 {
		account, _ := email.From.GetDomainAccount()
		if account != ctx.UserAccount && !ctx.IsAdmin {
			return oerrors.New("No Auth")
		}

		pmailLog.SmtpDebug(ctx, pmailLog.EventSMTPPlugin, "插件=发送前 开始")
		for _, hook := range hooks.HookList {
			if hook == nil {
				continue
			}
			hook.SendBefore(ctx, email)
		}
		pmailLog.SmtpDebug(ctx, pmailLog.EventSMTPPlugin, "插件=发送前 结束")

		if email == nil {
			return nil
		}

		// 转发流程：保存邮件记录
		_, _, err := saveEmail(ctx, len(emailData), email, s.Ctx.UserID, 1, nil, true, true)
		if err != nil {
			pmailLog.SmtpErrorf(ctx, pmailLog.EventSMTPMailSend, "保存失败 错误=%v", err)
		}

		errMsg := ""
		err, sendErr := send.Send(ctx, email)

		pmailLog.SmtpDebug(ctx, pmailLog.EventSMTPPlugin, "插件=发送后 开始")
		as3 := async.New(ctx)
		for _, hook := range hooks.HookList {
			if hook == nil {
				continue
			}
			as3.WaitProcess(func(hk any) {
				hk.(framework.EmailHook).SendAfter(ctx, email, sendErr)
			}, hook)
		}
		as3.Wait()
		pmailLog.SmtpDebug(ctx, pmailLog.EventSMTPPlugin, "插件=发送后 结束")

		if err != nil {
			errMsg = err.Error()
			pmailLog.SmtpErrorf(ctx, pmailLog.EventSMTPMailDeliver, "消息ID=%d 状态=失败 错误=%s", email.MessageId, errMsg)
			_, err := db.Instance.Exec(db.WithContext(ctx, "update email set status =2 ,error=? where id = ? "), errMsg, email.MessageId)
			if err != nil {
				pmailLog.SmtpErrorf(ctx, pmailLog.EventSMTPMailDeliver, "数据库更新失败 错误=%v", err)
			}
			_, err = db.Instance.Exec(db.WithContext(ctx, "update user_email set status =2  where email_id = ? "), email.MessageId)
			if err != nil {
				pmailLog.SmtpErrorf(ctx, pmailLog.EventSMTPMailDeliver, "数据库更新失败 错误=%v", err)
			}

		} else {
			pmailLog.SmtpInfof(ctx, pmailLog.EventSMTPMailDeliver, "消息ID=%d 状态=成功", email.MessageId)
			_, err := db.Instance.Exec(db.WithContext(ctx, "update email set status =1  where id = ? "), email.MessageId)
			if err != nil {
				pmailLog.SmtpErrorf(ctx, pmailLog.EventSMTPMailDeliver, "数据库更新失败 错误=%v", err)
			}
			_, err = db.Instance.Exec(db.WithContext(ctx, "update user_email set status =1  where email_id = ? "), email.MessageId)
			if err != nil {
				pmailLog.SmtpErrorf(ctx, pmailLog.EventSMTPMailDeliver, "数据库更新失败 错误=%v", err)
			}
		}

	} else {
		// 收件流程：接收外部邮件

		var dkimStatus, SPFStatus bool

		// DKIM 校验
		dkimStatus = parsemail.Check(ctx, bytes.NewReader(emailData))
		pmailLog.SmtpInfof(ctx, pmailLog.EventSMTPDKIM, "结果=%v 发件人=%s", dkimStatus, email.From.EmailAddress)

		// SPF 校验
		SPFStatus = spfCheck(s.RemoteAddress.String(), email.Sender, email.Sender.EmailAddress)
		pmailLog.SmtpInfof(ctx, pmailLog.EventSMTPSPF, "结果=%v IP=%s 发送者=%s", SPFStatus, s.RemoteAddress.String(), email.Sender.EmailAddress)

		// 执行 ReceiveParseAfter 插件链
		pmailLog.SmtpDebug(ctx, pmailLog.EventSMTPPlugin, "插件=接收解析后 开始")
		for _, hook := range hooks.HookList {
			if hook == nil {
				continue
			}
			hook.ReceiveParseAfter(ctx, email)
		}
		pmailLog.SmtpDebug(ctx, pmailLog.EventSMTPPlugin, "插件=接收解析后 结束")

		_, formDomain := email.From.GetDomainAccount()
		// 伪造邮件检测：本地域名但 SPF 不通过
		if array.InArray(formDomain, config.Instance.Domains) && SPFStatus == false {
			dkimStatus = false
			email.Status = 3
			pmailLog.SmtpWarnf(ctx, pmailLog.EventSMTPMailReject, "原因=伪造发件人 发件人=%s 域名=%s SPF=false", email.From.EmailAddress, formDomain)
		}

		users, dbEmail, _ := saveEmail(ctx, len(emailData), email, 0, 0, s.To, SPFStatus, dkimStatus)

		if email.MessageId > 0 {
			pmailLog.SmtpDebug(ctx, pmailLog.EventSMTPPlugin, "插件=规则匹配 开始")
			for _, user := range users {
				// 执行邮件规则
				rs := rule.GetAllRules(ctx, user.ID)
				for _, r := range rs {
					if rule.MatchRule(ctx, r, email) {
						rule.DoRule(ctx, r, email, user)
					}
				}
			}
			pmailLog.SmtpDebug(ctx, pmailLog.EventSMTPPlugin, "插件=规则匹配 结束")
		}

		// 执行 ReceiveSaveAfter 插件链
		pmailLog.SmtpDebug(ctx, pmailLog.EventSMTPPlugin, "插件=接收保存后 开始")
		var ue []*models.UserEmail
		err = db.Instance.Table(&models.UserEmail{}).Where("email_id=?", email.MessageId).Find(&ue)
		if err != nil {
			pmailLog.SmtpErrorf(ctx, pmailLog.EventSMTPMailRecv, "查询用户邮件关联失败 错误=%v", err)
		}
		as3 := async.New(ctx)
		for _, hook := range hooks.HookList {
			if hook == nil {
				continue
			}
			as3.WaitProcess(func(hk any) {
				hk.(framework.EmailHook).ReceiveSaveAfter(ctx, email, ue)
			}, hook)
		}
		as3.Wait()
		pmailLog.SmtpDebug(ctx, pmailLog.EventSMTPPlugin, "插件=接收保存后 结束")

		// IDLE 命令通知已连接的 IMAP 客户端
		for _, user := range users {
			imap_server.IdleNotice(ctx, user.ID, dbEmail)
		}

	}

	return nil
}

// saveEmail 将邮件保存到数据库，并根据收件人匹配用户。
// 对于收件流程（emailType=0），查找匹配的本地用户；找不到时根据垃圾邮件过滤策略决定丢弃或转交管理员。
func saveEmail(ctx *pmailContext.Context, size int, email *parsemail.Email, sendUserID int, emailType int, reallyTo []string, SPFStatus, dkimStatus bool) ([]*models.User, *models.Email, error) {
	var dkimV, spfV int8
	if dkimStatus {
		dkimV = 1
	}
	if SPFStatus {
		spfV = 1
	}

	if email == nil {
		return nil, nil, nil
	}

	msgID := email.MsgID
	if msgID == "" {
		msgID = parsemail.GenerateMsgID(config.Instance.Domain)
	}

	modelEmail := models.Email{
		Type:         cast.ToInt8(emailType),
		Subject:      email.Subject,
		ReplyTo:      json2string(email.ReplyTo),
		FromName:     email.From.Name,
		FromAddress:  email.From.EmailAddress,
		To:           json2string(email.To),
		Bcc:          json2string(email.Bcc),
		Cc:           json2string(email.Cc),
		Text:         sql.NullString{String: string(email.Text), Valid: true},
		Html:         sql.NullString{String: string(email.HTML), Valid: true},
		Sender:       json2string(email.Sender),
		Attachments:  json2string(email.Attachments),
		Size:         email.Size,
		SPFCheck:     spfV,
		DKIMCheck:    dkimV,
		SendUserID:   sendUserID,
		SendDate:     time.Now(),
		Status:       cast.ToInt8(email.Status),
		CreateTime:   time.Now(),
		CronSendTime: time.Now(),
		MsgID:        msgID,
	}

	_, err := db.Instance.Insert(&modelEmail)

	if err != nil {
		pmailLog.SmtpErrorf(ctx, pmailLog.EventSMTPMailRecv, "数据库插入邮件失败 错误=%v", err)
	}

	if modelEmail.Id > 0 {
		email.MessageId = cast.ToInt64(modelEmail.Id)
		email.MsgID = modelEmail.MsgID
	}

	// 收信人信息
	var users []*models.User

	// 如果是收信
	if emailType == 0 {
		// 找到收信人id
		accounts := []string{}
		// 优先取smtp协议中的收件人地址
		if len(reallyTo) > 0 {
			for _, s := range reallyTo {
				account := parsemail.BuilderUser(s)
				if account != nil {
					acc, domain := account.GetDomainAccount()
					if array.InArray(domain, config.Instance.Domains) && acc != "" {
						accounts = append(accounts, strings.ToLower(acc))
					}
				}
			}
		} else {
			for _, user := range append(append(email.To, email.Cc...), email.Bcc...) {
				account, _ := user.GetDomainAccount()
				if account != "" {
					accounts = append(accounts, strings.ToLower(account))
				}
			}
		}

		/** 这里会导致索引失效，可以尝试对lower结果加索引
		PostgreSQL: CREATE INDEX idx_user_account_lower ON "user" (LOWER(account));
		MySQL8+: ALTER TABLE user ADD INDEX ((LOWER(account)));
		*/
		where, params, _ := ToSQL(In("LOWER(account)", accounts))

		err = db.Instance.Table(&models.User{}).Where(where, params...).Find(&users)
		if err != nil {
			pmailLog.SmtpErrorf(ctx, pmailLog.EventSMTPMailRecv, "查询收件人失败 错误=%v", err)
		}

		if len(users) > 0 {
			pmailLog.SmtpInfof(ctx, pmailLog.EventSMTPMailRecv, "发件人=%s 收件人=%v 大小=%d 匹配用户数=%d", email.From.EmailAddress, accounts, size, len(users))
			for _, user := range users {
				ue := models.UserEmail{EmailID: modelEmail.Id, UserID: user.ID, Status: cast.ToInt8(email.Status)}
				_, err = db.Instance.Insert(&ue)
				if err != nil {
					pmailLog.SmtpErrorf(ctx, pmailLog.EventSMTPMailRecv, "插入用户邮件关联失败 错误=%v", err)
				}
			}
		} else {
			// 找不到收件人
			// 如果启用了保护功能，且验证失败，则丢弃邮件
			// 验证通过的邮件即使收件人不存在也应交给管理员，避免误丢弃合法邮件

			// 垃圾过滤
			if DropUnknownRecipientEmails &&
				((config.Instance.SpamFilterLevel == 1 && !SPFStatus && !dkimStatus) ||
					(config.Instance.SpamFilterLevel == 2 && !SPFStatus) ||
					(config.Instance.SpamFilterLevel == 3 && !dkimStatus)) {
				// 垃圾邮件：收件人不存在且验证失败，直接丢弃
				pmailLog.SmtpWarnf(ctx, pmailLog.EventSMTPMailReject, "发件人=%s 收件人=%v 原因=收件人不存在且验证失败 SPF=%v DKIM=%v 过滤级别=%d", email.From.EmailAddress, accounts, SPFStatus, dkimStatus, config.Instance.SpamFilterLevel)
				_, delErr := db.Instance.Delete(&models.Email{Id: modelEmail.Id})
				if delErr != nil {
					pmailLog.SmtpErrorf(ctx, pmailLog.EventSMTPMailReject, "数据库删除失败 错误=%v", delErr)
				}
				return nil, nil, nil
			}

			// DKIM 验证通过或未启用保护功能时，邮件丢给管理员账号
			pmailLog.SmtpInfof(ctx, pmailLog.EventSMTPMailForwardAdmin, "发件人=%s 收件人=%v 原因=收件人不存在 SPF=%v DKIM=%v", email.From.EmailAddress, accounts, SPFStatus, dkimStatus)

			err = db.Instance.Table(&models.User{}).Where("is_admin=1").Find(&users)
			for _, user := range users {
				ue := models.UserEmail{EmailID: modelEmail.Id, UserID: user.ID, Status: cast.ToInt8(email.Status)}
				_, err = db.Instance.Insert(&ue)
				if err != nil {
					pmailLog.SmtpErrorf(ctx, pmailLog.EventSMTPMailForwardAdmin, "数据库插入失败 错误=%v", err)
				}
			}
		}
	} else {
		ue := models.UserEmail{EmailID: modelEmail.Id, UserID: ctx.UserID}
		_, err = db.Instance.Insert(&ue)
		if err != nil {
			pmailLog.SmtpErrorf(ctx, pmailLog.EventSMTPMailSend, "插入用户邮件关联失败 错误=%v", err)
		}
	}

	return users, &modelEmail, nil
}

func json2string(d any) string {
	by, _ := json.Marshal(d)
	return string(by)
}

// spfCheck 对发件人 IP 地址执行 SPF 记录校验。
// 内网 IP 默认通过；外网 IP 查询域名 SPF 记录判断是否授权。
func spfCheck(remoteAddress string, sender *parsemail.User, senderString string) bool {
	ipAddress, _ := netip.ParseAddrPort(remoteAddress)

	ip := net.ParseIP(ipAddress.Addr().String())
	if ip.IsPrivate() {
		return true
	}

	tmp := strings.Split(sender.EmailAddress, "@")
	if len(tmp) < 2 {
		return false
	}

	res := spf.CheckHost(ip, tmp[1], senderString, "")

	if res == spf.None || res == spf.Pass {
		return true
	}
	return false
}
