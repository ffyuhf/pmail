package setup

import (
	"github.com/ffyuhf/pmail/config"
	"github.com/ffyuhf/pmail/signal"
	"github.com/ffyuhf/pmail/utils/errors"
)

// Finish 标记初始化完成
func Finish() error {
	cfg, err := config.ReadConfig()
	if err != nil {
		return errors.Wrap(err)
	}
	cfg.IsInit = true

	err = config.WriteConfig(cfg)
	if err != nil {
		return errors.Wrap(err)
	}
	// 初始化完成
	signal.InitChan <- true
	return nil
}
