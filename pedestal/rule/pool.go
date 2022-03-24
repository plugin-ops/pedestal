package rule

import (
	"errors"
	"sync"
)

var (
	pool      = map[string] /*rule name*/ Info{}
	poolMutex = &sync.RWMutex{}
)

// TODO 目前会覆盖
func SetRule(rule Rule) {
	poolMutex.Lock()
	pool[rule.Info().Name()] = rule.Info()
	poolMutex.Unlock()
}

func GetRule(name string) (Rule, error) {
	poolMutex.RLock()
	i := pool[name]
	poolMutex.RUnlock()
	switch i.RuleType() {
	case RuleTypeGo:
		return NewGolang(i.OriginalContent())
	default:
		return nil, errors.New("unknown type")
	}
}
