package serviceboot

import (
	"github.com/coffeehc/microserviceboot/base"
	"github.com/coffeehc/web"
	"github.com/coffeehc/web/pprof"
	"github.com/prometheus/client_golang/prometheus"
)

func NewHttpServer(configPath string, config *web.HttpServerConfig, service base.Service) (web.HttpServer, base.Error) {
	httpServer := web.NewHttpServer(config)
	if service.Init != nil {
		err := service.Init(configPath, httpServer)
		if err != nil {
			return nil, base.NewErrorWrapper(err)
		}
	}
	pprof.RegeditPprof(httpServer)
	health := newHealth(service.GetServiceInfo())
	err := httpServer.Register("/health", web.GET, health.health)
	if err != nil {
		return nil, base.NewErrorWrapper(err)
	}
	err = httpServer.RegisterHttpHandler("/metrics", web.GET, prometheus.Handler())
	if err != nil {
		return nil, base.NewErrorWrapper(err)
	}
	return httpServer, nil
}
