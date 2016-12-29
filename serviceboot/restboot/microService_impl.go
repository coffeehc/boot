package restboot

import (
	"context"
	"fmt"

	"github.com/coffeehc/httpx"
	"github.com/coffeehc/logger"
	"github.com/coffeehc/microserviceboot/base"
	"github.com/coffeehc/microserviceboot/base/restbase"
	"github.com/coffeehc/microserviceboot/serviceboot"
	"github.com/coffeehc/microserviceboot/serviceboot/internal"
)

//RestMicroServiceBuilder rest 服务的MicroServiceBuilder
var RestMicroServiceBuilder serviceboot.MicroServiceBuilder = microServiceBuild

type _RestMicroService struct {
	config     *Config
	httpServer httpx.Server
	service    restbase.RestService
}

func microServiceBuild(service base.Service) (serviceboot.MicroService, base.Error) {
	restService, ok := service.(restbase.RestService)
	if !ok {
		return nil, base.NewError(-1, "RestMicroService build", "service 不是Rest 服务")
	}
	return &_RestMicroService{
		service: restService,
	}, nil
}

func (ms *_RestMicroService) GetServiceInfo() base.ServiceInfo {
	return ms.config.GetServiceConfig().ServiceInfo
}

func (ms *_RestMicroService) Init(cxt context.Context) (*serviceboot.ServiceConfig, base.Error) {
	config := new(Config)
	configPath, err := internal.LoadConfig(config)
	if err != nil {
		return nil, err
	}
	ms.config = config
	serviceConfig := config.GetServiceConfig()
	err = internal.CheckServiceInfoConfig(ms.GetServiceInfo())
	if err != nil {
		return nil, err
	}
	if base.IsDevModule() {
		logger.Debug("open dev module")
		apiDefineRequestHandler := buildAPIDefineRequestHandler(ms.GetServiceInfo())
		if apiDefineRequestHandler != nil {
			ms.httpServer.Register(fmt.Sprintf("/apidefine/%s.api", ms.GetServiceInfo().GetServiceName()), httpx.GET, apiDefineRequestHandler)
		}
		if serviceConfig.EnableAccessInfo {
			ms.httpServer.AddFirstFilter("/*", httpx.AccessLogFilter)
		}
	}
	httpServer, err := serviceboot.NewHTTPServer(serviceConfig.GetHTTPServerConfig(), ms.GetServiceInfo())
	if err != nil {
		return nil, err
	}
	ms.httpServer = httpServer
	err = ms.registerEndpoints()
	if err != nil {
		return nil, err
	}
	if ms.service.Init != nil {
		err := ms.service.Init(cxt, configPath, httpServer)
		if err != nil {
			return nil, err
		}
	}
	return serviceConfig, nil
}

func (ms *_RestMicroService) Start(cxt context.Context) base.Error {
	err := internal.StartService(ms.service)
	if err != nil {
		return err
	}
	errSign := ms.httpServer.Start()
	go func() {
		err := <-errSign
		if err != nil {
			panic(base.NewError(base.ErrCodeBaseSystemInit, "RestMicroService Start", err.Error()))
		}
	}()
	return nil
}

func (ms *_RestMicroService) GetService() base.Service {
	return ms.service
}

func (ms *_RestMicroService) Stop() {
	if ms.httpServer != nil {
		ms.httpServer.Stop()
	}
	internal.StopService(ms.service)
}

func buildAPIDefineRequestHandler(serviceInfo base.ServiceInfo) httpx.RequestHandler {
	return func(reply httpx.Reply) {
		reply.With(serviceInfo.GetAPIDefine()).As(httpx.DefaultRenderText)
	}
}

func (ms *_RestMicroService) registerEndpoint(endPoint restbase.Endpoint) base.Error {
	metadata := endPoint.Metadata
	logger.Debug("register endpoint [%s] %s %s", metadata.Method, metadata.Path, metadata.Description)
	err := ms.httpServer.Register(metadata.Path, metadata.Method, endPoint.HandlerFunc)
	if err != nil {
		return base.NewError(base.ErrCodeBaseSystemInit, "RestMicroService register", err.Error())
	}
	return nil
}

func (ms *_RestMicroService) registerEndpoints() base.Error {
	endPoints := ms.service.GetEndPoints()
	if len(endPoints) == 0 {
		logger.Warn("not regedit any endpoint")
		return nil
	}
	for _, endPoint := range endPoints {
		err := ms.registerEndpoint(endPoint)
		if err != nil {
			return err
		}
	}
	return nil
}
