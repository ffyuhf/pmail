package smtp_server

import (
	"crypto/tls"
	"time"

	"github.com/emersion/go-smtp"
	"github.com/ffyuhf/pmail/config"
	log "github.com/sirupsen/logrus"
)

var instance *smtp.Server
var instanceTls *smtp.Server
var instanceTlsNew *smtp.Server

func StartWithTLSNew() {
	be := &Backend{}

	instanceTlsNew = smtp.NewServer(be)

	instanceTlsNew.Addr = ":587"
	instanceTlsNew.Domain = config.Instance.Domain
	// 修复：将超时从10秒增加到60秒，避免大HTML邮件传输超时被中断
	instanceTlsNew.ReadTimeout = 60 * time.Second
	instanceTlsNew.WriteTimeout = 60 * time.Second
	instanceTlsNew.MaxMessageBytes = 1024 * 1024 * 30
	instanceTlsNew.MaxRecipients = 50
	// 强制 TLS 认证：拒绝未加密连接上的认证
	instanceTlsNew.AllowInsecureAuth = false
	// 加载证书和密钥
	cer, err := tls.LoadX509KeyPair(config.Instance.SSLPublicKeyPath, config.Instance.SSLPrivateKeyPath)
	if err != nil {
		log.Fatal(err)
		return
	}
	// 配置 STARTTLS 的 TLS 支持
	instanceTlsNew.TLSConfig = &tls.Config{Certificates: []tls.Certificate{cer}}

	log.Println("Starting Smtp With STARTTLS Server Port:", instanceTlsNew.Addr)
	// 587端口使用STARTTLS（先明文连接，再升级TLS），而非隐式TLS
	if err := instanceTlsNew.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

func StartWithTLS() {
	be := &Backend{}

	instanceTls = smtp.NewServer(be)

	instanceTls.Addr = ":465"
	instanceTls.Domain = config.Instance.Domain
	// 修复：将超时从10秒增加到60秒，避免大HTML邮件传输超时被中断
	instanceTls.ReadTimeout = 60 * time.Second
	instanceTls.WriteTimeout = 60 * time.Second
	instanceTls.MaxMessageBytes = 1024 * 1024 * 30
	instanceTls.MaxRecipients = 50
	// 强制 TLS 认证：拒绝未加密连接上的认证
	instanceTls.AllowInsecureAuth = false
	// 加载证书和密钥
	cer, err := tls.LoadX509KeyPair(config.Instance.SSLPublicKeyPath, config.Instance.SSLPrivateKeyPath)
	if err != nil {
		log.Fatal(err)
		return
	}
	// 配置 TLS 支持
	instanceTls.TLSConfig = &tls.Config{Certificates: []tls.Certificate{cer}}

	log.Println("Starting Smtp With SSL Server Port:", instanceTls.Addr)
	if err := instanceTls.ListenAndServeTLS(); err != nil {
		log.Fatal(err)
	}
}

func Start() {
	be := &Backend{}

	instance = smtp.NewServer(be)

	instance.Addr = ":25"
	instance.Domain = config.Instance.Domain
	// 修复：将超时从10秒增加到60秒，避免大HTML邮件传输超时被中断
	instance.ReadTimeout = 60 * time.Second
	instance.WriteTimeout = 60 * time.Second
	instance.MaxMessageBytes = 1024 * 1024 * 30
	instance.MaxRecipients = 50
	// 强制 TLS 认证
	instance.AllowInsecureAuth = false
	// 加载证书和密钥
	cer, err := tls.LoadX509KeyPair(config.Instance.SSLPublicKeyPath, config.Instance.SSLPrivateKeyPath)
	if err != nil {
		log.Fatal(err)
		return
	}
	// 配置 TLS 支持
	instance.TLSConfig = &tls.Config{Certificates: []tls.Certificate{cer}}

	log.Println("Starting Smtp Server Port:", instance.Addr)
	if err := instance.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

func Stop() {
	if instance != nil {
		instance.Close()
	}
	if instanceTls != nil {
		instanceTls.Close()
	}

	if instanceTlsNew != nil {
		instanceTlsNew.Close()
	}
}
