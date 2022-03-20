package rule

import (
	"context"
	"fmt"
	"github.com/plugin-ops/pedestal/pedestal/action"
	"github.com/traefik/yaegi/interp"
	"reflect"
	"strings"
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

func (g *Golang) Set(recipient string, dependency action.Action) error {
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
	err := g.interpreter.Use(map[string]map[string]reflect.Value{
		"action/action": g.relyOn,
	})
	if err != nil {
		return err
	}
	g.program, err = g.interpreter.Compile(g.info.content)
	return err
}

func (g *Golang) parseInfo() (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("malformed definition in the rule description section")
		}
	}()

	infoContent := getInfoContent(g.info.OriginalContent())
	if len(infoContent) == 0 {
		return ErrNotRuleName
	}

	_, err = g.interpreter.Eval(infoContent)
	if err != nil {
		return err
	}
	{
		value, err := g.interpreter.Eval(SCRIPT_TAG_NAME)
		if err != nil {
			return err
		}
		if len(strings.TrimSpace(value.String())) == 0 {
			return ErrNotRuleName
		}
		g.info.name = value.String()
	}
	{
		value, err := g.interpreter.Eval(SCRIPT_TAG_DESCRIPTION)
		if err != nil {
			return err
		}
		g.info.desc = value.String()
	}
	{
		value, err := g.interpreter.Eval(SCRIPT_TAG_VERSION)
		if err != nil {
			return err
		}
		g.info.version = value.Float()
	}
	{
		value, err := g.interpreter.Eval(SCRIPT_TAG_RELY_ON)
		if err != nil {
			return err
		}
		g.info.relyOn = value.Interface().(map[string]string)
	}

	return nil
}
