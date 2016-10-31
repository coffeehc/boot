package restclient

import (
	"net"
	"net/http"
	"time"

	"github.com/benschw/dns-clb-go/clb"
	"github.com/coffeehc/logger"
)

var (
	LoadBalance_RoundRobin = LoadBalanceType("RoundRobin")
	LoadBalance_Random     = LoadBalanceType("Random")
)

type LoadBalancer struct {
	loadBalancer clb.LoadBalancer
	dialer       *net.Dialer
	transport    *http.Transport
}

func newLoadBalancer(nameServer string, port string, lbType clb.LoadBalancerType) *LoadBalancer {
	loadBalancer := clb.NewTtlCacheClb(nameServer, port, lbType, 1)
	return _newLoadBalancer(loadBalancer)
}

func _newLoadBalancer(loadBalancer clb.LoadBalancer) *LoadBalancer {
	this := &LoadBalancer{
		loadBalancer: loadBalancer,
		dialer: &net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		},
	}
	this.transport = &http.Transport{
		MaxIdleConnsPerHost: 10,
		Dial:                this.Dial,
	}
	return this
}

func (this *LoadBalancer) Dial(network, address string) (conn net.Conn, err error) {
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		return nil, err
	}
	addr, err := this.loadBalancer.GetAddress(host)
	if err != nil {
		logger.Error("%s\n", err)
		return nil, err
	}
	return this.dialer.Dial(network, addr.String())
}

func (this *LoadBalancer) getTransport() *http.Transport {
	return &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		Dial:                  this.Dial,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 30 * time.Second,
	}
}
