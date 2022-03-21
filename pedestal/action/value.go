package action

import (
	"fmt"
	"reflect"
)

type Value struct {
	v reflect.Value
	k reflect.Kind
}

func NewValue(v interface{}) *Value {
	value := &Value{
		v: reflect.ValueOf(v),
	}
	value.k = value.v.Kind()
	return value
}

func (v *Value) Interface() interface{} {
	return v.v.Interface()
}

func (v *Value) String() string {
	return fmt.Sprintf("%v", v.v.Interface())
}

func (v *Value) Value() reflect.Value {
	return v.v
}

func (v *Value) Kind() reflect.Kind {
	return v.k
}

func (v *Value) IsNil() bool {
	// TODO 应该根据kind判断
	return !v.v.IsValid()
}

func ConvertSliceToValueSlice(i ...interface{}) []*Value {
	vs := make([]*Value, len(i))
	for index, value := range i {
		vs[index] = NewValue(value)
	}
	return vs
}

func ConvertValueSliceToSlice(v ...*Value) []interface{} {
	is := make([]interface{}, len(v))
	for i, value := range v {
		is[i] = value.Interface()
	}
	return is
}
