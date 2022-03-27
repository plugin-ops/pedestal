package action

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
)

var (
	pool      = map[string] /*action name*/ map[float32] /*action version*/ Action{}
	poolMutex = &sync.RWMutex{}
)

func ListActionName(entry *logrus.Entry) []string {
	names := []string{}
	poolMutex.RLock()
	for _, s := range pool {
		for _, action := range s {
			names = append(names, fmt.Sprintf("%v@%v", action.Name(), action.Version()))
		}
	}
	poolMutex.RUnlock()
	entry.Tracef("get action list success, current Action list: %v\n", strings.Join(names, ";"))
	return names
}

func RegisterAction(entry *logrus.Entry, a Action) {
	entry.Infof("registry action: %v@%v\n", a.Name(), a.Version())
	poolMutex.Lock()
	if _, ok := pool[a.Name()]; !ok {
		pool[a.Name()] = map[float32]Action{}
	}
	if _, ok := pool[a.Name()][a.Version()]; ok {
		entry.Infof("action %v@%v already exists，will be overwritten\n", a.Name(), a.Version())
	}
	pool[a.Name()][a.Version()] = a
	poolMutex.Unlock()
}

func CleanAllAction(entry *logrus.Entry) {
	entry.Warnln("clean all action...")
	poolMutex.Lock()
	pool = map[string]map[float32]Action{}
	poolMutex.Unlock()
	runtime.GC()
}

// RemoveAction when the incoming version is -1, all versions of actions will be cleared
func RemoveAction(entry *logrus.Entry, name string, version float32) {
	poolMutex.Lock()
	if version == -1 {
		entry.Warnf("remove all actions named %v\n", name)
		pool[name] = map[float32]Action{}
	} else {
		entry.Warnf("remove action %v@%v\n", name, version)
		set, ok := pool[name]
		if ok {
			delete(set, version)
		}
	}
	poolMutex.Unlock()
	runtime.GC()
}

// GetAction returns the latest action when the incoming version is -1
func GetAction(entry *logrus.Entry, name string, version float32) (a Action, exist bool) {
	poolMutex.RLock()
	defer poolMutex.RUnlock()

	if _, ok := pool[name]; !ok {
		entry.Warnf("get action %v failed, because not exist\n", name)
		return
	}

	if version == -1 {
		var max float32 = -1
		for _, action := range pool[name] {
			if action.Version() > max {
				max = action.Version()
				a = action
			}
		}
	} else {
		for v, action := range pool[name] {
			if version == v {
				a = action
				break
			}
		}
	}

	if a == nil {
		entry.Warnf("get action %v@%v failed, because not exist\n", name, version)
		return a, false
	}
	entry.Infof("get action %v@%v", a.Name(), a.Version())
	return a, true
}

// CheckActionExist return not exist list
func CheckActionExist(names ...string) []string {
	notExist := []string{}

	temp := map[string][]float32{}
	for _, name := range names {
		nv := strings.Split(name, "@")
		if temp[nv[0]] == nil {
			temp[nv[0]] = []float32{}
		}
		if len(nv) == 1 {
			temp[nv[0]] = append(temp[nv[0]], -1)
		} else if len(nv) == 2 {
			f, err := strconv.ParseFloat(nv[1], 10)
			if err != nil {
				notExist = append(notExist, name)
			} else {
				temp[nv[0]] = append(temp[nv[0]], float32(f))
			}
		} else {
			notExist = append(notExist, name)
		}
	}

	poolMutex.RLock()

	for n, f := range temp {
		if len(pool[n]) == 0 {
			for _, ff := range f {
				if ff == -1 {
					notExist = append(notExist, n)
				} else {
					notExist = append(notExist, fmt.Sprintf("%v@%v", n, ff))
				}
			}
			continue
		}

		for _, ff := range f {
			if ff == -1 {
				continue
			}
			if _, ok := pool[n][ff]; !ok {
				notExist = append(notExist, fmt.Sprintf("%v@%v", n, ff))
			}
		}
	}

	poolMutex.RUnlock()
	if len(notExist) == 0 {
		return nil
	}
	return notExist
}
