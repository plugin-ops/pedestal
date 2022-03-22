package boot

import "github.com/plugin-ops/pedestal/pedestal/server/http"

func startHttpServer() error {
	return http.StartHttpServer()
}
