package action

import (
	"runtime"
	"sync"
)

type actionSet struct {
	set   map[string] /*action name*/ Action
	mutex *sync.RWMutex
}

func NewActionSet() *actionSet {
	return &actionSet{
		set:   map[string]Action{},
		mutex: &sync.RWMutex{},
	}
}

var as = NewActionSet()

func RegisterAction(a Action) {
	as.mutex.Lock()
	as.set[a.Name()] = a
	as.mutex.Unlock()
}

func CleanAllAction() {
	as.mutex.Lock()
	as.set = map[string]Action{}
	as.mutex.Unlock()
	runtime.GC()
}

func RemoveAction(name string) {
	as.mutex.Lock()
	delete(as.set, name)
	as.mutex.Unlock()
	runtime.GC()
}

func GetAction(name string) Action {
	as.mutex.RLock()
	defer as.mutex.RUnlock()
	return as.set[name]
}

// CheckActionExist return not exist list
func CheckActionExist(names ...string) []string {
	notExist := []string{}
	as.mutex.RLock()
	for _, name := range names {
		if _, ok := as.set[name]; !ok {
			notExist = append(notExist, name)
		}
	}
	as.mutex.RUnlock()
	if len(notExist) == 0 {
		return nil
	}
	return notExist
}
