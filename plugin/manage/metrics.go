package manage

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func (impl *pluginImpl) registerMetricsEndpoint(router gin.IRouter) {
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))
	router.GET("/ping", impl.ping())
	registerPprof(router)
}