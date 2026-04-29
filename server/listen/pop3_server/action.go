package pop3_server

import (
	"database/sql"
	errors2 "errors"
	"strings"

	"github.com/Jinnrry/gopop"
	"github.com/ffyuhf/pmail/consts"
	"github.com/ffyuhf/pmail/db"
	"github.com/ffyuhf/pmail/dto"
	"github.com/ffyuhf/pmail/dto/parsemail"
	"github.com/ffyuhf/pmail/models"
	"github.com/ffyuhf/pmail/services/del_email"
	"github.com/ffyuhf/pmail/services/detail"
	"github.com/ffyuhf/pmail/services/list"
	"github.com/ffyuhf/pmail/utils/array"
	"github.com/ffyuhf/pmail/utils/context"
	"github.com/ffyuhf/pmail/utils/errors"
	"github.com/ffyuhf/pmail/utils/id"
	pmailLog "github.com/ffyuhf/pmail/utils/log"
	"github.com/ffyuhf/pmail/utils/password"
	"github.com/ffyuhf/pmail/utils/ratelimit"
	"github.com/spf13/cast"
)

type action struct {
}

// ensureCtx 确保 POP3 会话上下文存在，并设置协议标识和客户端 IP。
// POP3 协议的日志前缀为 [POP3]，便于日志过滤。
func ensureCtx(session *gopop.Session) *context.Context {
	if session.Ctx == nil {
		tc := &context.Context{}
		tc.SetValue(context.LogID, id.GenLogID())
		tc.Protocol = pmailLog.ProtocolPOP3
		if session.Conn != nil {
			tc.ClientIP = ratelimit.ExtractIP(session.Conn.RemoteAddr().String())
		}
		session.Ctx = tc
	}
	return session.Ctx.(*context.Context)
}

// Custom 处理非标准 POP3 命令。
func (a action) Custom(session *gopop.Session, cmd string, args []string) ([]string, error) {
	ctx := ensureCtx(session)
	pmailLog.Pop3Debugf(ctx, pmailLog.EventPOP3Cmd, "命令=%s 参数=%v 原因=不支持", cmd, args)
	return nil, errors2.New("not supported cmd request")
}

// Capa 返回服务端支持的 POP3 命令列表。
func (a action) Capa(session *gopop.Session) ([]string, error) {
	ctx := ensureCtx(session)
	pmailLog.Pop3Debugf(ctx, pmailLog.EventPOP3Cmd, "命令=CAPA TLS=%v", session.InTls)

	ret := []string{
		"USER",
		"PASS",
		"TOP",
		"APOP",
		"STAT",
		"UIDL",
		"LIST",
		"RETR",
		"DELE",
		"REST",
		"NOOP",
		"QUIT",
	}
	if !session.InTls {
		ret = append(ret, "STLS")
	}

	return ret, nil
}

// User 处理 POP3 USER 命令，提交登录用户名。
func (a action) User(session *gopop.Session, username string) error {
	ctx := ensureCtx(session)

	infos := strings.Split(username, "@")
	if len(infos) > 1 {
		username = infos[0]
	}

	pmailLog.Pop3Debugf(ctx, pmailLog.EventPOP3Cmd, "命令=USER 用户=%s", username)
	session.User = username
	return nil
}

// pop3ClientIP extracts the client IP from a POP3 session's TCP connection.
func pop3ClientIP(session *gopop.Session) string {
	if session.Conn != nil {
		return ratelimit.ExtractIP(session.Conn.RemoteAddr().String())
	}
	return ""
}

// Pass 处理 POP3 PASS 命令，提交密码验证，包含暴力破解防护。
// 安全策略：拒绝非TLS连接上的认证，防止凭据明文传输（与SMTP AllowInsecureAuth=false对齐）。
func (a action) Pass(session *gopop.Session, pwd string) error {
	ctx := ensureCtx(session)

	// 强制TLS认证：非加密连接上拒绝提交密码，客户端必须先执行STLS升级为TLS
	if !session.InTls {
		pmailLog.Pop3Warnf(ctx, pmailLog.EventPOP3AuthRejected, "用户=%s 原因=明文连接", session.User)
		return errors2.New("must use STLS before authentication")
	}

	pmailLog.Pop3Debugf(ctx, pmailLog.EventPOP3Cmd, "命令=PASS 用户=%s 密码=%s", session.User, pwd)

	// 暴力破解防护：提取客户端 IP，检查速率限制
	clientIP := pop3ClientIP(session)
	if lockErr := ratelimit.Check(clientIP, session.User); lockErr != nil {
		pmailLog.Pop3Warnf(ctx, pmailLog.EventPOP3RateLimit, "IP=%s 用户=%s 原因=%v", clientIP, session.User, lockErr)
		return errors2.New("too many failed attempts, try again later")
	}

	// 指数退避延迟
	ratelimit.WaitDelay(clientIP, session.User)

	var user models.User

	// 仅用账号查询，不将密码作为查询条件（支持双算法验证）
	_, err := db.Instance.Where("account =? and disabled = 0", session.User).Get(&user)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		pmailLog.Pop3Errorf(ctx, pmailLog.EventPOP3Cmd, "命令=PASS 数据库查询失败 用户=%s 错误=%v", session.User, err)
	}

	if user.ID > 0 {
		// 使用双算法验证：先bcrypt，后旧MD5
		ok, needsUpgrade := password.Verify(pwd, user.Password)
		if ok {
			// 登录成功，清除速率限制记录
			ratelimit.RecordSuccess(clientIP, session.User)

			// 旧MD5密码自动升级为bcrypt
			if needsUpgrade {
				newHash := password.Encode(pwd)
				_, _ = db.Instance.Table("user").Where("id=?", user.ID).Update(map[string]interface{}{"password": newHash})
			}

			session.Status = gopop.TRANSACTION
			ctx.UserID = user.ID
			ctx.UserName = user.Name
			ctx.UserAccount = user.Account

			pmailLog.Pop3Infof(ctx, pmailLog.EventPOP3AuthSuccess, "用户=%s IP=%s", user.Account, clientIP)
			return nil
		}
	}

	// 认证失败，记录失败
	ratelimit.RecordFailure(clientIP, session.User)
	pmailLog.Pop3Warnf(ctx, pmailLog.EventPOP3AuthFail, "用户=%s IP=%s", session.User, clientIP)
	return errors2.New("password error")
}

// Apop 处理 POP3 APOP 登录命令，包含暴力破解防护。
// 安全策略：拒绝非TLS连接上的认证，防止凭据明文传输（与SMTP AllowInsecureAuth=false对齐）。
func (a action) Apop(session *gopop.Session, username, digest string) error {
	ctx := ensureCtx(session)

	infos := strings.Split(username, "@")
	if len(infos) > 1 {
		username = infos[0]
	}

	pmailLog.Pop3Debugf(ctx, pmailLog.EventPOP3Cmd, "命令=APOP 用户=%s 摘要=%s", username, digest)

	// 强制TLS认证：非加密连接上拒绝提交凭据，客户端必须先执行STLS升级为TLS
	if !session.InTls {
		pmailLog.Pop3Warnf(ctx, pmailLog.EventPOP3AuthRejected, "用户=%s 原因=明文连接", username)
		return errors2.New("must use STLS before authentication")
	}

	// 暴力破解防护：检查速率限制
	clientIP := pop3ClientIP(session)
	if lockErr := ratelimit.Check(clientIP, username); lockErr != nil {
		pmailLog.Pop3Warnf(ctx, pmailLog.EventPOP3RateLimit, "IP=%s 用户=%s 原因=%v", clientIP, username, lockErr)
		return errors2.New("too many failed attempts, try again later")
	}

	// 指数退避延迟
	ratelimit.WaitDelay(clientIP, username)

	var user models.User

	_, err := db.Instance.Where("account =? and disabled = 0", username).Get(&user)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		pmailLog.Pop3Errorf(ctx, pmailLog.EventPOP3Cmd, "命令=APOP 数据库查询失败 用户=%s 错误=%v", username, err)
	}

	if user.ID > 0 && digest == password.Md5Encode(user.Password) {
		// 登录成功，清除速率限制记录
		ratelimit.RecordSuccess(clientIP, username)

		session.User = username
		session.Status = gopop.TRANSACTION
		ctx.UserID = user.ID
		ctx.UserName = user.Name
		ctx.UserAccount = user.Account

		pmailLog.Pop3Infof(ctx, pmailLog.EventPOP3AuthSuccess, "用户=%s IP=%s 方法=APOP", user.Account, clientIP)
		return nil
	}

	// 认证失败，记录失败
	ratelimit.RecordFailure(clientIP, username)
	pmailLog.Pop3Warnf(ctx, pmailLog.EventPOP3AuthFail, "用户=%s IP=%s 方法=APOP", username, clientIP)
	return errors2.New("password error")
}

// Stat 处理 POP3 STAT 命令，查询邮件数量和总大小。
func (a action) Stat(session *gopop.Session) (msgNum, msgSize int64, err error) {
	ctx := ensureCtx(session)
	pmailLog.Pop3Debugf(ctx, pmailLog.EventPOP3Cmd, "命令=STAT")

	num, size := list.Stat(ctx)
	return num, size, nil
}

// Uidl 处理 POP3 UIDL 命令，查询邮件唯一标识符。
func (a action) Uidl(session *gopop.Session, msg string) ([]gopop.UidlItem, error) {
	ctx := ensureCtx(session)
	pmailLog.Pop3Debugf(ctx, pmailLog.EventPOP3Cmd, "命令=UIDL 消息=%s", msg)

	reqId := cast.ToInt64(msg)
	if reqId > 0 {
		return []gopop.UidlItem{
			{
				Id:      reqId,
				UnionId: msg,
			},
		}, nil
	}

	var res []listItem

	emailList, _ := list.GetEmailList(ctx, dto.SearchTag{Type: consts.EmailTypeReceive, Status: -1, GroupId: -1}, "", true, 0, 99999)
	for _, info := range emailList {
		res = append(res, listItem{
			Id:   cast.ToInt64(info.Id),
			Size: cast.ToInt64(info.Size),
		})
	}
	ret := []gopop.UidlItem{}
	for _, re := range res {
		ret = append(ret, gopop.UidlItem{
			Id:      re.Id,
			UnionId: cast.ToString(re.Id),
		})
	}

	return ret, nil
}

type listItem struct {
	Id   int64 `json:"id"`
	Size int64 `json:"size"`
}

// List 处理 POP3 LIST 命令，返回邮件列表。
func (a action) List(session *gopop.Session, msg string) ([]gopop.MailInfo, error) {
	ctx := ensureCtx(session)
	pmailLog.Pop3Debugf(ctx, pmailLog.EventPOP3Cmd, "命令=LIST 消息=%s", msg)

	var res []listItem
	var listId int
	if msg != "" {
		listId = cast.ToInt(msg)
		if listId == 0 {
			return nil, errors.New("params error")
		}
	}

	if listId != 0 {
		info, err := detail.GetEmailDetail(ctx, listId, false)
		if err != nil {
			return nil, err
		}
		item := listItem{
			Id:   cast.ToInt64(info.Id),
			Size: cast.ToInt64(info.Size),
		}
		if item.Size == 0 {
			item.Size = 9999
		}
		res = append(res, item)
	} else {
		emailList, _ := list.GetEmailList(ctx, dto.SearchTag{Type: consts.EmailTypeReceive, Status: -1, GroupId: -1}, "", true, 0, 99999)
		for _, info := range emailList {
			item := listItem{
				Id:   cast.ToInt64(info.Id),
				Size: cast.ToInt64(info.Size),
			}
			if item.Size == 0 {
				item.Size = 9999
			}
			res = append(res, item)
		}
	}
	ret := []gopop.MailInfo{}
	for _, re := range res {
		ret = append(ret, gopop.MailInfo{
			Id:   re.Id,
			Size: re.Size,
		})
	}

	return ret, nil
}

// Retr 处理 POP3 RETR 命令，获取邮件完整内容。
func (a action) Retr(session *gopop.Session, id int64) (string, int64, error) {
	ctx := ensureCtx(session)
	pmailLog.Pop3Debugf(ctx, pmailLog.EventPOP3Cmd, "命令=RETR 消息ID=%d", id)

	email, err := detail.GetEmailDetail(ctx, cast.ToInt(id), false)
	if err != nil {
		pmailLog.Pop3Errorf(ctx, pmailLog.EventPOP3Cmd, "命令=RETR 消息ID=%d 错误=%v", id, err)
		return "", 0, errors.New("server error")
	}

	ret := parsemail.NewEmailFromModel(email.Email).BuildBytes(ctx, false)
	return string(ret), cast.ToInt64(len(ret)), nil
}

// Delete 处理 POP3 DELE 命令，标记邮件为删除。
func (a action) Delete(session *gopop.Session, id int64) error {
	ctx := ensureCtx(session)
	pmailLog.Pop3Debugf(ctx, pmailLog.EventPOP3Cmd, "命令=DELE 消息ID=%d", id)

	session.DeleteIds = append(session.DeleteIds, id)
	session.DeleteIds = array.Unique(session.DeleteIds)
	return nil
}

// Rest 处理 POP3 RSET 命令，重置删除标记。
func (a action) Rest(session *gopop.Session) error {
	ctx := ensureCtx(session)
	pmailLog.Pop3Debugf(ctx, pmailLog.EventPOP3Cmd, "命令=RSET")
	session.DeleteIds = []int64{}
	return nil
}

// Top 处理 POP3 TOP 命令，获取邮件头部和指定行数的内容。
func (a action) Top(session *gopop.Session, id int64, n int) (string, error) {
	ctx := ensureCtx(session)
	pmailLog.Pop3Debugf(ctx, pmailLog.EventPOP3Cmd, "命令=TOP 消息ID=%d 行数=%d", id, n)

	email, err := detail.GetEmailDetail(ctx, cast.ToInt(id), false)
	if err != nil {
		pmailLog.Pop3Errorf(ctx, pmailLog.EventPOP3Cmd, "命令=TOP 消息ID=%d 错误=%v", id, err)
		return "", errors2.New("server error")
	}

	ret := parsemail.NewEmailFromModel(email.Email).BuildBytes(ctx, false)
	res := strings.Split(string(ret), "\n")
	headerEndLine := len(res) - 1
	for i, re := range res {
		if re == "\r" {
			headerEndLine = i
			break
		}
	}
	if len(res) <= headerEndLine+n+1 {
		return string(ret), nil
	}

	lines := array.Join(res[0:headerEndLine+n+1], "\n")
	return lines, nil
}

// Noop 处理 POP3 NOOP 命令，保持连接活跃。
func (a action) Noop(session *gopop.Session) error {
	ctx := ensureCtx(session)
	pmailLog.Pop3Debugf(ctx, pmailLog.EventPOP3Cmd, "命令=NOOP")
	return nil
}

// Quit 处理 POP3 QUIT 命令，执行标记删除并关闭会话。
func (a action) Quit(session *gopop.Session) error {
	ctx := ensureCtx(session)
	pmailLog.Pop3Debugf(ctx, pmailLog.EventPOP3Cmd, "命令=QUIT")

	var DelIds []int

	if len(session.DeleteIds) > 0 {
		for _, delId := range session.DeleteIds {
			DelIds = append(DelIds, cast.ToInt(delId))
		}

		del_email.DelEmail(ctx, DelIds, false)
	}

	return nil
}
