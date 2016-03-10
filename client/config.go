package client

import (
	"github.com/benschw/dns-clb-go/clb"
	"github.com/coffeehc/microserviceboot/common"
	"net"
)

type ServiceClientConfig struct {
	DevModule       bool
	Info            common.ServiceInfo
	DataCenter      string
	Domean          string
	DNSAddress      string
	LoadBalanceType LoadBalanceType
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
