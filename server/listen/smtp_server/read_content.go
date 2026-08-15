package smtp_server

import (
	"bytes"
	"database/sql"
	"encoding/json"
	oerrors "errors"
	"fmt"
	"io"
	"net"
	"net/netip"
	"net/textproto"
	"strings"
	"time"

	"github.com/ffyuhf/pmail/config"
	"github.com/ffyuhf/pmail/db"
	"github.com/ffyuhf/pmail/dto/parsemail"
	"github.com/ffyuhf/pmail/hooks"
	"github.com/ffyuhf/pmail/hooks/framework"
	"github.com/ffyuhf/pmail/listen/imap_server"
	"github.com/ffyuhf/pmail/models"
	dsnService "github.com/ffyuhf/pmail/services/dsn"
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

	// 修改日期: 20260610 — #1 修复 From 地址伪装检查空实现
	// 已认证用户发送邮件时，envelope From 必须与 header From 一致（管理员豁免）
	// 管理员不受限制，可自由发送任意 From 地址
	if s.From != "" {
		from := parsemail.BuilderUser(s.From)
		if email.From == nil {
			email.From = from
		}
		if email.From.EmailAddress != from.EmailAddress {
			// 已认证用户（非管理员）：envelope From 与 header From 不匹配，拒绝发送
			if s.Ctx.UserID > 0 && !ctx.IsAdmin {
				pmailLog.SmtpErrorf(ctx, pmailLog.EventSMTPMailSend, "原因=From地址伪装 envelope=%s header=%s 用户ID=%d", from.EmailAddress, email.From.EmailAddress, ctx.UserID)
				return oerrors.New("envelope from mismatch with header from")
			}
			// 未认证用户（接收路径）或管理员：记录告警但允许继续
			pmailLog.SmtpWarnf(ctx, pmailLog.EventSMTPMailRecv, "原因=From不匹配 envelope=%s header=%s", from.EmailAddress, email.From.EmailAddress)
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

			// 阶段二 P1（RFC 3461）：投递失败时，对请求了 DSN 的收件人生成退信报告。
			// 仅在存在 DSN 请求（DSNRcptOpts 非空）时执行，不影响不使用 DSN 的正常流程。
			if s.DSNRcptOpts != nil || s.DSMailOpts != nil {
				s.sendDSNOnFailure(ctx, email, emailData, errMsg, sendErr)
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

		// SPF/DKIM 校验完成后生成统一认证结论，供插件链、规则与落库共享
		email.Authentication = parsemail.NewEmailAuthentication(SPFStatus, dkimStatus)

		// 执行 ReceiveParseAfter 插件链
		pmailLog.SmtpDebug(ctx, pmailLog.EventSMTPPlugin, "插件=接收解析后 开始")
		for _, hook := range hooks.HookList {
			if hook == nil {
				continue
			}
			// 主程序认证结果不能被插件修改后带入数据库和后续钩子，每个插件执行前重置。
			email.Authentication = parsemail.NewEmailAuthentication(SPFStatus, dkimStatus)
			hook.ReceiveParseAfter(ctx, email)
		}
		pmailLog.SmtpDebug(ctx, pmailLog.EventSMTPPlugin, "插件=接收解析后 结束")
		// 钩子链执行完毕后再次重置，保证落库与后续流程使用权威认证结果。
		email.Authentication = parsemail.NewEmailAuthentication(SPFStatus, dkimStatus)

		// 修改日期: 20260610 — #12 修复 Sender 头未参与伪造检测
		// 同时检查 From 和 Sender 头的域名，防止通过 Sender 头伪造发件人
		_, formDomain := email.From.GetDomainAccount()
		isForgedFrom := array.InArray(formDomain, config.Instance.Domains) && !SPFStatus

		// 检查 Sender 头（邮件客户端可能优先显示 Sender 字段）
		var isForgedSender bool
		if email.Sender != nil {
			_, senderDomain := email.Sender.GetDomainAccount()
			if senderDomain != "" {
				isForgedSender = array.InArray(senderDomain, config.Instance.Domains) && !SPFStatus
			}
		}

		// 伪造邮件检测：From 或 Sender 的域名为本地域名但 SPF 不通过
		if isForgedFrom || isForgedSender {
			dkimStatus = false
			email.Status = 3
			forgedDomain := formDomain
			if isForgedSender && email.Sender != nil {
				_, forgedDomain = email.Sender.GetDomainAccount()
			}
			pmailLog.SmtpWarnf(ctx, pmailLog.EventSMTPMailReject, "原因=伪造发件人 发件人=%s 域名=%s SPF=false", email.From.EmailAddress, forgedDomain)
		}

		users, dbEmail, _ := saveEmail(ctx, len(emailData), email, 0, 0, s.To, SPFStatus, dkimStatus)

		if email.MessageId > 0 {
			pmailLog.SmtpDebug(ctx, pmailLog.EventSMTPPlugin, "插件=规则匹配 开始")
			for _, user := range users {
				// 执行邮件规则
				rs := rule.GetAllRules(ctx, user.ID)
				for _, r := range rs {
					if rule.MatchRule(ctx, r, email) {
						rule.DoRule(ctx, r, email, user, emailData)
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
	// 认证隔离：收信路径以 Authentication 为准；发件路径（emailType!=0）参数不被收信认证对象污染。
	if emailType == 0 && email != nil && email.Authentication != nil {
		SPFStatus = email.Authentication.SPFPassed
		dkimStatus = email.Authentication.DKIMPassed
	}
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

// sendDSNOnFailure 在邮件投递失败时，对请求了 DSN FAILURE 通知的收件人生成并发送退信报告。
// 该方法是阶段二 P1（RFC 3461）的核心逻辑，仅在客户端请求了 DSN 时被调用。
//
// 参数：
//   - ctx:        请求上下文（含日志 ID）
//   - email:      解析后的邮件对象（含发件人、收件人等信息）
//   - emailData:  原始邮件字节数据（用于构造 RET=FULL 的 DSN 报告）
//   - errMsg:     顶层错误消息
//   - sendErr:    按域名分组的发送错误（key 为域名，value 为错误）
func (s *Session) sendDSNOnFailure(
	ctx *pmailContext.Context,
	email *parsemail.Email,
	emailData []byte,
	errMsg string,
	sendErr map[string]error,
) {
	// 收集所有请求了 DSN FAILURE 通知的失败收件人
	var failedRcpts []dsnService.FailedRecipient

	for _, to := range s.To {
		rcptOpts := s.DSNRcptOpts[to]

		// 检查该收件人是否请求了失败通知
		if !dsnService.WantsFailureDSN(rcptOpts) {
			continue
		}

		// 尝试从 sendErr 中匹配该收件人的域名级错误
		addr := dsnService.ParseEmailAddress(to)
		parts := strings.SplitN(addr, "@", 2)
		domain := ""
		if len(parts) == 2 {
			domain = parts[1]
		}

		// 仅当该收件人的域名存在发送错误时才生成 DSN
		domainErr := ""
		if domainErrVal, ok := sendErr[domain]; ok {
			domainErr = domainErrVal.Error()
		} else {
			// 该收件人域名无错误，跳过
			continue
		}

		failedRcpts = append(failedRcpts, dsnService.FailedRecipient{
			Address:        addr,
			DSNStatus:      "5.4.4", // 无法路由 / 投递失败
			DiagnosticCode: domainErr,
			RcptOpts:       rcptOpts,
		})
	}

	if len(failedRcpts) == 0 {
		return
	}

	// 解析原始邮件的 MIME 头（用于 Auto-Submitted 检查和 DSN 第三部分）
	reader := bytes.NewReader(emailData)
	var originalHeaders textproto.MIMEHeader
	// 简单解析：逐行读取头部直到空行
	headerBuf := make(textproto.MIMEHeader)
	lineBuf := &bytes.Buffer{}
	for {
		b, err := reader.ReadByte()
		if err != nil {
			break
		}
		if b == '\n' {
			line := strings.TrimRight(lineBuf.String(), "\r")
			if line == "" {
				break // 空行 = 头部结束
			}
			// 解析 Header: Value
			colonIdx := strings.Index(line, ":")
			if colonIdx > 0 {
				key := strings.TrimSpace(line[:colonIdx])
				val := strings.TrimSpace(line[colonIdx+1:])
				headerBuf.Add(key, val)
			}
			lineBuf.Reset()
		} else {
			lineBuf.WriteByte(b)
		}
	}
	originalHeaders = headerBuf

	// 生成 DSN 报告
	dsnReport, err := dsnService.GenerateFailureDSN(
		email.From.EmailAddress,
		originalHeaders,
		failedRcpts,
		s.DSMailOpts,
		config.Instance.Domain,
		emailData,
	)
	if err != nil {
		pmailLog.SmtpErrorf(ctx, pmailLog.EventSMTPMailDeliver, "DSN报告生成失败 错误=%v", err)
		return
	}
	if dsnReport == nil {
		// 原始邮件包含 Auto-Submitted 头，跳过 DSN 生成（防止退信循环）
		pmailLog.SmtpInfof(ctx, pmailLog.EventSMTPMailDeliver, "DSN跳过=原始邮件已标记Auto-Submitted")
		return
	}

	// 构造 DSN 退信邮件并发送
	bounceEmail := &parsemail.Email{
		From: &parsemail.User{
			EmailAddress: fmt.Sprintf("MAILER-DAEMON@%s", config.Instance.Domain),
		},
		To: []*parsemail.User{
			{EmailAddress: email.From.EmailAddress},
		},
		Subject: "Delivery Status Notification (Failure)",
		Text:    dsnReport,
		HTML:    dsnReport,
	}

	if _, err := send.Send(ctx, bounceEmail); err != nil {
		pmailLog.SmtpErrorf(ctx, pmailLog.EventSMTPMailDeliver, "DSN退信发送失败 错误=%v", err)
	} else {
		pmailLog.SmtpInfof(ctx, pmailLog.EventSMTPMailDeliver, "DSN退信已发送 收件人=%s", email.From.EmailAddress)
	}
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
