package http_server

import (
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"time"

	"github.com/ffyuhf/pmail/config"
	"github.com/ffyuhf/pmail/controllers"
	"github.com/ffyuhf/pmail/utils/ip"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cast"
)

// 项目初始化引导用的服务，初始化引导结束后即退出
var setupServer *http.Server

func SetupStart() {
	// 生成 Setup Token，用于保护初始化接口不被未授权访问。
	// Token 会打印在启动日志中，用户通过带 Token 的 URL 访问 Setup 页面。
	// 修改日期: 20260425
	token := config.GenerateSetupToken()

	mux := http.NewServeMux()
	fe, err := fs.Sub(local, "dist")
	if err != nil {
		panic(err)
	}
	mux.Handle("/", http.FileServer(http.FS(fe)))
	mux.HandleFunc("/api/", contextIterceptor(controllers.Setup))
	// 挑战请求类似这样 /.well-known/acme-challenge/QPyMAyaWw9s5JvV1oruyqWHG7OqkHMJEHPoUz2046KM
	mux.HandleFunc("/.well-known/", controllers.AcmeChallenge)

	HttpPort := 80
	flag.IntVar(&HttpPort, "p", 80, "初始化阶段Http服务端口")
	flag.Parse()

	if HttpPort == 80 {
		HttpPort = cast.ToInt(os.Getenv("setup_port"))
	}

	if HttpPort <= 0 || HttpPort > 65535 {
		HttpPort = 80
	}

	config.Instance.SetSetupPort(HttpPort)
	log.Infof("HttpServer Start On Port :%d", HttpPort)
	if HttpPort == 80 {
		log.Infof("Please click http://%s/?token=%s to continue.\n", ip.GetIp(), token)
	} else {
		log.Infof("Please click http://%s:%d/?token=%s to continue.", ip.GetIp(), HttpPort, token)
	}

	setupServer = &http.Server{
		Addr:         fmt.Sprintf(":%d", HttpPort),
		Handler:      mux,
		ReadTimeout:  time.Second * 60,
		WriteTimeout: time.Second * 60,
	}
	err = setupServer.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		panic(err)
	}
}

func SetupStop() {
	err := setupServer.Close()
	log.Infof("Setup End!")
	if err != nil {
		panic(err)
	}
}
