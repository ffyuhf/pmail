package imap_server

import (
	"sync"
	"time"

	"github.com/emersion/go-imap/v2/imapserver"
	"github.com/ffyuhf/pmail/models"
	"github.com/ffyuhf/pmail/utils/context"
	pmailLog "github.com/ffyuhf/pmail/utils/log"
	"github.com/spf13/cast"
)

var userConnects sync.Map

// Idle 处理 IMAP IDLE 命令，保持连接并等待新邮件通知。
// 记录 IDLE 进入和结束事件，便于诊断长时间空闲连接问题。
func (s *serverSession) Idle(w *imapserver.UpdateWriter, stop <-chan struct{}) error {
	connects, ok := userConnects.Load(s.ctx.UserID)
	logId := cast.ToString(s.ctx.GetValue(context.LogID))

	pmailLog.ImapDebugf(s.ctx, pmailLog.EventIMAPIdleEnter, "用户=%s 文件夹=%s", s.ctx.UserAccount, s.currentMailbox)

	idleStart := time.Now()

	if !ok {

		connects = map[string]*imapserver.UpdateWriter{
			logId: w,
		}
		userConnects.Store(s.ctx.UserID, connects)
	} else {
		connects := connects.(map[string]*imapserver.UpdateWriter)
		if _, ok := connects[logId]; !ok {
			connects[logId] = w
			userConnects.Store(s.ctx.UserID, connects)
		}
	}

	go func() {
		<-stop
		duration := time.Since(idleStart)
		pmailLog.ImapInfof(s.ctx, pmailLog.EventIMAPIdleTimeout, "用户=%s 文件夹=%s 持续时间=%v", s.ctx.UserAccount, s.currentMailbox, duration)
		userConnects.Delete(logId)
	}()

	return nil
}

// IdleNotice 向指定用户的活跃 IMAP IDLE 连接推送新邮件通知。
func IdleNotice(ctx *context.Context, userId int, email *models.Email) error {
	if userId == 0 || email == nil || email.Id == 0 {
		return nil
	}

	connects, ok := userConnects.Load(userId)
	if ok {
		connects := connects.(map[string]*imapserver.UpdateWriter)
		for _, connect := range connects {
			connect.WriteNumMessages(1)
		}
	}
	return nil
}
