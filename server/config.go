package server

import (
	"github.com/coffeehc/microserviceboot/common"
)

type MicorServiceCofig struct {
	Service common.Service
	//WebConfig *web.ServerConfig
	DevModule bool
}
