package v1

import (
	"github.com/plugin-ops/pedestal/pedestal/app/api/http/base"
	server "github.com/plugin-ops/pedestal/pedestal/app/server/v1"

	"github.com/gogf/gf/v2/net/ghttp"
)

func AddRule(r *ghttp.Request) {
	req := new(server.AddRuleReqV1)
	base.BindRequestParams(r, req)
	base.SendResponseExit(r, base.NewBaseReq(server.AddRule(req)))
}

type RunRuleResV1 struct {
	base.BaseRes
	*server.RunRuleResV1
}

func RunRule(r *ghttp.Request) {
	req := new(server.RunRuleReqV1)
	base.BindRequestParams(r, req)
	res, err := server.RunRule(req)
	base.SendResponseExit(r, &RunRuleResV1{
		BaseRes:      base.NewBaseReq(err),
		RunRuleResV1: res,
	})
}

func ReloadAllRule(r *ghttp.Request) {
	base.SendResponseExit(r, base.NewBaseReq(server.ReloadAllRule()))
}
