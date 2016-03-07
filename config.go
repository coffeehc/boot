package microserviceboot

import (
	"github.com/coffeehc/microserviceboot/common"
	"github.com/coffeehc/web"
)

type MicorServiceCofig struct {
	Service   common.Service
	WebConfig *web.ServerConfig
	DevModule bool
}
