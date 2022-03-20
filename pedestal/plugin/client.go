package plugin

import (
	"github.com/plugin-ops/pedestal/pedestal/action"

	"github.com/hashicorp/go-plugin"
)

func ServePlugin(a action.Action) {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: handshakeConfig,
		Plugins: plugin.PluginSet{
			PluginName: &actionImpl{
				Srv: &driverGRPCServer{impl: a},
			},
		},
		GRPCServer: plugin.DefaultGRPCServer,
	})
}
