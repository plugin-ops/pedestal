package rule

import (
	"os"
	"path"
	"sync"

	"github.com/plugin-ops/pedestal/pedestal/config"
	"github.com/plugin-ops/pedestal/pedestal/log"

	"github.com/gogf/gf/v2/os/glog"
)

var (
	pool      = map[string] /*rule name*/ map[float32] /*rule version*/ Info{}
	poolMutex = &sync.RWMutex{}
)

func RegistryRuleAndStoreToLocal(stage *log.Stage, info Info) error {
	poolMutex.Lock()
	defer poolMutex.Unlock()

	registryRule(stage, info)

	glog.Infof(stage.Context(), "store rule %v to %v\n", info.Key(), config.RuleDir)
	f, err := os.Create(path.Join(config.RuleDir, info.Key()))
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(info.OriginalContent())
	return err
}

// GetRule returns the latest rule when the incoming version is -1
func GetRule(stage *log.Stage, name string, version float32) (rule Rule, exist bool, err error) {
	poolMutex.RLock()
	defer poolMutex.RUnlock()

	if _, ok := pool[name]; !ok {
		glog.Warningf(stage.Context(), "get rule %v failed, because not exist\n", name)
		return
	}
	var info Info
	if version == -1 {
		var max float32 = -1
		for _, i := range pool[name] {
			if i.Version() > max {
				max = i.Version()
				info = i
			}
		}
	} else {
		for v, i := range pool[name] {
			if version == v {
				info = i
				break
			}
		}
	}

	if info == nil {
		glog.Warningf(stage.Context(), "get rule %v%v failed, because not exist\n", name, version)
		return
	}
	glog.Infof(stage.Context(), "get rule %v", info.Key())

	switch info.RuleType() {
	case RuleTypeGo:
		rule, err = NewGolang(stage, info.OriginalContent())
		return rule, true, err
	default:
		return nil, false, ErrorUnknownRuleType
	}

}

func RegistryRule(stage *log.Stage, info Info) {
	poolMutex.Lock()
	registryRule(stage, info)
	poolMutex.Unlock()
}

func registryRule(stage *log.Stage, info Info) {
	glog.Infof(stage.Context(), "registry rule: %v\n", info.Key())
	if _, ok := pool[info.Name()]; !ok {
		pool[info.Name()] = map[float32]Info{}
	}
	if _, ok := pool[info.Name()][info.Version()]; ok {
		glog.Infof(stage.Context(), "rule %v already exists，will be overwritten\n", info.Key())
	}
	pool[info.Name()][info.Version()] = info
}
