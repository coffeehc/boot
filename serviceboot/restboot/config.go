package restboot

import (
	"fmt"
	"github.com/coffeehc/microserviceboot/base"
	"github.com/coffeehc/microserviceboot/serviceboot"
	"github.com/coffeehc/web"
	"time"
)

type Config struct {
	serviceboot.ServiceConfig
	WebServerConfig *WebConfig `yaml:"webserver"`
}

type WebConfig struct {
	ServerAddr   string        `yaml:"serverAddr"`
	ReadTimeout  time.Duration `yaml:"readTimeout"`
	WriteTimeout time.Duration `yaml:"writeTimeout"`
	Concurrency  int           `yaml:"concurrency"` //暂时没有使用
}

func (this *Config) GetWebServerConfig() *web.HttpServerConfig {
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
