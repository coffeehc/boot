package serviceboot

import (
	"github.com/coffeehc/microserviceboot/base"
	"github.com/coffeehc/web"
	"github.com/coffeehc/web/pprof"
	"github.com/prometheus/client_golang/prometheus"
)

func NewHttpServer(config *web.HttpServerConfig, serviceInfo base.ServiceInfo) (web.HttpServer, base.Error) {
	httpServer := web.NewHttpServer(config)
	pprof.RegeditPprof(httpServer)
	health := newHealth(serviceInfo)
	err := httpServer.Register("/health", web.GET, health.health)
	if err != nil {
		return nil, base.NewErrorWrapper("httpserver", err)
	}
	err = httpServer.RegisterHttpHandler("/metrics", web.GET, prometheus.Handler())
	if err != nil {
		return nil, base.NewErrorWrapper("httpserver", err)
	}
	return httpServer, nil
}
