package imap_server

import (
	"sync"
	"time"

	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapserver"
	"github.com/ffyuhf/pmail/utils/context"
	"github.com/ffyuhf/pmail/utils/id"
	pmailLog "github.com/ffyuhf/pmail/utils/log"
	"github.com/ffyuhf/pmail/utils/ratelimit"
	log "github.com/sirupsen/logrus"
)

// Server 是一个服务器实例。
//
// 服务器包含一个用户列表。
type Server struct {
	mutex sync.Mutex
}

// NewServer 创建一个新的服务器。
func NewServer() *Server {
	return &Server{}
}

type Status int8

const (
	UNAUTHORIZED Status = 1
	AUTHORIZED   Status = 2
	SELECTED     Status = 3
	LOGOUT       Status = 4
)

type serverSession struct {
	server         *Server // immutable
	ctx            *context.Context
	status         Status
	currentMailbox string
	connectTime    time.Time
	deleteUidList  []int
	remoteAddr     string // 客户端 IP 地址，用于暴力破解防护
}

// NewSession 创建新的 IMAP 会话，设置协议标识并记录连接事件。
func (s *Server) NewSession(remoteAddr string) imapserver.Session {
	tc := &context.Context{}
	tc.SetValue(context.LogID, id.GenLogID())
	// 设置协议标识，用于日志格式化器输出 [IMAP] 前缀
	tc.Protocol = pmailLog.ProtocolIMAP
	tc.ClientIP = ratelimit.ExtractIP(remoteAddr)

	pmailLog.ImapInfof(tc, pmailLog.EventIMAPSessionNew, "客户端IP=%s", tc.ClientIP)

	return &serverSession{
		server:      s,
		ctx:         tc,
		connectTime: time.Now(),
		remoteAddr:  ratelimit.ExtractIP(remoteAddr),
	}
}

// Close 关闭 IMAP 会话，记录会话持续时长。
func (s *serverSession) Close() error {
	duration := time.Since(s.connectTime)
	pmailLog.ImapInfof(s.ctx, pmailLog.EventIMAPSessionClose, "用户=%s 持续时间=%v", s.ctx.UserAccount, duration)
	return nil
}

func (s *serverSession) Subscribe(mailbox string) error {
	return nil
}

func (s *serverSession) Unsubscribe(mailbox string) error {
	return nil
}

func (s *serverSession) Append(mailbox string, r imap.LiteralReader, options *imap.AppendOptions) (*imap.AppendData, error) {
	log.WithContext(s.ctx).Errorf("Append 功能未实现")
	return nil, nil
}

func (s *serverSession) Unselect() error {
	s.currentMailbox = ""
	return nil
}
