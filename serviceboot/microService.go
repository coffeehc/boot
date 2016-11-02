package serviceboot

import (
	"github.com/coffeehc/microserviceboot/base"
)

type MicroService interface {
	Init() (*ServiceConfig, base.Error)
	Start() base.Error
	Stop()
	GetService() base.Service
}

type MicroServiceBuilder func(base.Service) (MicroService, base.Error)
