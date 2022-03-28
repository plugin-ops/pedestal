package log

import (
	"github.com/gogf/gf/v2/os/glog"
	"github.com/plugin-ops/pedestal/pedestal/config"
)

func init() {
	initLogger()
}

type logKey string

const logContextKey logKey = "log_key"

func initLogger() {
	glog.SetCtxKeys(logContextKey)
}

func InitLoggerOnPedestal() error {
	if err := glog.SetPath(config.LogDir); err != nil {
		return err
	}
	glog.SetFile("pedestal-{Y-m-d}.log")
	glog.SetStdoutPrint(false)
	return nil
}
