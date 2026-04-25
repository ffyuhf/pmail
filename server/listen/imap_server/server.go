package imap_server

import (
	"sync"
	"time"

	"github.com/ffyuhf/pmail/utils/context"
	"github.com/ffyuhf/pmail/utils/id"
	"github.com/ffyuhf/pmail/utils/ratelimit"
	"github.com/emersion/go-imap/v2"
	log "github.com/sirupsen/logrus"

	"github.com/emersion/go-imap/v2/imapserver"
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

// NewSession 创建新的 IMAP 会话，从连接中提取远程 IP。
func (s *Server) NewSession(remoteAddr string) imapserver.Session {
	tc := &context.Context{}
	tc.SetValue(context.LogID, id.GenLogID())

	return &serverSession{
		server:      s,
		ctx:         tc,
		connectTime: time.Now(),
		remoteAddr:  ratelimit.ExtractIP(remoteAddr),
	}
}

func (s *serverSession) Close() error {
	return nil
}

func (s *serverSession) Subscribe(mailbox string) error {
	return nil
}

func (s *serverSession) Unsubscribe(mailbox string) error {
	return nil
}

func (s *serverSession) Append(mailbox string, r imap.LiteralReader, options *imap.AppendOptions) (*imap.AppendData, error) {
	log.WithContext(s.ctx).Errorf("Append Not Implemented")
	return nil, nil
}

func (s *serverSession) Unselect() error {
	s.currentMailbox = ""
	return nil
}
