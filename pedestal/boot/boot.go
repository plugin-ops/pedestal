package boot

import (
	"fmt"
	"github.com/plugin-ops/pedestal/pedestal/rule"
	"os"

	"github.com/plugin-ops/pedestal/pedestal/app/api/http"
	"github.com/plugin-ops/pedestal/pedestal/config"
	"github.com/plugin-ops/pedestal/pedestal/execute"
	"github.com/plugin-ops/pedestal/pedestal/log"
	"github.com/plugin-ops/pedestal/pedestal/plugin"

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
				fmt.Println("pedestal running failed: ", err)
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
	c.Flags().StringVarP(&config.PluginDir, "plugin-dir", "p", "./plugin", "Plugin storage directory")
	c.Flags().StringVarP(&config.RuleDir, "rule-dir", "r", "", "Rule storage address")
	c.Flags().StringVarP(&config.LogDir, "log-dir", "l", "", "Log storage address")
	c.Flags().BoolVarP(&config.EnableHttpServer, "enable-http-server", "s", false, "enable http server")
}

func run(cmd *cobra.Command, _ []string) (err error) {
	stage := log.NewStage().Enter("StartPedestal")

	defer clean()

	err = initModule(stage)
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

func initModule(stage *log.Stage) (err error) {
	err = log.InitLoggerOnPedestal()
	if err != nil {
		return err
	}

	err = plugin.ReLoadPluginWithDir(stage, config.PluginDir)
	if err != nil {
		return err
	}

	err = rule.ReLoadRuleWithDir(stage, config.RuleDir)
	if err != nil {
		return err
	}

	err = execute.InitExecute()
	if err != nil {
		return err
	}

	return nil
}

func clean() {
	stage := log.NewStage().Enter("CleaningAfterThePedestalIsStopped")
	plugin.UninstallAllPlugin(stage)
}

func startHttpServer() error {
	return http.StartHttpServer()
}
