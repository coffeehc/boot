package restbase

import (
	"github.com/coffeehc/microserviceboot/base"
)

// RestService Rest service interface define
type RestService interface {
	base.Service
	GetEndPoints() []Endpoint
}
