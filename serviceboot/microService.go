package serviceboot

import (
	"context"

	"git.xiagaogao.com/coffee/boot/errors"
)

//MicroService micro serviceboot interface define
type MicroService interface {
	Start(ctx context.Context, serviceConfig *ServiceConfig, configPath string) errors.Error
	Stop(context.Context)
	GetService() Service
	GetServiceInfo() ServiceInfo
	AddCleanFunc(func())
}
