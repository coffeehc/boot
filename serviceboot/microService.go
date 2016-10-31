package serviceboot

import (
	"github.com/coffeehc/microserviceboot/base"
)

type MicroService interface {
	Init() base.Error
	Start() base.Error
}

type MicroServiceBuilder func(base.Service) (MicroService, base.Error)
