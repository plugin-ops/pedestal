package action

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/plugin-ops/pedestal/pedestal/log"

	"github.com/gogf/gf/v2/os/glog"
)

var (
	pool      = map[string] /*action name*/ map[float32] /*action version*/ Action{}
	poolMutex = &sync.RWMutex{}
)

func ListActionKey(stage *log.Stage) []string {
	stage = stage.Go("ListActionKey")
	names := []string{}
	poolMutex.RLock()
	for _, s := range pool {
		for _, action := range s {
			names = append(names, GenerateActionKey(action))
		}
	}
	poolMutex.RUnlock()
	glog.Infof(stage.Context(), "get action list success, current Action list: %v\n", strings.Join(names, ";"))
	return names
}

func RegisterAction(stage *log.Stage, a Action) {
	stage = stage.Go("RegisterAction")
	glog.Infof(stage.Context(), "registry action: %v\n", GenerateActionKey(a))
	poolMutex.Lock()
	if _, ok := pool[a.Name()]; !ok {
		pool[a.Name()] = map[float32]Action{}
	}
	if _, ok := pool[a.Name()][a.Version()]; ok {
		glog.Infof(stage.Context(), "action %v already exists，will be overwritten\n", GenerateActionKey(a))
	}
	pool[a.Name()][a.Version()] = a
	poolMutex.Unlock()
}

func CleanAllAction(stage *log.Stage) {
	stage = stage.Go("CleanAllAction")
	glog.Warning(stage.Context(), "clean all action...")
	poolMutex.Lock()
	pool = map[string]map[float32]Action{}
	poolMutex.Unlock()
	runtime.GC()
}

// RemoveAction when the incoming version is -1, all versions of actions will be cleared
func RemoveAction(stage *log.Stage, name string, version float32) {
	stage = stage.Go("RemoveAction")
	poolMutex.Lock()
	if version == -1 {
		glog.Warningf(stage.Context(), "remove all actions named %v\n", name)
		pool[name] = map[float32]Action{}
	} else {
		glog.Warningf(stage.Context(), "remove action %v@%v\n", name, version)
		set, ok := pool[name]
		if ok {
			delete(set, version)
		}
	}
	poolMutex.Unlock()
	runtime.GC()
}

// GetAction returns the latest action when the incoming version is -1
func GetAction(stage *log.Stage, name string, version float32) (a Action, exist bool) {
	stage = stage.Go("GetAction")
	poolMutex.RLock()
	defer poolMutex.RUnlock()

	if _, ok := pool[name]; !ok {
		glog.Warningf(stage.Context(), "get action %v failed, because not exist", name)
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
		glog.Warningf(stage.Context(), "get action %v%v failed, because not exist", name, version)
		return a, false
	}
	glog.Infof(stage.Context(), "get action %v", GenerateActionKey(a))
	return a, true
}

// CheckActionExist return not exist list
func CheckActionExist(key ...string) []string {
	notExist := []string{}

	temp := map[string][]float32{}
	for _, name := range key {
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
