package manage

import (
	"net"
	"os"

	"git.xiagaogao.com/coffee/boot"
	"git.xiagaogao.com/coffee/boot/bootutils"
	"git.xiagaogao.com/coffee/boot/errors"
	"git.xiagaogao.com/coffee/httpx"
	"go.uber.org/zap"
)

type Service interface {
	GetHttpService() httpx.Service
	GetEndpoint() string
	Start(onShutdown func())
}

type serviceImpl struct {
	serviceInfo  *boot.ServiceInfo
	errorService errors.Service
	logger       *zap.Logger
	httpService  httpx.Service
	endpoint     string
}

func (impl *serviceImpl) GetHttpService() httpx.Service {
	return impl.httpService
}

func (impl *serviceImpl) GetEndpoint() string {
	return impl.endpoint
}

func (impl *serviceImpl) Start(onShutdown func()) {
	impl.httpService.Start(onShutdown)
}

func (impl *serviceImpl) registerManager() {
	router := impl.httpService.GetGinEngine()
	// router.Use(gin.BasicAuth(gin.Accounts{
	// 	"root": "abc###123",
	// }))
	impl.registerServiceRuntimeInfoEndpoint(router)
	impl.registerHealthEndpoint(router)
	impl.registerMetricsEndpoint(router)
}

func NewManageService(serviceInfo *boot.ServiceInfo, errorService errors.Service, logger *zap.Logger) (Service, errors.Error) {
	service := &serviceImpl{
		errorService: errorService,
		logger:       logger,
		serviceInfo:  serviceInfo,
	}
	manageEndpoint, ok := os.LookupEnv("ENV_MANAGE_ENDPOINT")
	if !ok {
		manageEndpoint = *manage_endpoint
	}
	manageEndpoint, err := bootutils.WarpServerAddr(manageEndpoint, errorService)
	if err != nil {
		return nil, err
	}
	lis, err1 := net.Listen("tcp4", manageEndpoint)
	if err1 != nil {
		return nil, errorService.SystemError("启动ManageServer失败，请设置：ENV_MANAGE_ENDPOINT,或者--env_manage_endpoint", zap.Error(err1))
	}
	manageEndpoint = lis.Addr().String()
	lis.Close()
	service.endpoint = manageEndpoint
	logger.Debug("设置管理Endpoint", zap.String("endpoint", manageEndpoint))
	service.httpService = httpx.NewService("manage", &httpx.Config{
		ServerAddr: manageEndpoint,
	}, logger)
	service.registerManager()
	return service, nil
}
