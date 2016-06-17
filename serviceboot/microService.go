package serviceboot

import (
	"errors"

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

func newMicroService(service base.Service) (*MicroService, error) {
	if flag.Lookup("help") != nil {
		flag.PrintDefaults()
		os.Exit(0)
	}
	serviceInfo := service.GetServiceInfo()
	if serviceInfo == nil {
		return nil, errors.New("没有指定 ServiceInfo")
	}
	logger.Info("ServiceName: %s", serviceInfo.GetServiceName())
	logger.Info("Version: %s", serviceInfo.GetVersion())
	logger.Info("Descriptor: %s", serviceInfo.GetDescriptor())
	return &MicroService{
		service: service,
	}, nil
}

func (this *MicroService) init() error {
	logger.Info("Service startting")
	serverConfig := new(ServiceConfig)
	*configPath = base.GetDefaultConfigPath(*configPath)
	err := base.LoadConfig(*configPath, serverConfig)
	if err != nil {
		logger.Warn("加载服务器配置失败,%s", err)
	}
	webConfig := serverConfig.GetWebServerConfig()
	this.server = web.NewServer(webConfig)
	if this.service.Init != nil {
		err := this.service.Init(*configPath, this.server)
		if err != nil {
			return err
		}
	}
	serviceInfo := this.service.GetServiceInfo()
	err = this.registerEndpoints()
	if err != nil {
		return err
	}
	pprof.RegeditPprof(this.server)
	if base.IsDevModule() {
		debugConfig := serverConfig.getDebugConfig()
		logger.Debug("open dev module")
		apiDefineRequestHandler := buildApiDefineRequestHandler(serviceInfo)
		if apiDefineRequestHandler != nil {
			this.server.Register(fmt.Sprintf("/apidefine/%s.api", serviceInfo.GetServiceName()), web.GET, apiDefineRequestHandler)
		}
		if debugConfig.IsEnableAccessInfo() {
			this.server.AddFirstFilter("/*", web.SimpleAccessLogFilter)
		}
	}

	return nil
}

func (this *MicroService) start() error {
	startTime := time.Now()
	serviceInfo := this.service.GetServiceInfo()
	//TODO 拦截异常返回
	err := this.server.Start()
	if err != nil {
		return err
	}
	serviceDiscoveryRegister := this.service.GetServiceDiscoveryRegister()
	if serviceDiscoveryRegister != nil {
		_, port, _ := net.SplitHostPort(this.server.GetServerAddress())
		p, _ := strconv.Atoi(port)
		err = serviceDiscoveryRegister.RegService(serviceInfo, this.service.GetEndPoints(), p)
		if err != nil {
			logger.Info("注册服务[%s]失败,%s", this.service.GetServiceInfo().GetServiceName(), err)
			return err
		}
		logger.Info("注册服务[%s]成功", this.service.GetServiceInfo().GetServiceName())
	}
	logger.Info("Service started [%s]", time.Since(startTime))
	return nil
}

func buildApiDefineRequestHandler(serviceInfo base.ServiceInfo) web.RequestHandler {
	return func(request *http.Request, pathFragments map[string]string, reply web.Reply) {
		reply.With(serviceInfo.GetApiDefine()).As(web.Transport_Text)
	}
}

func (this *MicroService) registerEndpoint(endPoint base.EndPoint) error {
	metadata := endPoint.Metadata
	logger.Info("register endpoint [%s] %s %s", metadata.Method, metadata.Path, metadata.Description)
	return this.server.Register(metadata.Path, metadata.Method, endPoint.HandlerFunc)
}

func (this *MicroService) registerEndpoints() error {
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
