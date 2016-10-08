package serviceboot

import (
	"fmt"

	"flag"
	"net"
	"os"
	"strconv"

	"github.com/coffeehc/logger"
	"github.com/coffeehc/microserviceboot/base"
	"github.com/coffeehc/web"
	"github.com/coffeehc/web/pprof"
)

type MicroService struct {
	config     *ServiceConfig
	httpServer web.HttpServer
	service    base.Service
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
	logger.Debug("serviceboot Config is %#v", serviceConfig)
	microService.config = serviceConfig
	webConfig := serviceConfig.GetWebServerConfig()
	microService.httpServer = web.NewHttpServer(webConfig)
	if microService.service.Init != nil {
		err := microService.service.Init(*configPath, microService.httpServer)
		if err != nil {
			return err
		}
	}
	serviceInfo := microService.service.GetServiceInfo()
	err = microService.registerEndpoints()
	if err != nil {
		return err
	}
	pprof.RegeditPprof(microService.httpServer)
	if base.IsDevModule() {
		debugConfig := serviceConfig.getDebugConfig()
		logger.Debug("open dev module")
		apiDefineRequestHandler := buildApiDefineRequestHandler(serviceInfo)
		if apiDefineRequestHandler != nil {
			microService.httpServer.Register(fmt.Sprintf("/apidefine/%s.api", serviceInfo.GetServiceName()), web.GET, apiDefineRequestHandler)
		}
		if debugConfig.IsEnableAccessInfo() {
			microService.httpServer.AddFirstFilter("/*", web.SimpleAccessLogFilter)
		}
	}

	return nil
}

func (microService *MicroService) start() base.Error {
	serviceInfo := microService.service.GetServiceInfo()
	//TODO 拦截异常返回
	errSign := microService.httpServer.Start()
	defer func() {
		err := <-errSign
		if err != nil {
			panic(base.NewError(base.ERROR_CODE_BASE_INIT_ERROR, err.Error()))
		}
	}()
	serviceDiscoveryRegister := microService.service.GetServiceDiscoveryRegister()
	if !microService.config.DisableServiceRegister && serviceDiscoveryRegister != nil {
		_, port, _ := net.SplitHostPort(microService.httpServer.GetServerAddress())
		p, _ := strconv.Atoi(port)
		registerError := serviceDiscoveryRegister.RegService(serviceInfo, microService.service.GetEndPoints(), p)
		if registerError != nil {
			logger.Info("注册服务[%s]失败,%s", microService.service.GetServiceInfo().GetServiceName(), registerError.Error())
			return registerError
		}
		logger.Info("注册服务[%s]成功", microService.service.GetServiceInfo().GetServiceName())
	}
	return nil
}

func buildApiDefineRequestHandler(serviceInfo base.ServiceInfo) web.RequestHandler {
	return func(reply web.Reply) {
		reply.With(serviceInfo.GetApiDefine()).As(web.Default_Render_Text)
	}
}

func (this *MicroService) registerEndpoint(endPoint base.EndPoint) base.Error {
	metadata := endPoint.Metadata
	logger.Debug("register endpoint [%s] %s %s", metadata.Method, metadata.Path, metadata.Description)
	err := this.httpServer.Register(metadata.Path, metadata.Method, endPoint.HandlerFunc)
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
