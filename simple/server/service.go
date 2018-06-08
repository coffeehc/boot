package main

import (
	"context"

	"git.xiagaogao.com/coffee/boot/errors"
	"git.xiagaogao.com/coffee/boot/serviceboot"
	"git.xiagaogao.com/coffee/boot/simple/simplemodel"
	"google.golang.org/grpc"
)

type ServiceImpl struct {
}

func (service *ServiceImpl) Init(cxt context.Context, configPath string, serviceConfig *serviceboot.ServiceConfig) errors.Error {
	return nil
}
func (service *ServiceImpl) RegisterServer(s *grpc.Server) errors.Error {
	simplemodel.RegisterGreeterServer(s, &_GreeterServer{})
	return nil
}
func (service *ServiceImpl) Run(cxt context.Context) errors.Error {
	return nil
}
func (service *ServiceImpl) Stop(cxt context.Context) errors.Error {
	return nil
}

func (service *ServiceImpl) GetServiceDiscoveryRegister() (serviceboot.ServiceDiscoveryRegister, errors.Error) {
	//consulClient, err := consultool.NewClient(&consultool.ConsulConfig{})
	//if err != nil {
	//	return nil, err
	//}
	//return consultool.NewConsulServiceRegister(consulClient)
	return nil, nil
}
