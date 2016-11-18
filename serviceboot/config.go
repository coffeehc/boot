package serviceboot

import (
	"flag"
	"fmt"
	"github.com/coffeehc/commons"
	"github.com/coffeehc/microserviceboot/base"
	"github.com/coffeehc/web"
	"time"
)

var configPath = flag.String("config", "", "配置文件路径")

type ServiceConfig struct {
	ServiceInfo            *base.SimpleServiceInfo `yaml:"service_info"`
	Debug                  *DebugConfig            `yaml:"debug"`
	DisableServiceRegister bool                    `yaml:"disable_service_register"`
	WebServerConfig        *WebConfig              `yaml:"web_server_config"`
}

func (this *ServiceConfig) GetDebugConfig() *DebugConfig {
	if this.Debug == nil {
		this.Debug = &DebugConfig{}
	}
	return this.Debug
}

type WebConfig struct {
	ServerAddr   string        `yaml:"server_addr"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
	Concurrency  int           `yaml:"concurrency"` //暂时没有使用
}

func (this *WebConfig) GetServerAddr() string {
	if this.ServerAddr == "" {
		this.ServerAddr = fmt.Sprintf("%s:8888", commons.GetLocalIp())
	}
	return this.ServerAddr
}

func (this *ServiceConfig) GetWebServerConfig() *web.HttpServerConfig {
	webConfig := new(web.HttpServerConfig)
	wc := this.WebServerConfig
	if wc == nil {
		wc = new(WebConfig)
	}
	webConfig.ServerAddr = wc.GetServerAddr()
	if wc.Concurrency == 0 {
		wc.Concurrency = 100000
	}
	webConfig.ReadTimeout = wc.ReadTimeout
	webConfig.WriteTimeout = wc.WriteTimeout
	webConfig.DefaultRender = web.Default_Render_Json
	return webConfig
}
