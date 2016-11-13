package grpcboot

import (
	"net"

	"google.golang.org/grpc"

	"github.com/coffeehc/logger"
	"github.com/coffeehc/microserviceboot/base"
	"github.com/coffeehc/microserviceboot/base/grpcbase"
	"github.com/coffeehc/microserviceboot/serviceboot"
	"github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc/grpclog"
	"net/http"
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
	service    grpcbase.GRpcService
	config     *Config
	listener   net.Listener
	grpcServer *grpc.Server
}

func (this *GRpcMicroService) Init() (*serviceboot.ServiceConfig, base.Error) {
	grpclog.SetLogger(&_logger{})
	serviceConfig := new(Config)
	configPath := serviceboot.LoadConfigPath(serviceConfig)
	this.config = serviceConfig
	if this.service.Init != nil {
		err := this.service.Init(configPath)
		if err != nil {
			return nil, err
		}
	}
	grpcServerConfig := this.config.GetGRpcServerConfig()
	grpcOptions := grpcServerConfig.GetGrpcOptions()
	if len(this.service.GetGrpcOptions()) > 0 {
		grpcOptions = append(grpcOptions, this.service.GetGrpcOptions()...)
	}
	this.grpcServer = grpc.NewServer(grpcOptions...)
	this.service.RegisterServer(this.grpcServer)
	grpc_prometheus.Register(this.grpcServer)
	http.Handle("/metrics", prometheus.Handler())
	return serviceConfig.GetBaseConfig(), nil
}

func (this *GRpcMicroService) Start() base.Error {
	//启动服务器
	lis, err := net.Listen("tcp", this.config.GetBaseConfig().GetServerAddr())
	if err != nil {
		return base.NewError(-1, logger.Error("failed to listen: %v", err))
	}
	this.listener = lis
	go this.grpcServer.Serve(lis)
	return nil
}

func (this *GRpcMicroService) Stop() {
	if this.listener != nil {
		this.listener.Close()
	}
}

func (this *GRpcMicroService) GetService() base.Service {
	return this.service
}
