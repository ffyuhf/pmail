package main

import (
	"fmt"

	"github.com/ffyuhf/pmail/config"
	"github.com/ffyuhf/pmail/listen/cron_server"
	"github.com/ffyuhf/pmail/res_init"
	log "github.com/sirupsen/logrus"
)

var (
	gitHash   string
	buildTime string
	goVersion string
	version   string
)

func main() {

	config.Init()

	if version == "" {
		version = "TestVersion"
	}

	// 服务器启动横幅：输出构建版本信息，便于运维排查问题
	banner := fmt.Sprintf(
		"*******************************************************************\n"+
			"***        服务启动成功\n"+
			"***        服务版本: %s\n"+
			"***        Git提交哈希: %s\n"+
			"***        构建日期: %s\n"+
			"***        Go版本: %s\n"+
			"*******************************************************************",
		version, gitHash, buildTime, goVersion,
	)
	log.Info(banner)

	// 定时任务启动
	go cron_server.Start()

	// 核心服务启动
	res_init.Init(version)

	log.Info("服务已停止")
}
