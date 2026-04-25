package res_init

import (
	"encoding/json"
	"os"
	"time"

	"fmt"

	"github.com/ffyuhf/pmail/config"
	"github.com/ffyuhf/pmail/db"
	"github.com/ffyuhf/pmail/dto/parsemail"
	"github.com/ffyuhf/pmail/hooks"
	"github.com/ffyuhf/pmail/listen/http_server"
	"github.com/ffyuhf/pmail/listen/imap_server"
	"github.com/ffyuhf/pmail/listen/pop3_server"
	"github.com/ffyuhf/pmail/listen/smtp_server"
	"github.com/ffyuhf/pmail/services/auth"
	"github.com/ffyuhf/pmail/services/setup/ssl"
	"github.com/ffyuhf/pmail/session"
	"github.com/ffyuhf/pmail/signal"
	"github.com/ffyuhf/pmail/utils/file"
	"github.com/ffyuhf/pmail/utils/ratelimit"
	log "github.com/sirupsen/logrus"

	// 新增：HTTP 就绪探测
	"net/http"
)

func Init(serverVersion string) {

	if !config.IsInit {
		dirInit()

		go http_server.SetupStart()
		<-signal.InitChan
		http_server.SetupStop()
	}

	for {
		config.Init()
		// 移除：启动前立即更新证书的调用，避免 HTTP 未启动时触发 ACME
		// ssl.Update(false)

		// DKIM 密钥自检：确保密钥文件存在，不存在则自动生成 2048 位 RSA 密钥对
		auth.DkimGen()

		parsemail.Init()
		err := db.Init(serverVersion)
		if err != nil {
			panic(err)
		}
		session.Init()
		ratelimit.Init()
		hooks.Init(serverVersion)
		// 启动 SMTP 服务器
		go smtp_server.Start()
		go smtp_server.StartWithTLS()
		go smtp_server.StartWithTLSNew()
		// 启动 HTTP 服务器
		go http_server.HttpsStart()
		go http_server.HttpStart()

		// 等待 HTTP 服务就绪后再执行证书检查/续期，避免 ACME 验证失败
		waitHTTPReady()

		// 启动后检查一遍证书（非重启模式）
		ssl.Update(false)

		// 启动 POP3 服务器
		go pop3_server.Start()
		go pop3_server.StartWithTls()
		// 启动 IMAP 服务器
		go imap_server.StarTLS()

		configStr, _ := json.Marshal(config.Instance)
		log.Warnf("Config File Info:  %s", configStr)

		select {
		case <-signal.RestartChan:
			log.Infof("Server Restart!")
			smtp_server.Stop()
			http_server.HttpsStop()
			http_server.HttpStop()
			pop3_server.Stop()
			imap_server.Stop()
			hooks.Stop()
		case <-signal.StopChan:
			log.Infof("Server Stop!")
			smtp_server.Stop()
			http_server.HttpsStop()
			http_server.HttpStop()
			pop3_server.Stop()
			imap_server.Stop()
			hooks.Stop()
			return
		}
		log.Infof("Server Stop Success!")
		time.Sleep(5 * time.Second)

	}

}

func dirInit() {
	if !file.PathExist("./config") {
		err := os.MkdirAll("./config", 0744)
		if err != nil {
			panic(err)
		}
	}

	if !file.PathExist("./config/dkim") {
		err := os.MkdirAll("./config/dkim", 0744)
		if err != nil {
			panic(err)
		}
	}

	if !file.PathExist("./config/ssl") {
		err := os.MkdirAll("./config/ssl", 0744)
		if err != nil {
			panic(err)
		}
	}
}

// 新增：HTTP 就绪探测（最多等待 ~90 秒）
func waitHTTPReady() {
	port := 80
	if config.Instance != nil && config.Instance.HttpPort > 0 {
		port = config.Instance.HttpPort
	}
	url := fmt.Sprintf("http://127.0.0.1:%d/api/ping", port)

	for i := 0; i < 90; i++ {
		resp, err := http.Get(url)
		if err == nil && resp != nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			log.Infof("HTTP ready: %s", url)
			return
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(1 * time.Second)
	}
	log.Warnf("HTTP not ready after 90s, skipping immediate SSL update")
}
