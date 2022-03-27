package action

import (
	"fmt"
	"reflect"
	"strconv"
)

type Value struct {
	v reflect.Value
	k reflect.Kind
}

func NewValue(v interface{}) *Value {
	if _, ok := v.(*Value); ok {
		return v.(*Value)
	}
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

func (v *Value) Float64() float64 {
	f, _ := strconv.ParseFloat(v.String(), 10)
	return f
}

func (v *Value) Bool() bool {
	if v.Kind() == reflect.Bool {
		return v.v.Bool()
	}
	return false
}

func (v *Value) Value() reflect.Value {
	return v.v
}

func (v *Value) Kind() reflect.Kind {
	return v.k
}

func (v *Value) IsNil() bool {
	if !v.v.IsValid() {
		return true
	}
	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Ptr, reflect.UnsafePointer, reflect.Interface, reflect.Slice:
		return v.v.IsNil()
	default:
		return false
	}
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
