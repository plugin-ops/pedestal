package rule

import (
	"context"
	"errors"
	"github.com/plugin-ops/pedestal/pedestal/action"
	"strings"
)

type Rule interface {
	Info() Info
	Set(recipient string, dependency action.Action) error
	Get(name string) (*action.Value, error) // Generally, you can get it after the rule is executed.
	Do(ctx context.Context) error
	Compile() error // be careful not to recompile
}

type Info interface {
	Name() string
	Description() string
	Version() float64
	OriginalContent() string
	Author() string
	GetRelyOn() map[string]string // map[dependency]recipient
}

type info struct {
	name    string
	desc    string
	author  string
	version float64
	content string
	relyOn  map[string]string
}

func (g *info) Name() string {
	return g.name
}

func (g *info) Description() string {
	return g.desc
}

func (g *info) Version() float64 {
	return g.version
}

func (g *info) OriginalContent() string {
	return g.content
}

func (g *info) Author() string {
	return g.author
}

func (g *info) GetRelyOn() map[string]string {
	return g.relyOn
}

const (
	BODY_TAG = "//--body--"
)

const ( // The pedestal will take the value of the parameter below from the scriptInfo
	SCRIPT_TAG_NAME        = "script_name"
	SCRIPT_TAG_RELY_ON     = "script_rely_on" // json format
	SCRIPT_TAG_VERSION     = "script_version"
	SCRIPT_TAG_DESCRIPTION = "script_description"
)

func getInfoContent(ruleContent string) string {
	index := strings.Index(ruleContent, BODY_TAG)
	if index == -1 {
		return ""
	}
	return ruleContent[:index]
}

var (
	ErrNotRuleName = errors.New("unKnown rule name")
)
