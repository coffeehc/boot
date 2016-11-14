package grpcboot

import (
	"google.golang.org/grpc"

	"github.com/coffeehc/microserviceboot/base"
	"github.com/coffeehc/microserviceboot/base/grpcbase"
	"github.com/coffeehc/microserviceboot/serviceboot"
	"github.com/coffeehc/web"
	"github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc/grpclog"
)

var GRpcMicroServiceBuilder serviceboot.MicroServiceBuilder = microServiceBuilder

func microServiceBuilder(service base.Service) (serviceboot.MicroService, base.Error) {
	grpcService, ok := service.(grpcbase.GRpcService)
	if !ok {
		return nil, base.NewError(-1, "service 不是Rest 服务")
	}
	return &GRpcMicroService{
		service: grpcService,
	}, nil
}

type GRpcMicroService struct {
	service grpcbase.GRpcService
	config  *Config
	//listener   net.Listener
	httpServer web.HttpServer
	grpcServer *grpc.Server
}

func (this *GRpcMicroService) Init() (*serviceboot.ServiceConfig, base.Error) {
	grpclog.SetLogger(&_logger{})
	serviceConfig := new(Config)
	configPath := serviceboot.LoadConfigPath(serviceConfig)
	this.config = serviceConfig
	httpServer, err := serviceboot.NewHttpServer(configPath, serviceConfig.GetBaseConfig().GetWebServerConfig(), this.service)
	if err != nil {
		return nil, err
	}
	this.httpServer = httpServer
	if this.service.Init != nil {
		err := this.service.Init(configPath, httpServer)
		if err != nil {
			return nil, err
		}
	}
	grpcServerConfig := this.config.GetGRpcServerConfig()
	grpcOptions := grpcServerConfig.GetGrpcOptions()
	if len(this.service.GetGrpcOptions()) > 0 {
		grpcOptions = append(grpcOptions, this.service.GetGrpcOptions()...)
	}
	grpc.EnableTracing = false
	if base.IsDevModule() {
		grpc.EnableTracing = true
	}
	this.grpcServer = grpc.NewServer(grpcOptions...)
	this.service.RegisterServer(this.grpcServer)
	grpc_prometheus.Register(this.grpcServer)
	grpcFilter := &grpcFilter{this.grpcServer}
	this.httpServer.AddFirstFilter("/*", grpcFilter.filter)
	err1 := this.httpServer.RegisterHttpHandler("/metrics", web.GET, prometheus.Handler())
	if err1 != nil {
		return nil, base.NewErrorWrapper(err1)
	}
	return serviceConfig.GetBaseConfig(), nil
}

func (this *GRpcMicroService) Start() base.Error {
	//启动服务器
	errSign := this.httpServer.Start()
	go func() {
		err := <-errSign
		if err != nil {
			panic(base.NewError(base.ERROR_CODE_BASE_INIT_ERROR, err.Error()))
		}
	}()
	return nil
}

func (this *GRpcMicroService) Stop() {
	if this.httpServer != nil {
		this.httpServer.Stop()
	}
}

func (this *GRpcMicroService) GetService() base.Service {
	return this.service
}
