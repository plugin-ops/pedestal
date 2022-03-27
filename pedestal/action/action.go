package action

type Action interface {
	Name() string
	Do(params ...*Value) (map[string]*Value, error)
	Version() float32
	Description() string
}
