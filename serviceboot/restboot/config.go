package restboot

import "github.com/coffeehc/microserviceboot/serviceboot"

type Config struct {
	BaseConfig *serviceboot.ServiceConfig `yaml:"base_config"`
}

func (this *Config) GetServiceConfig() *serviceboot.ServiceConfig {
	return this.BaseConfig
}
