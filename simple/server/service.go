package main

import (
	"github.com/coffeehc/microserviceboot/base"
	"google.golang.org/grpc"
)

type Service struct {
}

func (this *Service) GetGrpcOptions() []grpc.ServerOption {
	return nil
}
func (this *Service) Init(configPath string) base.Error {
	return nil
}
func (this *Service) RegisterServer(s *grpc.Server) base.Error {
	return nil
}
func (this *Service) Run() base.Error {
	return nil
}
func (this *Service) Stop() base.Error {
	return nil
}
func (this *Service) GetServiceInfo() base.ServiceInfo {
	return &ServiceInfo{}
}
func (this *Service) GetServiceDiscoveryRegister() base.ServiceDiscoveryRegister {
	return nil
}
