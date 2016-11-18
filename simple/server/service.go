package main

import (
	"github.com/coffeehc/microserviceboot/base"
	"github.com/coffeehc/microserviceboot/consultool"
	"github.com/coffeehc/microserviceboot/simple/simplemodel"
	"github.com/coffeehc/web"
	"google.golang.org/grpc"
)

type Service struct {
}

func (this *Service) GetGrpcOptions() []grpc.ServerOption {
	return nil
}
func (this *Service) Init(configPath string, httpServer web.HttpServer) base.Error {
	return nil
}
func (this *Service) RegisterServer(s *grpc.Server) base.Error {
	simplemodel.RegisterGreeterServer(s, &_GreeterServer{})
	return nil
}
func (this *Service) Run() base.Error {
	return nil
}
func (this *Service) Stop() base.Error {
	return nil
}

func (this *Service) GetServiceDiscoveryRegister() (base.ServiceDiscoveryRegister, base.Error) {
	consulClient, err := consultool.NewConsulClient(&consultool.ConsulConfig{})
	if err != nil {
		return nil, err
	}
	return consultool.NewConsulServiceRegister(consulClient)
}
