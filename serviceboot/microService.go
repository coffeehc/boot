package serviceboot

import (
	"errors"

	"fmt"
	"net/http"

	"flag"
	"github.com/coffeehc/logger"
	"github.com/coffeehc/microserviceboot/base"
	"github.com/coffeehc/web"
	"github.com/coffeehc/web/pprof"
	"os"
	"time"
)

type MicroService struct {
	server                   *web.Server
	service                  base.Service
	serviceDiscoveryRegister base.ServiceDiscoveryRegister
}

func newMicroService(service base.Service, serviceDiscoveryRegedit base.ServiceDiscoveryRegister) (*MicroService, error) {
	if flag.Lookup("help") != nil {
		flag.PrintDefaults()
		os.Exit(0)
	}
	serviceInfo := service.GetServiceInfo()
	if serviceInfo == nil {
		return nil, errors.New("没有指定 ServiceInfo")
	}
	webConfig := new(web.ServerConfig)
	webConfig.ServerAddr = fmt.Sprintf("%s:%d", base.GetLocalIp(), *port)
	webConfig.DefaultTransport = web.Transport_Json
	logger.Info("ServiceName: %s", serviceInfo.GetServiceName())
	logger.Info("Version: %s", serviceInfo.GetVersion())
	logger.Info("Descriptor: %s", serviceInfo.GetDescriptor())
	return &MicroService{
		server:                   web.NewServer(webConfig),
		service:                  service,
		serviceDiscoveryRegister: serviceDiscoveryRegedit,
	}, nil
}

func (this *MicroService) Start() error {
	logger.Info("Service startting")
	startTime := time.Now()
	serviceInfo := this.service.GetServiceInfo()
	err := this.regeditEndpoints()
	if err != nil {
		return err
	}
	pprof.RegeditPprof(this.server)
	if base.IsDevModule() {
		logger.Debug("open dev module")
		apiDefineRequestHandler := buildApiDefineRquestHandler(serviceInfo)
		if apiDefineRequestHandler != nil {
			this.server.Regedit(fmt.Sprintf("/apidefine/%s.api", serviceInfo.GetServiceName()), web.GET, apiDefineRequestHandler)
		}

		this.server.AddFirstFilter("/*", web.SimpleAccessLogFilter)
	}
	//TODO 拦截异常返回
	err = this.server.Start()
	if err != nil {
		return err
	}
	if this.serviceDiscoveryRegister != nil {
		err = this.serviceDiscoveryRegister.RegService(serviceInfo, this.service.GetEndPoints(), *port)
		if err != nil {
			return err
		}
	}
	logger.Info("Service started [%s]", time.Since(startTime))
	return nil
}

func buildApiDefineRquestHandler(serviceInfo base.ServiceInfo) web.RequestHandler {
	return func(request *http.Request, pathFragments map[string]string, reply web.Reply) {
		reply.With(serviceInfo.GetApiDefine()).As(web.Transport_Text)
	}
}

func (this *MicroService) regeditEndpoint(endPoint base.EndPoint) error {
	metadata := endPoint.Metadata
	logger.Info("register endpoint [%s] %s %s", metadata.Method, metadata.Path, metadata.Description)
	return this.server.Regedit(metadata.Path, metadata.Method, endPoint.HandlerFunc)
}

func (this *MicroService) regeditEndpoints() error {
	endPoints := this.service.GetEndPoints()
	if len(endPoints) == 0 {
		logger.Warn("not regedit any endpoint")
		return nil
	}
	for _, endPoint := range endPoints {
		err := this.regeditEndpoint(endPoint)
		if err != nil {
			return err
		}
	}
	return nil
}
