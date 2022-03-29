package v1

import (
	"github.com/plugin-ops/pedestal/pedestal/app/api/http/base"
	server "github.com/plugin-ops/pedestal/pedestal/app/server/v1"

	"github.com/gogf/gf/v2/net/ghttp"
)

type ListActionResV1 struct {
	base.BaseRes
	*server.ListActionResV1
}

func ListAction(r *ghttp.Request) {
	res, err := server.ListAction()
	if err != nil {
		base.SendResponseExit(r, base.NewBaseReq(err))
	}

	base.SendResponseExit(r, &ListActionResV1{
		BaseRes:         base.NewBaseReq(nil),
		ListActionResV1: res,
	})
}
