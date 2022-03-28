package action

import (
	"fmt"
)

type Action interface {
	Name() string
	Do(params ...*Value) (map[string]*Value, error)
	Version() float32
	Description() string
}

func GenerateActionKey(a Action) string {
	return fmt.Sprintf("%v@%v", a.Name(), a.Version())
}
