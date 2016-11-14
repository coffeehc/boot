package restboot

import (
	"fmt"
	"github.com/coffeehc/logger"
	"github.com/coffeehc/microserviceboot/base"
	"github.com/coffeehc/microserviceboot/base/restbase"
	"github.com/coffeehc/microserviceboot/serviceboot"
	"github.com/coffeehc/web"
	"github.com/coffeehc/web/pprof"
)

var RestMicroServiceBuilder serviceboot.MicroServiceBuilder = microServiceBuild

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

func (this *MicroService_Rest) Init() (*serviceboot.ServiceConfig, base.Error) {
	serviceConfig := new(serviceboot.ServiceConfig)
	configPath := serviceboot.LoadConfigPath(serviceConfig)
	this.config = serviceConfig
	httpServer, err := serviceboot.NewHttpServer(configPath, serviceConfig.GetWebServerConfig(), this.service)
	if err != nil {
		return nil, err
	}
	this.httpServer = httpServer
	serviceInfo := this.service.GetServiceInfo()
	err = this.registerEndpoints()
	if err != nil {
		return nil, err
	}
	pprof.RegeditPprof(this.httpServer)
	if base.IsDevModule() {
		debugConfig := serviceConfig.GetDebugConfig()
		logger.Debug("open dev module")
		apiDefineRequestHandler := buildApiDefineRequestHandler(serviceInfo)
		if apiDefineRequestHandler != nil {
			this.httpServer.Register(fmt.Sprintf("/apidefine/%s.api", serviceInfo.GetServiceName()), web.GET, apiDefineRequestHandler)
		}
		if debugConfig.IsEnableAccessInfo() {
			this.httpServer.AddFirstFilter("/*", web.SimpleAccessLogFilter)
		}
	}

	return serviceConfig, nil
}

func (this *MicroService_Rest) Start() base.Error {
	errSign := this.httpServer.Start()
	go func() {
		err := <-errSign
		if err != nil {
			panic(base.NewError(base.ERROR_CODE_BASE_INIT_ERROR, err.Error()))
		}
	}()
	return nil
}

func (this *MicroService_Rest) GetService() base.Service {
	return this.service
}

func (this *MicroService_Rest) Stop() {
	if this.httpServer != nil {
		this.httpServer.Stop()
	}
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
