package grpcboot

import (
	"google.golang.org/grpc"

	"context"
	"github.com/coffeehc/microserviceboot/base"
	"github.com/coffeehc/microserviceboot/base/grpcbase"
	"github.com/coffeehc/microserviceboot/serviceboot"
	"github.com/coffeehc/web"
	"github.com/grpc-ecosystem/go-grpc-prometheus"
	"google.golang.org/grpc/grpclog"
	"net"
)

var GRpcMicroServiceBuilder serviceboot.MicroServiceBuilder = microServiceBuilder

func microServiceBuilder(service base.Service) (serviceboot.MicroService, base.Error) {
	grpcService, ok := service.(grpcbase.GRpcService)
	if !ok {
		return nil, base.NewError(-1, "GrpcMicroService build", "service 不是Rest 服务")
	}
	return &GRpcMicroService{
		service: grpcService,
	}, nil
}

type GRpcMicroService struct {
	service    grpcbase.GRpcService
	config     *Config
	httpServer web.HttpServer
	grpcServer *grpc.Server
}

func (this *GRpcMicroService) Init(cxt context.Context) (*serviceboot.ServiceConfig, base.Error) {
	grpclog.SetLogger(&grpcbase.GrpcLogger{})
	config := new(Config)
	configPath, err := serviceboot.LoadConfig(config)
	if err != nil {
		return nil, err
	}
	this.config = config
	err = serviceboot.CheckServiceInfoConfig(this.GetServiceInfo())
	if err != nil {
		return nil, err
	}
	webServerConfig := config.GetBaseConfig().GetWebServerConfig()
	//构建 TSL
	if webServerConfig.TLSConfig == nil {
		webServerConfig.TLSConfig, err = newDefaultTlsConfig()
		if err != nil {
			return nil, err
		}
	}
	tcpAddr, _ := net.ResolveTCPAddr("tcp", webServerConfig.ServerAddr)
	webServerConfig.TLSConfig.ServerName = tcpAddr.IP.String()
	httpServer, err := serviceboot.NewHttpServer(webServerConfig, this.GetServiceInfo())
	if err != nil {
		return nil, err
	}
	this.httpServer = httpServer
	serviceboot.ServiceRegister(configPath, this.GetService(), this.GetServiceInfo(), config.GetBaseConfig(), cxt)
	if this.service.Init != nil {
		err := this.service.Init(configPath, httpServer, cxt)
		if err != nil {
			return nil, err
		}
	}
	grpcServerConfig := this.config.GetGRpcServerConfig()
	grpcOptions := grpcServerConfig.GetGrpcOptions()
	if len(this.service.GetGrpcOptions()) > 0 {
		grpcOptions = append(grpcOptions, this.service.GetGrpcOptions()...)
	}
	//TODO ???
	//grpcOptions = append(grpcOptions,grpc.Creds(credentials.NewServerTLSFromCert(&webServerConfig.TLSConfig.Certificates[0])))
	grpc.EnableTracing = false
	if base.IsDevModule() {
		grpc.EnableTracing = true
	}
	this.grpcServer = grpc.NewServer(grpcOptions...)
	this.service.RegisterServer(this.grpcServer)
	grpc_prometheus.Register(this.grpcServer)
	grpcFilter := &grpcFilter{this.grpcServer}
	this.httpServer.AddFirstFilter("*", grpcFilter.filter)
	return config.GetBaseConfig(), nil
}

func (this *GRpcMicroService) Start() base.Error {
	//启动服务器
	errSign := this.httpServer.Start()
	go func() {
		err := <-errSign
		if this.httpServer != nil && err != nil {
			panic(base.NewError(base.ERRCODE_BASE_SYSTEM_INIT_ERROR, "GrpcMicroService start", err.Error()))
		}
	}()
	return nil
}

func (this *GRpcMicroService) Stop() {
	if this.httpServer != nil {
		httpServer := this.httpServer
		this.httpServer = nil
		httpServer.Stop()
	}
}

func (this *GRpcMicroService) GetService() base.Service {
	return this.service
}

func (this *GRpcMicroService) GetServiceInfo() base.ServiceInfo {
	return this.config.GetServiceConfig().ServiceInfo
}
