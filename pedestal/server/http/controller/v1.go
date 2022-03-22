package controller

import (
	"errors"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/plugin-ops/pedestal/pedestal/config"
	"github.com/plugin-ops/pedestal/pedestal/execute"
	"github.com/plugin-ops/pedestal/pedestal/plugin"
	"github.com/plugin-ops/pedestal/pedestal/rule"
	"path"
)

var V1Api = &v1{}

type v1 struct{}

func (*v1) ReloadAllPlugins(r *ghttp.Request) {
	err := plugin.ReLoadPluginWithDir(config.PluginDir)
	SendResponseExit(r, NewBaseReq(err))
}

type RemovePluginReqV1 struct {
	ActionName string `json:"action_name"`
}

func (*v1) RemovePlugin(r *ghttp.Request) {
	req := new(RemovePluginReqV1)
	BindRequestParams(r, req)

	plugin.RemovePlugin(req.ActionName)
	SendResponseExit(r, NewBaseReq(nil))
}

type AddPluginReqV1 struct {
	PluginFile *ghttp.UploadFile
}

func (*v1) AddPlugin(r *ghttp.Request) {
	req := new(AddPluginReqV1)
	BindRequestParams(r, req)

	name, err := req.PluginFile.Save(config.PluginDir, true)
	if err != nil {
		SendResponseExit(r, NewBaseReq(err))
	}
	err = plugin.AddPlugin(path.Join(config.PluginDir, name))
	if err != nil {
		SendResponseExit(r, NewBaseReq(err))
	}
	err = plugin.CleanUselessPluginFile()
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

func (*v1) RunRule(r *ghttp.Request) {
	req := new(RunRuleReqV1)
	BindRequestParams(r, req)

	var ru rule.Rule
	var err error
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
