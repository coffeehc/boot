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
	WebServerConfig          *WebConfig   `yaml:"webserver"`
	Debug                    *DebugConfig `yaml:"debug"`
	DisEnableServiceRegister bool         `yaml:"disenableServiceRegister"`
}

type WebConfig struct {
	ServerAddr     string        `yaml:"serverAddr"`
	ReadTimeout    time.Duration `yaml:"readTimeout"`
	WriteTimeout   time.Duration `yaml:"eriteTimeout"`
	MaxHeaderBytes int           `yaml:"maxHeaderBytes"`
}

func (this *ServiceConfig) getDebugConfig() *DebugConfig {
	if this.Debug == nil {
		this.Debug = &DebugConfig{}
	}
	return this.Debug
}

func (this *ServiceConfig) GetWebServerConfig() *web.ServerConfig {
	webConfig := new(web.ServerConfig)
	wc := this.WebServerConfig
	if wc == nil {
		wc = new(WebConfig)
	}
	if wc.ServerAddr == "" {
		webConfig.ServerAddr = fmt.Sprintf("%s:8888", base.GetLocalIp())
	} else {
		webConfig.ServerAddr = wc.ServerAddr
	}
	//host,ip,err:=net.SplitHostPort(webConfig.ServerAddr)
	//if host == ""
	webConfig.ReadTimeout = wc.ReadTimeout
	webConfig.WriteTimeout = wc.WriteTimeout
	webConfig.MaxHeaderBytes = wc.MaxHeaderBytes
	webConfig.DefaultTransport = web.Transport_Json
	return webConfig
}
