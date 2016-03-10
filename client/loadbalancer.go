package client

import (
	"fmt"
	"github.com/benschw/dns-clb-go/clb"
	"net"
	"net/http"
	"time"
)

type LoadBalancer struct {
	loadBalancer clb.LoadBalancer
	dialer       *net.Dialer
	transport    *http.Transport
}

func newDefaultLoadBalancer(lbType clb.LoadBalancerType) *LoadBalancer {
	loadBalancer := clb.NewDefaultClb(lbType)
	return _newLoadBalancer(loadBalancer)
}

func newLoadBalancer(nameServer string, port string, lbType clb.LoadBalancerType) *LoadBalancer {
	loadBalancer := clb.NewClb(nameServer, port, lbType)
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
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return nil, err
	}
	addr, err := this.loadBalancer.GetAddress(host)
	if err != nil {
		fmt.Printf("%s\n", err)
		return nil, err
	}
	return this.dialer.Dial(network, net.JoinHostPort(addr.Address, port))
}

func (this *LoadBalancer) getTransport() *http.Transport {
	return &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		Dial:                  this.Dial,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 30 * time.Second,
	}
}
