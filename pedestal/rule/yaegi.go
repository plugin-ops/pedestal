package rule

import (
	"context"
	"fmt"
	"reflect"

	"github.com/plugin-ops/pedestal/pedestal/action"

	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
)

type Golang struct {
	info        *info
	program     *interp.Program
	interpreter *interp.Interpreter
	relyOn      map[string]reflect.Value
}

func NewGolang(content string) (*Golang, error) {
	g := &Golang{
		info: &info{
			content: content,
			relyOn:  map[string]string{},
			params:  map[string]interface{}{},
		},
		interpreter: interp.New(interp.Options{}),
		relyOn:      map[string]reflect.Value{},
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
	g.info.params[recipient] = value.Interface()
	return nil
}

func (g *Golang) AddRelyOn(recipient string, dependency action.Action) error {
	g.relyOn[recipient] = reflect.ValueOf(action.ConvertActionToFunc(dependency))
	return nil
}

func (g *Golang) Get(name string) (*action.Value, error) {
	value, err := g.interpreter.Eval(name)
	return action.NewValue(value), err
}

func (g *Golang) Do(ctx context.Context) error {
	_, err := g.interpreter.ExecuteWithContext(ctx, g.program)
	return err
}

func (g *Golang) Compile() error {
	if g.program != nil {
		return nil
	}
	params := map[string]reflect.Value{}
	for k, v := range g.info.GetParams() {
		params[k] = reflect.ValueOf(v)
	}
	err := g.interpreter.Use(map[string]map[string]reflect.Value{
		"action/action": g.relyOn,
		"value/value":   params,
	})
	if err != nil {
		return err
	}
	// TODO 临时添加部分依赖用于测试, 后续应当被删除
	_ = g.interpreter.Use(stdlib.Symbols)
	g.program, err = g.interpreter.Compile(g.info.BodyContent())
	return err
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
	g.info.contentBody = getBodyContent(g.info.OriginalContent())
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
			g.info.version = value.Float()
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
			relyOn, ok := value.Interface().(map[string]interface{})
			if ok {
				g.info.params = relyOn
			} else {
				return ErrDependencyFormat
			}
		}
	}

	return nil
}
