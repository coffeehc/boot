package serviceboot

import (
	"fmt"

	"github.com/coffeehc/commons"
	"github.com/coffeehc/httpx"
	"github.com/coffeehc/microserviceboot/base"
)

// ServiceConfig 服务配置
type ServiceConfig struct {
	ServiceInfo            *base.SimpleServiceInfo `yaml:"service_info"`
	EnableAccessInfo       bool                    `yaml:"enableAccessInfo"`
	DisableServiceRegister bool                    `yaml:"disable_service_register"`
	HTTPConfig             *httpx.Config           `yaml:"web_server_config"`
}

//GetHTTPServerConfig 获取 HTTP config
func (sc *ServiceConfig) GetHTTPServerConfig() *httpx.Config {
	if sc.HTTPConfig == nil {
		sc.HTTPConfig = new(httpx.Config)
	}
	if sc.HTTPConfig.ServerAddr == "" {
		sc.HTTPConfig.ServerAddr = fmt.Sprintf("%s:8888", commons.GetLocalIP())
	}
	sc.HTTPConfig.DefaultRender = httpx.DefaultRenderJSON
	return sc.HTTPConfig
}
