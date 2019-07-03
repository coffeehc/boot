package main

import (
	"context"

	"git.xiagaogao.com/coffee/boot"
	"git.xiagaogao.com/coffee/boot/errors"
	"git.xiagaogao.com/coffee/boot/serviceboot"
	"git.xiagaogao.com/coffee/boot/simple/simplemodel"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type ServiceImpl struct {
	logger       *zap.Logger
	errorSerivce errors.Service
}

func (service *ServiceImpl) Init(cxt context.Context, kit serviceboot.ServiceKit) errors.Error {
	service.logger = kit.GetLogger()
	service.errorSerivce = kit.GetRootErrorService()
	return nil
}
func (service *ServiceImpl) RegisterServer(s *grpc.Server) errors.Error {
	simplemodel.RegisterGreeterServer(s, &_GreeterServer{service.logger, service.errorSerivce})
	return nil
}
func (service *ServiceImpl) Run(cxt context.Context) errors.Error {
	return nil
}
func (service *ServiceImpl) Stop(cxt context.Context) errors.Error {
	return nil
}

func (service *ServiceImpl) GetServiceInfo() *boot.ServiceInfo {
	return getServiceInfo()
}

func getServiceInfo() *boot.ServiceInfo {
	return &boot.ServiceInfo{
		ServiceName: "simple_service",
		Version:     "0.0.1",
		Descriptor:  "测试Server",
		Scheme:      "http",
	}
}
