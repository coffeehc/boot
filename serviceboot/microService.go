package serviceboot

import (
	"fmt"
	"net/http"

	"flag"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/coffeehc/logger"
	"github.com/coffeehc/microserviceboot/base"
	"github.com/coffeehc/web"
	"github.com/coffeehc/web/pprof"
)

type MicroService struct {
	config  *ServiceConfig
	server  *web.Server
	service base.Service
}

func newMicroService(service base.Service) (*MicroService, base.Error) {
	if flag.Lookup("help") != nil {
		flag.PrintDefaults()
		os.Exit(0)
	}
	serviceInfo := service.GetServiceInfo()
	if serviceInfo == nil {
		return nil, base.NewError(base.ERROR_CODE_BASE_INIT_ERROR, "没有指定 ServiceInfo")
	}
	logger.Info("ServiceName: %s", serviceInfo.GetServiceName())
	logger.Info("Version: %s", serviceInfo.GetVersion())
	logger.Info("Descriptor: %s", serviceInfo.GetDescriptor())
	return &MicroService{
		service: service,
	}, nil
}

func (microService *MicroService) init() base.Error {
	logger.Info("Service startting")
	serviceConfig := new(ServiceConfig)
	*configPath = base.GetDefaultConfigPath(*configPath)
	err := base.LoadConfig(*configPath, serviceConfig)
	if err != nil {
		logger.Warn("加载服务器配置[%s]失败,%s", *configPath, err)
	}
	microService.config = serviceConfig
	webConfig := serviceConfig.GetWebServerConfig()
	microService.server = web.NewServer(webConfig)
	if microService.service.Init != nil {
		err := microService.service.Init(*configPath, microService.server)
		if err != nil {
			return err
		}
	}
	serviceInfo := microService.service.GetServiceInfo()
	err = microService.registerEndpoints()
	if err != nil {
		return err
	}
	pprof.RegeditPprof(microService.server)
	if base.IsDevModule() {
		debugConfig := serviceConfig.getDebugConfig()
		logger.Debug("open dev module")
		apiDefineRequestHandler := buildApiDefineRequestHandler(serviceInfo)
		if apiDefineRequestHandler != nil {
			microService.server.Register(fmt.Sprintf("/apidefine/%s.api", serviceInfo.GetServiceName()), web.GET, apiDefineRequestHandler)
		}
		if debugConfig.IsEnableAccessInfo() {
			microService.server.AddFirstFilter("/*", web.SimpleAccessLogFilter)
		}
	}

	return nil
}

func (microService *MicroService) start() base.Error {
	startTime := time.Now()
	serviceInfo := microService.service.GetServiceInfo()
	//TODO 拦截异常返回
	err := microService.server.Start()
	if err != nil {
		return base.NewError(base.ERROR_CODE_BASE_INIT_ERROR, err.Error())
	}
	serviceDiscoveryRegister := microService.service.GetServiceDiscoveryRegister()
	if !microService.config.DisableServiceRegister && serviceDiscoveryRegister != nil {
		_, port, _ := net.SplitHostPort(microService.server.GetServerAddress())
		p, _ := strconv.Atoi(port)
		registerError := serviceDiscoveryRegister.RegService(serviceInfo, microService.service.GetEndPoints(), p)
		if registerError != nil {
			logger.Info("注册服务[%s]失败,%s", microService.service.GetServiceInfo().GetServiceName(), registerError.Error())
			return registerError
		}
		logger.Info("注册服务[%s]成功", microService.service.GetServiceInfo().GetServiceName())
	}
	logger.Info("Service started [%s]", time.Since(startTime))
	return nil
}

func buildApiDefineRequestHandler(serviceInfo base.ServiceInfo) web.RequestHandler {
	return func(request *http.Request, pathFragments map[string]string, reply web.Reply) {
		reply.With(serviceInfo.GetApiDefine()).As(web.Transport_Text)
	}
}

func (this *MicroService) registerEndpoint(endPoint base.EndPoint) base.Error {
	metadata := endPoint.Metadata
	logger.Debug("register endpoint [%s] %s %s", metadata.Method, metadata.Path, metadata.Description)
	err := this.server.Register(metadata.Path, metadata.Method, endPoint.HandlerFunc)
	if err != nil {
		return base.NewError(base.ERROR_CODE_BASE_INIT_ERROR, err.Error())
	}
	return nil
}

func (this *MicroService) registerEndpoints() base.Error {
	endPoints := this.service.GetEndPoints()
	if len(endPoints) == 0 {
		logger.Warn("not regedit any endpoint")
		return nil
	}
	for _, endPoint := range endPoints {
		err := this.registerEndpoint(endPoint)
		if err != nil {
			return err
		}
	}
	return nil
}
