package action

import "github.com/gogf/gf/v2/container/gvar"

type Value struct {
	*gvar.Var
}

func NewValue(i interface{}) *Value {
	return &Value{gvar.New(i, true)}
}

func ConvertSliceToValueSlice(i ...interface{}) []*Value {
	vs := make([]*Value, len(i))
	for index, value := range i {
		vs[index] = NewValue(value)
	}
	return vs
}

func ConvertValueSliceToSlice(v map[string]*Value) map[string]interface{} {
	is := map[string]interface{}{}
	for i, value := range v {
		is[i] = value.Interface()
	}
	return is
}
