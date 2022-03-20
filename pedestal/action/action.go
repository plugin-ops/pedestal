package action

type Action interface {
	Name() string
	Do(params ...*Value) ([]*Value, error)
	Version() float32
	Description() string
}

func ConvertActionToFunc(a Action) func(...interface{}) ([]interface{}, error) {
	return func(i ...interface{}) ([]interface{}, error) {
		resp, err := a.Do(ConvertSliceToValueSlice(i)...)
		return ConvertValueSliceToSlice(resp...), err
	}
}
