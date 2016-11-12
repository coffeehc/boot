package restboot

import (
	"fmt"
	"github.com/coffeehc/commons"
	"github.com/coffeehc/microserviceboot/serviceboot"
	"github.com/coffeehc/web"
	"time"
)

type Config struct {
	BaseConfig      *serviceboot.ServiceConfig `yaml:"base_config"`
	WebServerConfig *WebConfig                 `yaml:"web_server_config"`
}

type WebConfig struct {
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
	Concurrency  int           `yaml:"concurrency"` //暂时没有使用
}

func (this *Config) GetBaseConfig() *serviceboot.ServiceConfig {
	if this.BaseConfig == nil {
		this.BaseConfig = new(serviceboot.ServiceConfig)
	}
	return this.BaseConfig
}

func (this *Config) GetWebServerConfig() *web.HttpServerConfig {
	webConfig := new(web.HttpServerConfig)
	wc := this.WebServerConfig
	if wc == nil {
		wc = new(WebConfig)
	}
	if this.BaseConfig.ServerAddr == "" {
		webConfig.ServerAddr = fmt.Sprintf("%s:8888", commons.GetLocalIp())
	} else {
		webConfig.ServerAddr = this.BaseConfig.ServerAddr
	}
	if wc.Concurrency == 0 {
		wc.Concurrency = 100000
	}
	webConfig.ReadTimeout = time.Duration(wc.ReadTimeout / time.Second)
	webConfig.WriteTimeout = time.Duration(wc.WriteTimeout / time.Second)
	webConfig.DefaultRender = web.Default_Render_Json
	return webConfig
}
