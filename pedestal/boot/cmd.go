package boot

import (
	"errors"
	"io/ioutil"

	"github.com/plugin-ops/pedestal/pedestal/config"
	"github.com/plugin-ops/pedestal/pedestal/execute"
	"github.com/plugin-ops/pedestal/pedestal/plugin"
	"github.com/plugin-ops/pedestal/pedestal/rule"
	"github.com/plugin-ops/pedestal/pedestal/util"
)

func runCmd() error {
	if err := checkConfig(); err != nil {
		return err
	}
	return runRule()
}

func checkConfig() error {
	if !util.FileExist(config.ScriptPath) {
		return errors.New("unknown script address")
	}
	return nil
}

func runRule() error {
	exec, err := execute.NewBuiltInExecutor()
	if err != nil {
		return err
	}

	f, err := ioutil.ReadFile(config.ScriptPath)
	if err != nil {
		return err
	}
	r, err := rule.NewGolang(string(f))
	if err != nil {
		return err
	}

	err = plugin.ReLoadPluginWithDir(config.PluginDir)
	if err != nil {
		return err
	}

	_, err = exec.Execute(r, nil)
	return err
}
