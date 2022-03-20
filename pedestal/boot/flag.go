package boot

import (
	"fmt"
	"os"

	"github.com/plugin-ops/pedestal/pedestal/config"

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

	rootCmd.Flags().StringVarP(&config.PluginDir, "plugin-dir", "p", "./plugin", "Plugin storage directory, there should be a src directory in this directory, and all plugins are placed in the src directory")
	rootCmd.Flags().StringVarP(&config.RulePath, "rule-path", "r", "", "Script file address, the plugin will execute this rule")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println("plugin-ops abnormal termination:", err)
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, _ []string) error {
	defer clean()
	return runCmd()
}
