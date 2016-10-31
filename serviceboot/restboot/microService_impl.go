package restboot

import (
	"fmt"
	"github.com/coffeehc/logger"
	"github.com/coffeehc/microserviceboot/base"
	"github.com/coffeehc/microserviceboot/base/restbase"
	"github.com/coffeehc/microserviceboot/serviceboot"
	"github.com/coffeehc/web"
	"github.com/coffeehc/web/pprof"
	"net"
	"strconv"
)

const RestMicroServiceBuilder serviceboot.MicroServiceBuilder = microServiceBuild

type MicroService_Rest struct {
	config     *serviceboot.ServiceConfig
	httpServer web.HttpServer
	service    restbase.RestService
}

func microServiceBuild(service base.Service) (serviceboot.MicroService, base.Error) {
	restService, ok := service.(restbase.RestService)
	if !ok {
		return nil, base.NewError(-1, "service 不是Rest 服务")
	}

	return &MicroService_Rest{
		service: restService,
	}, nil
}

func (microService *MicroService_Rest) Init() base.Error {
	logger.Info("Service startting")
	serviceConfig := new(Config)
	configPath := serviceboot.LoadConfigPath(serviceConfig)
	microService.config = serviceConfig
	webConfig := serviceConfig.GetWebServerConfig()
	microService.httpServer = web.NewHttpServer(webConfig)
	if microService.service.Init != nil {
		err := microService.service.Init(configPath, microService.httpServer)
		if err != nil {
			return err
		}
	}
	serviceInfo := microService.service.GetServiceInfo()
	err := microService.registerEndpoints()
	if err != nil {
		return err
	}
	pprof.RegeditPprof(microService.httpServer)
	if base.IsDevModule() {
		debugConfig := serviceConfig.GetDebugConfig()
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

func (microService *MicroService_Rest) Start() base.Error {
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

func (this *MicroService_Rest) registerEndpoint(endPoint restbase.EndPoint) base.Error {
	metadata := endPoint.Metadata
	logger.Debug("register endpoint [%s] %s %s", metadata.Method, metadata.Path, metadata.Description)
	err := this.httpServer.Register(metadata.Path, metadata.Method, endPoint.HandlerFunc)
	if err != nil {
		return base.NewError(base.ERROR_CODE_BASE_INIT_ERROR, err.Error())
	}
	return nil
}

func (this *MicroService_Rest) registerEndpoints() base.Error {
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
