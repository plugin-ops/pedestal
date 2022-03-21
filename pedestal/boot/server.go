package boot

import "github.com/plugin-ops/pedestal/pedestal/serve/http"

func startHttpServer() error {
	return http.StartHttpServer()
}
