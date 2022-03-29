package v1

import (
	"path"

	"github.com/plugin-ops/pedestal/pedestal/config"
	"github.com/plugin-ops/pedestal/pedestal/log"
	"github.com/plugin-ops/pedestal/pedestal/plugin"
)

func ReloadAllPlugins() error {
	stage := log.NewStage().Enter("ReloadAllPlugins")

	return plugin.ReLoadPluginWithDir(stage, config.PluginDir)
}

type RemovePluginReqV1 struct {
	ActionName    string  `json:"action_name"`
	ActionVersion float32 `json:"action_version"`
}

func RemovePlugin(req *RemovePluginReqV1) {
	stage := log.NewStage().Enter("RemovePlugin")

	plugin.UninstallPlugin(stage, req.ActionName, req.ActionVersion)
}

func AddPlugin(fileName string) error {
	stage := log.NewStage().Enter("AddPlugin")

	err := plugin.LoadPluginWithLocal(stage, path.Join(config.PluginDir, fileName))
	if err != nil {
		return err
	}
	return plugin.CleanUselessPluginFile(stage)
}
