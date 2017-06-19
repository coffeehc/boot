package grpcboot

import (
	"context"
	"net"

	"github.com/coffeehc/httpx"
	"github.com/coffeehc/logger"
	"github.com/coffeehc/microserviceboot/base"
	"github.com/coffeehc/microserviceboot/base/grpcbase"
	"github.com/coffeehc/microserviceboot/serviceboot"
	"github.com/coffeehc/microserviceboot/serviceboot/internal"
	"github.com/grpc-ecosystem/go-grpc-prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
)

//GRPCMicroServiceBuilder 默认用于 grpc 的MicroServiceBuilder 实现
var GRPCMicroServiceBuilder serviceboot.MicroServiceBuilder = microServiceBuilder

func microServiceBuilder(service base.Service) (serviceboot.MicroService, base.Error) {
	grpcService, ok := service.(grpcbase.GRPCService)
	if !ok {
		return nil, base.NewError(-1, "GrpcMicroService build", "service 不是grpc 服务")
	}
	return &_GRPCMicroService{
		service:    grpcService,
		cleanFuncs: make([]func(), 0),
	}, nil
}

type _GRPCMicroService struct {
	service    grpcbase.GRPCService
	config     *Config
	httpServer httpx.Server
	grpcServer *grpc.Server
	cleanFuncs []func()
}

func (ms *_GRPCMicroService) Init(cxt context.Context) (*serviceboot.ServiceConfig, base.Error) {
	grpclog.SetLogger(&grpcLogger{})
	config := new(Config)
	configPath, err := internal.LoadConfig(config)
	if err != nil {
		return nil, err
	}
	ms.config = config
	err = internal.CheckServiceInfoConfig(ms.GetServiceInfo())
	if err != nil {
		return nil, err
	}
	httpServerConfig, err := config.GetServiceConfig().GetHTTPServerConfig()
	if err != nil {
		return nil, err
	}
	//构建 TSL
	if httpServerConfig.TLSConfig == nil {
		httpServerConfig.TLSConfig, err = serviceboot.NewDefaultTLSConfig()
		if err != nil {
			return nil, err
		}
	}
	tcpAddr, _ := net.ResolveTCPAddr("tcp", httpServerConfig.ServerAddr)
	httpServerConfig.TLSConfig.ServerName = tcpAddr.IP.String()
	httpServer, err := serviceboot.NewHTTPServer(httpServerConfig, ms.GetServiceInfo())
	if err != nil {
		return nil, err
	}
	ms.httpServer = httpServer
	grpcOptions := ms.config.GetGRPCOptions()
	if len(ms.service.GetGRPCOptions()) > 0 {
		grpcOptions = append(grpcOptions, ms.service.GetGRPCOptions()...)
	}
	ms.grpcServer = grpc.NewServer(grpcOptions...)
	err = ms.service.Init(cxt, configPath, httpServer)
	if err != nil {
		return nil, err
	}
	ms.service.RegisterServer(ms.grpcServer)
	grpc_prometheus.Register(ms.grpcServer)
	grpcFilter := &grpcFilter{ms.grpcServer}
	ms.httpServer.AddFirstFilter("*", grpcFilter.filter)
	if base.IsDevModule() && config.GetServiceConfig().EnableAccessInfo {
		ms.httpServer.AddFirstFilter("*", httpx.AccessLogFilter)
	}
	return config.GetServiceConfig(), nil
}

func (ms *_GRPCMicroService) Start(cxt context.Context) base.Error {
	err := internal.StartService(ms.service)
	if err != nil {
		return err
	}
	//启动服务器
	errSign := ms.httpServer.Start()
	go func() {
		err := <-errSign
		if ms.httpServer != nil && err != nil {
			panic(base.NewError(base.ErrCode_System, "GrpcMicroService start", err.Error()))
		}
	}()
	return nil
}

func (ms *_GRPCMicroService) AddCleanFunc(f func()) {
	ms.cleanFuncs = append(ms.cleanFuncs, f)
}

func (ms *_GRPCMicroService) Stop() {
	if ms.httpServer != nil {
		httpServer := ms.httpServer
		ms.httpServer = nil
		httpServer.Stop()
	}
	internal.StopService(ms.service)
	for _, f := range ms.cleanFuncs {
		func() {
			defer func() {
				if err := recover(); err != nil {
					logger.Error("clean func painc :%s", err)
				}
			}()
			f()
		}()
	}
}

func (ms *_GRPCMicroService) GetService() base.Service {
	return ms.service
}

func (ms *_GRPCMicroService) GetServiceInfo() base.ServiceInfo {
	return ms.config.GetServiceConfig().ServiceInfo
}
