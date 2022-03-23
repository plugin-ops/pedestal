package boot

import (
	"fmt"
	"github.com/plugin-ops/pedestal/pedestal/config"
	"github.com/plugin-ops/pedestal/pedestal/execute"
	"github.com/plugin-ops/pedestal/pedestal/plugin"
	"os"

	"github.com/spf13/cobra"
)

var version string

func init() {
	config.Version = version
}

func Run() {
	var rootCmd = &cobra.Command{
		Use:   "plugin-ops",
		Short: "PLUGIN_OPS",
		Long:  "Plugin-Ops\n\nVersion:\n  " + version,
		Run: func(cmd *cobra.Command, args []string) {
			if err := run(cmd, args); nil != err {
				os.Exit(1)
			}
		},
	}

	registerFlag(rootCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println("plugin-ops abnormal termination:", err)
		os.Exit(1)
	}
}

func registerFlag(c *cobra.Command) {
	c.Flags().StringVarP(&config.PluginDir, "plugin-dir", "p", "./plugin", "Plugin storage directory, there should be a src directory in this directory, and all plugins are placed in the src directory")
	c.Flags().StringVarP(&config.RulePath, "rule-path", "r", "", "Script file address, the plugin will execute this rule")
	c.Flags().BoolVarP(&config.EnableHttpServer, "enable-http-server", "s", false, "enable http server")
}

func run(cmd *cobra.Command, _ []string) (err error) {
	defer clean()

	err = plugin.ReLoadPluginWithDir(config.PluginDir)
	if err != nil {
		return err
	}

	err = execute.InitExecute()
	if err != nil {
		return err
	}

	errChan := make(chan error)
	if config.EnableHttpServer {
		go func() {
			errChan <- startHttpServer()
		}()
	}

	return <-errChan
}
