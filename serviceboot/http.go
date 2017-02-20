package serviceboot

import (
	"github.com/coffeehc/httpx"
	"github.com/coffeehc/httpx/pprof"
	"github.com/coffeehc/microserviceboot/base"
	"github.com/prometheus/client_golang/prometheus"
)

//NewHTTPServer 创建 http server
func NewHTTPServer(config *httpx.Config, serviceInfo base.ServiceInfo) (httpx.Server, base.Error) {
	httpServer := httpx.NewServer(config)
	pprof.RegeditPprof(httpServer)
	health := newHealth(serviceInfo)
	err := httpServer.Register("/health", httpx.GET, health.health)
	if err != nil {
		return nil, base.NewErrorWrapper("http server",0, err)
	}
	err = httpServer.RegisterHandler("/metrics", httpx.GET, prometheus.Handler())
	if err != nil {
		return nil, base.NewErrorWrapper("http server",0, err)
	}
	return httpServer, nil
}
