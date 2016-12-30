package main

import (
	"github.com/coffeehc/httpx"
	"github.com/coffeehc/microserviceboot/base"
	"github.com/coffeehc/microserviceboot/consultool"
	"github.com/coffeehc/microserviceboot/simple/simplemodel"
	"google.golang.org/grpc"
)

type _Service struct {
}

func (service *_Service) GetGrpcOptions() []grpc.ServerOption {
	return nil
}
func (service *_Service) Init(configPath string, httpServer httpx.Server) base.Error {
	return nil
}
func (service *_Service) RegisterServer(s *grpc.Server) base.Error {
	simplemodel.RegisterGreeterServer(s, &_GreeterServer{})
	return nil
}
func (service *_Service) Run() base.Error {
	return nil
}
func (service *_Service) Stop() base.Error {
	return nil
}

func (service *_Service) GetServiceDiscoveryRegister() (base.ServiceDiscoveryRegister, base.Error) {
	consulClient, err := consultool.NewClient(&consultool.ConsulConfig{})
	if err != nil {
		return nil, err
	}
	return consultool.NewConsulServiceRegister(consulClient)
}
