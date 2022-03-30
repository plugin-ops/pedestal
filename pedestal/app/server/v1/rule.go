package v1

import (
	"errors"
	"fmt"
	"github.com/plugin-ops/pedestal/pedestal/config"

	"github.com/plugin-ops/pedestal/pedestal/execute"
	"github.com/plugin-ops/pedestal/pedestal/log"
	"github.com/plugin-ops/pedestal/pedestal/rule"
)

type AddRuleReqV1 struct {
	RuleContent string `json:"rule_content"`
	RuleType    string `json:"rule_type" enum:"go"`
}

func AddRule(req *AddRuleReqV1) error {
	stage := log.NewStage().Enter("AddRule")

	var ru rule.Rule
	var err error
	switch rule.RuleType(req.RuleType) {
	case rule.RuleTypeGo:
		ru, err = rule.NewGolang(stage, req.RuleContent)
	default:
		err = errors.New("known rule type")
	}
	if err != nil {
		return err
	}
	rule.RegistryRule(stage, ru.Info())
	return nil
}

type RunRuleReqV1 struct {
	RuleName    string     `json:"rule_name"`
	RuleVersion float32    `json:"rule_version"`
	RunParams   []*KVReqV1 `json:"run_params"`
}

type KVReqV1 struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type RunRuleResV1 struct {
	TaskID string `json:"task_id"`
}

func RunRule(req *RunRuleReqV1) (*RunRuleResV1, error) {
	stage := log.NewStage().Enter("RunRule")

	ru, exist, err := rule.GetRule(stage, req.RuleName, req.RuleVersion)
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, fmt.Errorf("rule [%v@%v] not exist", req.RuleName, req.RuleVersion)
	}
	params := make(map[string]interface{})
	for _, p := range req.RunParams {
		params[p.Key] = p.Value
	}

	id, err := execute.GetExecutor().AddTask(ru, params, nil)
	return &RunRuleResV1{
		TaskID: id,
	}, err
}

func ReloadAllRule() error {
	stage := log.NewStage().Enter("ReloadAllRule")

	return rule.ReLoadRuleWithDir(stage, config.RuleDir)
}
