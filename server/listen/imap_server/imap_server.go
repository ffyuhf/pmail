package imap_server

import (
	"crypto/tls"
	"os"

	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapserver"
	"github.com/ffyuhf/pmail/config"
	pmailLog "github.com/ffyuhf/pmail/utils/log"
)

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
		Caps: imap.CapSet{
			imap.CapIMAP4rev1: {},
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
