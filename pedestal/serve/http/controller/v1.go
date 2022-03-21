package controller

import (
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/plugin-ops/pedestal/pedestal/config"
	"github.com/plugin-ops/pedestal/pedestal/plugin"
)

var V1Api = &v1{}

type v1 struct{}

func (v *v1) ReloadAllPlugins(r *ghttp.Request) {
	err := plugin.ReLoadPluginWithDir(config.PluginDir)
	SendResponseExit(r, NewBaseReq(err))
}
