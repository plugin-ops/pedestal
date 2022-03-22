package http

import (
	"errors"
	"fmt"

	"github.com/plugin-ops/pedestal/pedestal/config"
	"github.com/plugin-ops/pedestal/pedestal/server/http/controller"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

func StartHttpServer() error {
	s := g.Server()

	registerRouter(s)

	s.SetAddr(fmt.Sprintf("%v:%v", config.HttpIP, config.HttpPort))

	err := s.Start()
	if err != nil {
		return err
	}
	g.Wait()
	return errors.New("http server down")
}

func registerRouter(s *ghttp.Server) {
	s.Group("/v1", func(group *ghttp.RouterGroup) {
		group.POST("plugins/all/reload", controller.V1Api.ReloadAllPlugins)
		group.POST("plugins/run", controller.V1Api.RunRule)
	})
}
