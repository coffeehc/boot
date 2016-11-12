package serviceboot

import (
	"flag"
	"fmt"
	"github.com/coffeehc/commons"
)

var configPath = flag.String("config", "", "配置文件路径")

type ServiceConfig struct {
	Debug                  *DebugConfig `yaml:"debug"`
	DisableServiceRegister bool         `yaml:"disable_service_register"`
	ServerAddr             string       `yaml:"server_addr"`
}

func (this *ServiceConfig) GetDebugConfig() *DebugConfig {
	if this.Debug == nil {
		this.Debug = &DebugConfig{}
	}
	return this.Debug
}

func (this *ServiceConfig) GetServerAddr() string {
	if this.ServerAddr == "" {
		this.ServerAddr = fmt.Sprintf("%s:8888", commons.GetLocalIp())
	}
	return this.ServerAddr
}
