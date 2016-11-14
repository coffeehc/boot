package restbase

import (
	"github.com/coffeehc/microserviceboot/base"
)

type RestService interface {
	base.Service
	GetEndPoints() []EndPoint
}
