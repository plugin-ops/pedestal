package rule

import (
	"context"
	"fmt"
	"reflect"

	"github.com/plugin-ops/pedestal/pedestal/action"
	"github.com/plugin-ops/pedestal/pedestal/log"

	"github.com/gogf/gf/v2/os/glog"
	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
)

type Golang struct {
	info        *info
	program     *interp.Program
	interpreter *interp.Interpreter
	relyOn      map[string]reflect.Value
	err         error
	hasError    *bool
	stage       *log.Stage
}

func NewGolang(stage *log.Stage, content string) (*Golang, error) {
	g := &Golang{
		info: &info{
			content:  content,
			ruleType: RuleTypeGo,
			relyOn:   map[string]string{},
			params:   map[string]string{},
		},
		interpreter: interp.New(interp.Options{}),
		relyOn:      map[string]reflect.Value{},
		hasError:    new(bool),
		stage:       stage,
	}
	err := g.parseInfo()
	if err != nil {
		return nil, err
	}
	return g, nil
}

func (g *Golang) Info() Info {
	return g.info
}

func (g *Golang) Set(recipient string, value *action.Value) error {
	glog.Infof(g.stage.Context(), "%v set %v = %v\n", g.info.Key(), recipient, value.String())
	g.info.params[recipient] = value.String()
	return nil
}

func (g *Golang) AddRelyOn(recipient string, dependency action.Action) error {
	glog.Infof(g.stage.Context(), "%v add rely on: %v\n", g.info.Key(), action.GenerateActionKey(dependency))
	g.relyOn[recipient] = reflect.ValueOf(newGolangAction(dependency, g.hasError))
	return nil
}

func (g *Golang) Get(name string) (*action.Value, error) {
	value, err := g.interpreter.Eval(name)
	glog.Infof(g.stage.Context(), "%v get %v's value: %v\n", g.info.Key(), name, action.NewValue(value).Interface())
	return action.NewValue(value), err
}

func (g *Golang) Do(ctx context.Context) error {
	glog.Infof(g.stage.Context(), "%v starts work\n", g.info.Key())
	_, err := g.interpreter.ExecuteWithContext(ctx, g.program)
	if err != nil {
		return err
	}
	return g.err
}

func (g *Golang) Compile() error {
	if g.program != nil {
		return nil
	}
	params := map[string]reflect.Value{
		"Value":                    reflect.ValueOf((*action.Value)(nil)),
		"NewValue":                 reflect.ValueOf(action.NewValue),
		"ConvertSliceToValueSlice": reflect.ValueOf(action.ConvertSliceToValueSlice),
		"ConvertValueSliceToSlice": reflect.ValueOf(action.ConvertValueSliceToSlice),
		"GenerateActionKey":        reflect.ValueOf(action.GenerateActionKey),
	}
	for k, v := range g.info.GetParams() {
		params[k] = reflect.ValueOf(v)
	}
	err := g.interpreter.Use(map[string]map[string]reflect.Value{
		"action/action": g.relyOn,
		"value/value":   params,
		"rule/rule": {
			"Error": reflect.ValueOf(g.SetError),
		},
	})
	if err != nil {
		return err
	}
	// TODO 临时添加部分依赖用于测试, 后续应当被删除
	_ = g.interpreter.Use(stdlib.Symbols)
	g.program, err = g.interpreter.Compile(g.info.OriginalContent())
	return err
}

func (g *Golang) SetError(i interface{}) {
	if i != nil {
		glog.Errorf(g.stage.Context(), "%v generate an error:%v\n", g.info.Key(), i)
		g.err = fmt.Errorf("%v", i)
		*g.hasError = true
	}
}

func (g *Golang) parseInfo() (err error) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("recover panic:", r)
			err = fmt.Errorf("malformed definition in the rule description section")
		}
	}()

	g.info.contentDescription = getInfoContent(g.info.OriginalContent())
	if len(g.info.contentDescription) == 0 {
		return ErrNotRuleName
	}
	_, err = g.interpreter.Eval(g.info.DescContent())
	if err != nil {
		return err
	}
	{
		value, _ := g.interpreter.Eval(RULE_TAG_NAME)
		if value.IsValid() {
			g.info.name = value.String()
		} else {
			return ErrNotRuleName
		}
	}
	{
		value, _ := g.interpreter.Eval(RULE_TAG_DESCRIPTION)
		if value.IsValid() {
			g.info.desc = value.String()
		}
	}
	{
		value, _ := g.interpreter.Eval(RULE_TAG_VERSION)
		if value.IsValid() {
			g.info.version = float32(value.Float())
		}
	}
	{
		value, _ := g.interpreter.Eval(RULE_TAG_RELY_ON)
		if value.IsValid() {
			relyOn, ok := value.Interface().(map[string]string)
			if ok {
				g.info.relyOn = relyOn
			} else {
				return ErrDependencyFormat
			}
		}
	}
	{
		value, _ := g.interpreter.Eval(RULE_TAG_PARAMS)
		if value.IsValid() {
			if value.Kind() != reflect.Map {
				return ErrDependencyFormat
			}
			iter := value.MapRange()
			for iter.Next() {
				g.info.params[iter.Key().String()] = iter.Value().String()
			}
		}
	}

	return nil
}

type GolangAction struct {
	a        action.Action
	hasError *bool
}

func newGolangAction(a action.Action, hasError *bool) *GolangAction {
	return &GolangAction{
		a:        a,
		hasError: hasError,
	}
}

func (g *GolangAction) Name() string {
	return g.a.Name()
}

func (g *GolangAction) Do(params ...*action.Value) (map[string]*action.Value, error) {
	if *g.hasError {
		return nil, nil
	}
	return g.a.Do(params...)
}

func (g *GolangAction) Version() float32 {
	return g.a.Version()
}

func (g *GolangAction) Description() string {
	return g.a.Description()
}
