package restbase

import (
	"github.com/coffeehc/microserviceboot/base"
	"github.com/coffeehc/web"
)

type RestService interface {
	base.Service
	Init(configPath string, server web.HttpServer) base.Error
	GetEndPoints() []EndPoint
}
