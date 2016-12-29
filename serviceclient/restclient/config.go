package restclient

import (
	"time"

	"github.com/benschw/dns-clb-go/clb"
)

var defaultHttpClientConfiguration = &HttpClientConfiguration{
	KeepAlive: 30,
}

type LoadBalanceType string

type HttpClientConfiguration struct {
	KeepAlive             time.Duration //KeepAlive 的秒数
	Timeout               time.Duration // 连接超时时间
	MaxIdleConns          int
	MaxIdleConnsPerHost   int
	IdleConnTimeout       time.Duration
	TLSHandshakeTimeout   time.Duration
	ExpectContinueTimeout time.Duration
}

func (config *HttpClientConfiguration) GetKeepAlive() time.Duration {
	if config.KeepAlive == 0 {
		config.KeepAlive = time.Duration(30)
	}
	return config.KeepAlive * time.Second
}
func (config *HttpClientConfiguration) GetTimeout() time.Duration {
	if config.Timeout == 0 {
		config.Timeout = time.Duration(30)
	}
	return config.Timeout * time.Second
}

func (config *HttpClientConfiguration) GetIdleConnTimeout() time.Duration {
	if config.IdleConnTimeout == 0 {
		config.IdleConnTimeout = time.Duration(90)
	}
	return config.IdleConnTimeout * time.Second
}

func (config *HttpClientConfiguration) GetTLSHandshakeTimeout() time.Duration {
	if config.TLSHandshakeTimeout == 0 {
		config.TLSHandshakeTimeout = time.Duration(10)
	}
	return config.TLSHandshakeTimeout * time.Second
}

func (config *HttpClientConfiguration) GetExpectContinueTimeout() time.Duration {
	if config.ExpectContinueTimeout == 0 {
		config.ExpectContinueTimeout = time.Duration(1)
	}
	return config.ExpectContinueTimeout * time.Second
}

func (config *HttpClientConfiguration) GetMaxIdleConns() int {
	if config.MaxIdleConns == 0 {
		config.MaxIdleConns = 40
	}
	return config.MaxIdleConns
}

func (config *HttpClientConfiguration) GetMaxIdleConnsPerHost() int {
	if config.MaxIdleConnsPerHost == 0 {
		config.MaxIdleConnsPerHost = 10
	}
	return config.MaxIdleConnsPerHost
}

type ServiceClientConsulConfig struct {
	DataCenter      string          `yaml:"dataCenter"`
	Domain          string          `yaml:"domain"`
	NameServer      string          `yaml:"nameServe"`
	LoadBalanceType LoadBalanceType `yaml:"loadBalanceType"`
}

func (config ServiceClientConsulConfig) GetNameServer() string {
	if config.NameServer == "" {
		return "127.0.0.1:8600"
	}
	return config.NameServer
}

func (config ServiceClientConsulConfig) GetDomain() string {
	if config.Domain == "" {
		return "xiagaogao"
	}
	return config.Domain
}

func (config ServiceClientConsulConfig) GetDataCenter() string {
	if config.DataCenter == "" {
		return "dc"
	}
	return config.DataCenter
}

func (config ServiceClientConsulConfig) GetLoadBalanceType() clb.LoadBalancerType {
	var loadBalanceType clb.LoadBalancerType
	switch config.LoadBalanceType {
	case LoadBalance_RoundRobin:
		loadBalanceType = clb.RoundRobin
	case LoadBalance_Random:
		loadBalanceType = clb.Random
	default:
		loadBalanceType = clb.RoundRobin
	}
	return loadBalanceType
}
