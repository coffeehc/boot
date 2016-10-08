package serviceboot

import (
	"flag"
	"fmt"
	"time"

	"github.com/coffeehc/microserviceboot/base"
	"github.com/coffeehc/web"
)

var configPath = flag.String("config", "", "配置文件路径")

type ServiceConfig struct {
	WebServerConfig        *WebConfig   `yaml:"webserver"`
	Debug                  *DebugConfig `yaml:"debug"`
	DisableServiceRegister bool         `yaml:"disableServiceRegister"`
}

type WebConfig struct {
	ServerAddr   string        `yaml:"serverAddr"`
	ReadTimeout  time.Duration `yaml:"readTimeout"`
	WriteTimeout time.Duration `yaml:"writeTimeout"`
	Concurrency  int           `yaml:"concurrency"` //暂时没有使用
}

func (this *ServiceConfig) getDebugConfig() *DebugConfig {
	if this.Debug == nil {
		this.Debug = &DebugConfig{}
	}
	return this.Debug
}

func (this *ServiceConfig) GetWebServerConfig() *web.HttpServerConfig {
	webConfig := new(web.HttpServerConfig)
	wc := this.WebServerConfig
	if wc == nil {
		wc = new(WebConfig)
	}
	if wc.ServerAddr == "" {
		webConfig.ServerAddr = fmt.Sprintf("%s:8888", base.GetLocalIp())
	} else {
		webConfig.ServerAddr = wc.ServerAddr
	}
	if wc.Concurrency == 0 {
		wc.Concurrency = 100000
	}
	webConfig.ReadTimeout = time.Duration(wc.ReadTimeout / time.Second)
	webConfig.WriteTimeout = time.Duration(wc.WriteTimeout / time.Second)
	webConfig.DefaultRender = web.Default_Render_Json
	return webConfig
}
