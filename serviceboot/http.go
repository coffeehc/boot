package serviceboot

import (
	"github.com/coffeehc/microserviceboot/base"
	"github.com/coffeehc/web"
)

func NewHttpServer(configPath string, config *web.HttpServerConfig, service base.Service) (web.HttpServer, base.Error) {
	httpServer := web.NewHttpServer(config)
	if service.Init != nil {
		err := service.Init(configPath, httpServer)
		if err != nil {
			return nil, base.NewErrorWrapper(err)
		}
	}
	return httpServer, nil
}
