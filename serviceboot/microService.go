package serviceboot

import (
	"context"
	"github.com/coffeehc/microserviceboot/base"
)

type MicroService interface {
	Init(context.Context) (*ServiceConfig, base.Error)
	Start() base.Error
	Stop()
	GetService() base.Service
	GetServiceInfo() base.ServiceInfo
}

type MicroServiceBuilder func(base.Service) (MicroService, base.Error)
