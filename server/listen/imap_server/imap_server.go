package imap_server

import (
	"crypto/tls"
	"os"

	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapserver"
	"github.com/ffyuhf/pmail/config"
	pmailLog "github.com/ffyuhf/pmail/utils/log"
)

// 修复 BUG-1/4/5/6/7（2026-06-09）：
// 在 Caps 中声明所有已实现但未公布的 IMAP 能力，确保客户端正确识别服务器功能。
// 新增：IDLE、NAMESPACE、UIDPLUS、MOVE、CHILDREN。
//
// RFC 合规性扩展（2026-06-09）：
//   - P0-2：添加 CapLiteralPlus（RFC 7888），支持无限制非同步字面量
//   - P1-1：添加 CapESearch（RFC 4731），支持扩展搜索结果（Min/Max/Count）

var instanceTLS *imapserver.Server

func Stop() {
	if instanceTLS != nil {
		instanceTLS.Close()
		instanceTLS = nil
	}
}

// StarTLS 启动TLS端口监听，不加密的代码就懒得写了
func StarTLS() {

	crt, err := tls.LoadX509KeyPair(config.Instance.SSLPublicKeyPath, config.Instance.SSLPrivateKeyPath)
	if err != nil {
		panic(err)
	}
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{crt},
	}

	memServer := NewServer()

	option := &imapserver.Options{
		NewSession: func(conn *imapserver.Conn) (imapserver.Session, *imapserver.GreetingData, error) {
			remoteAddr := ""
			if conn.NetConn() != nil {
				remoteAddr = conn.NetConn().RemoteAddr().String()
			}
			return memServer.NewSession(remoteAddr), nil, nil
		},
		// BUG-1: IDLE — 客户端无法识别 IDLE 支持
		// BUG-4: NAMESPACE — 已实现 session_namespace.go 但未公布
		// BUG-5: UIDPLUS — COPY 返回 CopyData 但未公布
		// BUG-6: MOVE — 已实现 session_move.go 但未公布
		// BUG-7: CHILDREN — LIST 返回 HasChildren 属性但未公布
		Caps: imap.CapSet{
			imap.CapIMAP4rev1:   {},
			imap.CapIdle:        {},
			imap.CapNamespace:   {},
			imap.CapUIDPlus:     {},
			imap.CapMove:        {},
			imap.CapChildren:    {},
			imap.CapLiteralPlus: {}, // P0-2（RFC 7888）：非同步字面量无大小限制
			imap.CapESearch:     {}, // P1-1（RFC 4731）：扩展搜索，返回 Min/Max/Count
		},
		TLSConfig:    tlsConfig,
		InsecureAuth: false,
	}

	if config.Instance.LogLevel == "debug" {
		option.DebugWriter = os.Stdout
	}

	instanceTLS = imapserver.New(option)
	pmailLog.ImapInfof(nil, pmailLog.EventIMAPSessionNew, "IMAP服务启动 端口=993 模式=TLS")
	if err := instanceTLS.ListenAndServeTLS(":993"); err != nil {
		panic(err)
	}
}
