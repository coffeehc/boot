package consultool

import (
	"github.com/coffeehc/logger"
	"github.com/coffeehc/microserviceboot/base"
	"github.com/hashicorp/consul/api"
)

//NewClientByConfigFile 通过配置文件创建一个新的 Consul Client
func NewClientByConfigFile(configPath string) (*api.Client, base.Error) {
	config := loadConsulConfig(configPath)
	return NewClient(config)
}

//NewClient 创建 consul client
func NewClient(config *ConsulConfig) (*api.Client, base.Error) {
	if config == nil {
		config = &ConsulConfig{}
	}
	consulClient, err := api.NewClient(warpConsulConfig(config))
	if err != nil {
		logger.Error("创建 Consul Client 失败")
		return nil, base.NewError(base.ErrCodeBaseSystemInit, "consul init", err.Error())
	}
	return consulClient, nil
}

func warpConsulConfig(consulConfig *ConsulConfig) *api.Config {
	if consulConfig == nil {
		return nil
	}
	config := api.DefaultConfig()
	config.Address = consulConfig.GetAddress()
	config.Scheme = consulConfig.GetScheme()
	config.Datacenter = consulConfig.GetDataCenter()
	config.WaitTime = consulConfig.GetWaitTime()
	config.Token = consulConfig.GetToken()
	if consulConfig.BasicAuth != nil {
		config.HttpAuth = &api.HttpBasicAuth{
			Username: consulConfig.BasicAuth.Username,
			Password: consulConfig.BasicAuth.Password,
		}
	}
	return config
}
