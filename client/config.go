package client

import (
	"fmt"
	"github.com/benschw/dns-clb-go/clb"
	"github.com/coffeehc/microserviceboot/base"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net"
)

type ServiceClientConfig struct {
	Info            base.ServiceInfo `yaml:"serviceInfo"`
	DataCenter      string           `yaml:"dataCenter"`
	Domain          string           `yaml:"domain"`
	DNSAddress      string           `yaml:"nameServer"`
	LoadBalanceType LoadBalanceType  `yaml:"loadBalanceType"`
	DirectBaseUrl   string           `yaml:"directBaseUrl"`
}

func LoadServiceClientConfig(configFile string) (*ServiceClientConfig, error) {
	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("加载配置文件错误:%s", err)
	}
	config := new(ServiceClientConfig)
	err = yaml.Unmarshal(data, config)
	if err != nil {
		return nil, fmt.Errorf("解析配置文件错误:%s", err)
	}
	return config, nil
}

type LoadBalanceType string

var (
	LoadBalance_RoundRobin = LoadBalanceType("RoundRobin")
	LoadBalance_Random     = LoadBalanceType("Random")
)

func (this *ServiceClientConfig) GetLoadBalancer() *LoadBalancer {
	if this.LoadBalanceType == "" {
		this.LoadBalanceType = LoadBalance_Random
	}
	var loadBalanceType clb.LoadBalancerType
	switch this.LoadBalanceType {
	case LoadBalance_RoundRobin:
		loadBalanceType = clb.RoundRobin
	case LoadBalance_Random:
		loadBalanceType = clb.Random
	default:
		loadBalanceType = clb.RoundRobin
	}
	if this.DNSAddress == "" {
		return newDefaultLoadBalancer(loadBalanceType)
	}
	host, port, err := net.SplitHostPort(this.DNSAddress)
	if err != nil {
		return newDefaultLoadBalancer(loadBalanceType)
	}
	return newLoadBalancer(host, port, loadBalanceType)

}
