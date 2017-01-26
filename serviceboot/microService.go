package serviceboot

import (
	"context"

	"github.com/coffeehc/microserviceboot/base"
)

//MicroService micro service interface define
type MicroService interface {
	Init(context.Context) (*ServiceConfig, base.Error)
	Start(context.Context) base.Error
	Stop()
	GetService() base.Service
	GetServiceInfo() base.ServiceInfo
	AddCleanFunc(func())
}

//MicroServiceBuilder MicroService Builder function define
type MicroServiceBuilder func(base.Service) (MicroService, base.Error)
