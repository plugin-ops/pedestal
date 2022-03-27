package rule

import (
	"os"
	"path"
	"sync"

	"github.com/plugin-ops/pedestal/pedestal/config"

	"github.com/sirupsen/logrus"
)

var (
	pool      = map[string] /*rule name*/ map[float32] /*rule version*/ Info{}
	poolMutex = &sync.RWMutex{}
)

func RegistryRuleAndStoreToLocal(entry *logrus.Entry, info Info) error {
	poolMutex.Lock()
	defer poolMutex.Unlock()

	registryRule(entry, info)

	entry.Infof("store rule %v to %v\n", info.Key(), config.RuleDir)
	f, err := os.Create(path.Join(config.RuleDir, info.Key()))
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(info.OriginalContent())
	return err
}

// GetRule returns the latest rule when the incoming version is -1
func GetRule(entry *logrus.Entry, name string, version float32) (rule Rule, exist bool, err error) {
	poolMutex.RLock()
	defer poolMutex.RUnlock()

	if _, ok := pool[name]; !ok {
		entry.Warnf("get rule %v failed, because not exist\n", name)
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
		entry.Warnf("get rule %v%v failed, because not exist\n", name, version)
		return
	}
	entry.Infof("get rule %v", info.Key())

	switch info.RuleType() {
	case RuleTypeGo:
		rule, err = NewGolang(info.OriginalContent())
		return rule, true, err
	default:
		return nil, false, ErrorUnknownRuleType
	}

}

func RegistryRule(entry *logrus.Entry, info Info) {
	poolMutex.Lock()
	registryRule(entry, info)
	poolMutex.Unlock()
}

func registryRule(entry *logrus.Entry, info Info) {
	entry.Infof("registry rule: %v\n", info.Key())
	if _, ok := pool[info.Name()]; !ok {
		pool[info.Name()] = map[float32]Info{}
	}
	if _, ok := pool[info.Name()][info.Version()]; ok {
		entry.Infof("rule %v already exists，will be overwritten\n", info.Key())
	}
	pool[info.Name()][info.Version()] = info
}
