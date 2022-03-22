package controller

import (
	"errors"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/plugin-ops/pedestal/pedestal/config"
	"github.com/plugin-ops/pedestal/pedestal/execute"
	"github.com/plugin-ops/pedestal/pedestal/plugin"
	"github.com/plugin-ops/pedestal/pedestal/rule"
)

var V1Api = &v1{}

type v1 struct{}

func (v *v1) ReloadAllPlugins(r *ghttp.Request) {
	err := plugin.ReLoadPluginWithDir(config.PluginDir)
	SendResponseExit(r, NewBaseReq(err))
}

type RunRuleReqV1 struct {
	RuleContent string     `json:"rule_content"`
	RuleType    string     `json:"rule_type" enum:"go"`
	RunParams   []*KVReqV1 `json:"run_params"`
}

type KVReqV1 struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type RunRuleResV1 struct {
	BaseRes
	TaskID string `json:"task_id"`
}

func (v *v1) RunRule(r *ghttp.Request) {
	req := new(RunRuleReqV1)
	err := r.Parse(req)
	if err != nil {
		SendResponseExit(r, NewBaseReq(err))
	}

	var ru rule.Rule
	switch rule.RuleType(req.RuleType) {
	case rule.RuleTypeGo:
		ru, err = rule.NewGolang(req.RuleContent)
	default:
		err = errors.New("known rule type")
	}
	if err != nil {
		SendResponseExit(r, NewBaseReq(err))
	}
	params := make(map[string]interface{})
	for _, p := range req.RunParams {
		params[p.Key] = p.Value
	}

	id, err := execute.GetExecutor().Add(ru, params, nil)
	SendResponseExit(r, &RunRuleResV1{
		BaseRes: NewBaseReq(err),
		TaskID:  id,
	})
}
