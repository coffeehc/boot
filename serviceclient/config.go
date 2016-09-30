package serviceclient

import (
	"time"

	"github.com/benschw/dns-clb-go/clb"
)

var Default_HttpClientConfiguration = &HttpClientConfiguration{
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

func (this *HttpClientConfiguration) GetKeepAlive() time.Duration {
	if this.KeepAlive == 0 {
		this.KeepAlive = time.Duration(30)
	}
	return this.KeepAlive * time.Second
}
func (this *HttpClientConfiguration) GetTimeout() time.Duration {
	if this.Timeout == 0 {
		this.Timeout = time.Duration(30)
	}
	return this.Timeout * time.Second
}

func (this *HttpClientConfiguration) GetIdleConnTimeout() time.Duration {
	if this.IdleConnTimeout == 0 {
		this.IdleConnTimeout = time.Duration(90)
	}
	return this.IdleConnTimeout * time.Second
}

func (this *HttpClientConfiguration) GetTLSHandshakeTimeout() time.Duration {
	if this.TLSHandshakeTimeout == 0 {
		this.TLSHandshakeTimeout = time.Duration(10)
	}
	return this.TLSHandshakeTimeout * time.Second
}

func (this *HttpClientConfiguration) GetExpectContinueTimeout() time.Duration {
	if this.ExpectContinueTimeout == 0 {
		this.ExpectContinueTimeout = time.Duration(1)
	}
	return this.ExpectContinueTimeout * time.Second
}

func (this *HttpClientConfiguration) GetMaxIdleConns() int {
	if this.MaxIdleConns == 0 {
		this.MaxIdleConns = 40
	}
	return this.MaxIdleConns
}

func (this *HttpClientConfiguration) GetMaxIdleConnsPerHost() int {
	if this.MaxIdleConnsPerHost == 0 {
		this.MaxIdleConnsPerHost = 10
	}
	return this.MaxIdleConnsPerHost
}

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
