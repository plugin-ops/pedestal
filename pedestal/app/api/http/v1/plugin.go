package v1

import (
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/plugin-ops/pedestal/pedestal/app/api/http/base"
	server "github.com/plugin-ops/pedestal/pedestal/app/server/v1"
	"github.com/plugin-ops/pedestal/pedestal/config"
)

func ReloadAllPlugins(r *ghttp.Request) {
	base.SendResponseExit(r, base.NewBaseReq(server.ReloadAllPlugins()))
}

func RemovePlugin(r *ghttp.Request) {
	req := new(server.RemovePluginReqV1)
	base.BindRequestParams(r, req)
	server.RemovePlugin(req)
	base.SendResponseExit(r, base.NewBaseReq(nil))
}

type AddPluginReqV1 struct {
	PluginFile *ghttp.UploadFile
}

func AddPlugin(r *ghttp.Request) {
	req := new(AddPluginReqV1)
	base.BindRequestParams(r, req)

	name, err := req.PluginFile.Save(config.PluginDir, false)
	if err != nil {
		base.SendResponseExit(r, base.NewBaseReq(err))
	}

	base.SendResponseExit(r, base.NewBaseReq(server.AddPlugin(name)))
}
