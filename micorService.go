package microserviceboot

import (
	"errors"

	"fmt"
	"github.com/coffeehc/logger"
	"github.com/coffeehc/microserviceboot/common"
	"github.com/coffeehc/web"
	"net/http"
)

type MicorService struct {
	config                  *MicorServiceCofig
	server                  *web.Server
	service                 common.Service
	serviceDiscoveryRegedit ServiceDiscoveryRegister
}

func newMicorService(config *MicorServiceCofig, serviceDiscoveryRegedit ServiceDiscoveryRegister) (*MicorService, error) {
	if config.WebConfig == nil {
		config.WebConfig = new(web.ServerConfig)
	}
	config.WebConfig.DefaultTransport = web.Transport_Json
	serviceInfo := config.Service.GetServiceInfo()
	if serviceInfo == nil {
		return nil, errors.New("没有指定 ServiceInfo")
	}
	logger.Info("ServiceName: %s", serviceInfo.GetServiceName())
	logger.Info("Version: %s", serviceInfo.GetVersion())
	logger.Info("Descriptor: %s", serviceInfo.GetDescriptor())
	return &MicorService{
		config:                  config,
		server:                  web.NewServer(config.WebConfig),
		service:                 config.Service,
		serviceDiscoveryRegedit: serviceDiscoveryRegedit,
	}, nil
}

func (this *MicorService) Start() error {
	serviceInfo := this.service.GetServiceInfo()
	logger.Info("MicorService start")
	err := this.regeditEndpoints()
	if err != nil {
		return err
	}
	if this.config.DevModule {
		logger.Debug("open dev module")
		apiDefineRquestHandler := buildApiDefineRquestHandler(serviceInfo)
		if apiDefineRquestHandler != nil {
			this.server.Regedit(fmt.Sprintf("/apidefine/%s.api", serviceInfo.GetServiceName()), web.GET, apiDefineRquestHandler)
		}
		this.server.AddFirstFilter("/*", web.SimpleAccessLogFilter)
	}
	err = this.server.Start()
	if err != nil {
		return err
	}
	if this.serviceDiscoveryRegedit != nil {
		err = this.serviceDiscoveryRegedit.RegService(this.config.WebConfig.ServerAddr, serviceInfo, this.service.GetEndPoints())
		if err != nil {
			return err
		}
	}
	return nil
}

func buildApiDefineRquestHandler(serviceInfo common.ServiceInfo) web.RequestHandler {
	return func(request *http.Request, pathFragments map[string]string, reply web.Reply) {
		reply.With(serviceInfo.GetApiDefine()).As(web.Transport_Text)
	}
}

func (this *MicorService) regeditEndpoint(endPoint common.EndPoint) error {
	//TODO
	return this.server.Regedit(endPoint.Path, endPoint.Method, endPoint.HandlerFunc)
}

func (this *MicorService) regeditEndpoints() error {
	endPoints := this.service.GetEndPoints()
	if len(endPoints) == 0 {
		return errors.New("not regedit any endpoint")
	}
	for _, endPoint := range endPoints {
		err := this.regeditEndpoint(endPoint)
		if err != nil {
			return err
		}
	}
	return nil
}
