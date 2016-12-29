package restboot

import "github.com/coffeehc/microserviceboot/serviceboot"

//Config restboot config
type Config struct {
	ServiceConfig *serviceboot.ServiceConfig `yaml:"service_config"`
}

//GetServiceConfig 实现的 ServiceConfiguration 的接口
func (config *Config) GetServiceConfig() *serviceboot.ServiceConfig {
	return config.ServiceConfig
}
