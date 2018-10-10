package manage

import (
	"os"

	"git.xiagaogao.com/coffee/boot"
	"git.xiagaogao.com/coffee/boot/errors"
	"git.xiagaogao.com/coffee/httpx"
	"github.com/gin-gonic/gin"
	"github.com/spf13/pflag"
	"go.uber.org/zap"
)

var manage_endpoint = pflag.String("env_manage_endpoint", "0.0.0.0:7777", "管理服务地址")

func RegisterManager(router gin.IRouter, serviceInfo boot.ServiceInfo, errorService errors.Service, logger *zap.Logger) {
	router.Use(gin.BasicAuth(gin.Accounts{
		"root": "36NiH7*CsjXOm@SD",
	}))
	server := &manageServerImpl{
		serviceInfo:  serviceInfo,
		errorService: errorService,
		logger:       logger,
	}
	server.registerServiceRuntimeInfoEndpoint(router)
	server.registerHealthEndpoint(router)
	server.registerMetricsEndpoint(router)

}

func NewManageService(serviceInfo boot.ServiceInfo, errorService errors.Service, logger *zap.Logger) httpx.Service {
	manageEndpoint, ok := os.LookupEnv("ENV_MANAGE_ENDPOINT")
	if !ok {
		manageEndpoint = *manage_endpoint
	}
	service := httpx.NewService("manage", &httpx.Config{
		ServerAddr: manageEndpoint,
	}, logger)
	RegisterManager(service.GetGinEngine(), serviceInfo, errorService, logger)
	return service
}

type manageServerImpl struct {
	serviceInfo  boot.ServiceInfo
	errorService errors.Service
	logger       *zap.Logger
}
