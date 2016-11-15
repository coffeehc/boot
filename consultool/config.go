package consultool

import (
	"time"

	"github.com/hashicorp/consul/api"
)

type ConsulConfig struct {
	Address    string        `yaml:"address"`
	Scheme     string        `yaml:"scheme"`
	DataCenter string        `yaml:"daraCenter"`
	WaitTime   time.Duration `yaml:"waitTime"`
	Token      string        `yaml:"token"`
	BasicAuth  HttpBasicAuth `yaml:"basic_auth"`
}

type HttpBasicAuth struct {
	Username string
	Password string
}

func (this *ConsulConfig) GetAddress() string {
	if this.Address == "" {
		return "127.0.0.1:8500"
	}
	return this.Address
}

func (this *ConsulConfig) GetScheme() string {
	if this.Scheme == "" {
		return "http"
	}
	return this.Scheme
}

func (this *ConsulConfig) GetDataCenter() string {
	if this.DataCenter == "" {
		return "dc"
	}
	return this.DataCenter
}

func (this *ConsulConfig) GetWaitTime() time.Duration {
	return this.WaitTime
}

func (this *ConsulConfig) GetToken() string {
	return this.Token
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
	config.HttpAuth = &api.HttpBasicAuth{
		Username: consulConfig.BasicAuth.Username,
		Password: consulConfig.BasicAuth.Password,
	}
	return config
}
