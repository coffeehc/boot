package serviceboot

import (
	"flag"
)

var configPath = flag.String("config", "", "配置文件路径")

type ServiceConfig struct {
	Debug                  *DebugConfig `yaml:"debug"`
	DisableServiceRegister bool         `yaml:"disableServiceRegister"`
}

func (this *ServiceConfig) GetDebugConfig() *DebugConfig {
	if this.Debug == nil {
		this.Debug = &DebugConfig{}
	}
	return this.Debug
}
