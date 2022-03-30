package http

import (
	"errors"
	"fmt"
	v1 "github.com/plugin-ops/pedestal/pedestal/app/api/http/v1"
	"github.com/plugin-ops/pedestal/pedestal/config"
	"path"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

func StartHttpServer() error {
	s := g.Server()

	registerRouter(s)

	err := setConfig(s)
	if err != nil {
		return err
	}

	err = s.Start()
	if err != nil {
		return err
	}
	g.Wait()
	return errors.New("http server down")
}

func registerRouter(s *ghttp.Server) {
	s.Group("/v1", func(group *ghttp.RouterGroup) {
		group.POST("plugin/all/reload", v1.ReloadAllPlugins)
		group.POST("plugin/remove", v1.RemovePlugin)
		group.POST("plugin/add", v1.AddPlugin)

		group.GET("action/list", v1.ListAction)

		group.POST("rule/all/reload", v1.ReloadAllRule)
		group.POST("rule/run", v1.RunRule)
		group.POST("rule/add", v1.AddRule)
	})
}

func setConfig(s *ghttp.Server) (err error) {

	s.SetAddr(fmt.Sprintf("%v:%v", config.HttpIP, config.HttpPort))
	s.SetLogStdout(true)
	s.SetAccessLogEnabled(true)
	s.SetErrorLogEnabled(true)

	err = s.SetLogPath(path.Join(config.LogDir, "api.log"))
	if err != nil {
		return err
	}

	return nil
}
