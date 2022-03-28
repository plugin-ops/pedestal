package rule

import (
	"context"
	"errors"
	"fmt"
	"github.com/plugin-ops/pedestal/pedestal/action"
	"strings"
)

type Rule interface {
	Info() Info
	Set(recipient string, value *action.Value) error
	AddRelyOn(recipient string, dependency action.Action) error
	Get(name string) (*action.Value, error) // Generally, you can get it after the rule is executed.
	Do(ctx context.Context) error
	Compile() error // be careful not to recompile
}

type Info interface {
	Name() string
	Description() string
	Version() float32
	OriginalContent() string
	Author() string
	RuleType() RuleType
	GetRelyOn() map[string]string // map[dependency]recipient
	GetParams() map[string]string // the dynamic parameters and default values that the rule can accept are recorded here
	Key() string
}

type info struct {
	name               string
	desc               string
	author             string
	ruleType           RuleType
	version            float32
	content            string
	contentDescription string
	relyOn             map[string]string
	params             map[string]string
}

func (g *info) Name() string {
	return g.name
}

func (g *info) Description() string {
	return g.desc
}

func (g *info) Version() float32 {
	return g.version
}

func (g *info) OriginalContent() string {
	return g.content
}

func (g *info) DescContent() string {
	return g.contentDescription
}

func (g *info) Author() string {
	return g.author
}

func (g *info) RuleType() RuleType {
	return g.ruleType
}

func (g *info) GetRelyOn() map[string]string {
	return g.relyOn
}

func (g *info) GetParams() map[string]string {
	return g.params
}

func (g *info) Key() string {
	return fmt.Sprintf("%v@%v", g.name, g.version)
}

const (
	BODY_TAG = "//--body--"
)

const ( // The pedestal will take the value of the parameter below from the ruleInfo
	RULE_TAG_NAME        = "rule_name"
	RULE_TAG_RELY_ON     = "rule_rely_on" // json format
	RULE_TAG_VERSION     = "rule_version"
	RULE_TAG_DESCRIPTION = "rule_description"
	RULE_TAG_PARAMS      = "rule_params" //json format
)

func getInfoContent(ruleContent string) string {
	index := strings.Index(ruleContent, BODY_TAG)
	if index == -1 {
		return ""
	}
	return ruleContent[:index]
}

var (
	ErrNotRuleName       = errors.New("unknown rule name")
	ErrDependencyFormat  = errors.New("dependency format error")
	ErrorUnknownRuleType = errors.New("unknown rule type")
)

type RuleConfig map[string]interface{}

type RuleType string

const (
	RuleTypeGo = "go"
)
