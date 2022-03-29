package base

import "github.com/gogf/gf/v2/net/ghttp"

type BaseRes struct {
	Code    int    `json:"code" example:"0"`
	Message string `json:"message" example:"ok"`
}

func NewBaseReq(err error) BaseRes {
	res := BaseRes{}

	if err == nil {
		res.Code = 0
		res.Message = "ok"
	} else {
		res.Code = -1
		res.Message = err.Error()
	}

	return res
}

func SendResponseExit(r *ghttp.Request, resp interface{}) {
	err := r.Response.WriteJsonExit(resp)
	if err != nil {
		_ = r.Response.WriteJsonExit(NewBaseReq(err))
	}
}

// BindRequestParams failure will return an error directly to the interface
func BindRequestParams(r *ghttp.Request, pointer interface{}) {
	err := r.Parse(pointer)
	if err != nil {
		SendResponseExit(r, NewBaseReq(err))
	}
}
