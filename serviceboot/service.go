package serviceboot

import (
	"context"

	"git.xiagaogao.com/coffee/boot/errors"
	"google.golang.org/grpc"
)

// Service 接口定义
type Service interface {
	Init(cxt context.Context, configPath string, serviceConfig *ServiceConfig) errors.Error
	Run(cxt context.Context) errors.Error
	Stop(cxt context.Context) errors.Error
	GetServiceDiscoveryRegister() (ServiceDiscoveryRegister, errors.Error)
	RegisterServer(s *grpc.Server) errors.Error
}
