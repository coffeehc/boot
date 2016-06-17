package serviceclient

import (
	"github.com/benschw/dns-clb-go/clb"
)

type LoadBalanceType string

type ServiceClientConsulConfig struct {
	DataCenter      string          `yaml:"dataCenter"`
	Domain          string          `yaml:"domain"`
	NameServer      string          `yaml:"nameServe"`
	LoadBalanceType LoadBalanceType `yaml:"loadBalanceType"`
}

func (this ServiceClientConsulConfig) GetNameServer() string {
	if this.NameServer == "" {
		return "127.0.0.1:8600"
	}
	return this.NameServer
}

func (this ServiceClientConsulConfig) GetDomain() string {
	if this.Domain == "" {
		return "xiagaogao"
	}
	return this.Domain
}

func (this ServiceClientConsulConfig) GetDataCenter() string {
	if this.DataCenter == "" {
		return "dc"
	}
	return this.DataCenter
}

func (this ServiceClientConsulConfig) GetLoadBalanceType() clb.LoadBalancerType {
	var loadBalanceType clb.LoadBalancerType
	switch this.LoadBalanceType {
	case LoadBalance_RoundRobin:
		loadBalanceType = clb.RoundRobin
	case LoadBalance_Random:
		loadBalanceType = clb.Random
	default:
		loadBalanceType = clb.RoundRobin
	}
	return loadBalanceType
}
