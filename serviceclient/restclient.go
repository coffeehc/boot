package serviceclient

import (
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/afex/hystrix-go/hystrix"
	"github.com/coffeehc/logger"
	"github.com/coffeehc/microserviceboot/base"
	"github.com/go-resty/resty"
)

func init() {
	hystrix.DefaultMaxConcurrent = 2000
	hystrix.DefaultVolumeThreshold = 4000
	hystrix.DefaultTimeout = 30000
}

func newServiceClientByDirectBaseUrl(serviceInfo base.ServiceInfo, directBaseUrl string) (*resty.Client, base.Error) {
	d := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}
	transport := &http.Transport{
		MaxIdleConnsPerHost: 10,
		Dial:                d.Dial,
	}
	return NewRestClient(directBaseUrl, transport, nil)
}

func newServiceClientByConsul(serviceInfo base.ServiceInfo, serviceClientDNSConfig ServiceClientConsulConfig) (*resty.Client, base.Error) {
	logger.Info("dataCenter is [%s]", serviceClientDNSConfig.GetDataCenter())
	logger.Info("dnsAddress is [%s]", serviceClientDNSConfig.GetNameServer())
	ip, port, err := net.SplitHostPort(serviceClientDNSConfig.GetNameServer())
	if err != nil {
		return nil, base.NewError(base.ERROR_CODE_BASE_INIT_ERROR,err.Error())
	}
	loadBalancer := newLoadBalancer(ip, port, serviceClientDNSConfig.GetLoadBalanceType())
	return NewRestClient(fmt.Sprintf("http://%s/", buildServiceDomain(serviceInfo.GetServiceName(), serviceClientDNSConfig)), loadBalancer.getTransport(), nil)
}

func NewRestClient(baseUrl string, transport *http.Transport, clientSetting func(client *resty.Client)) (*resty.Client, base.Error) {
	client := resty.New()
	client.SetTransport(transport)
	client.SetHeader("Accept", "application/json")
	client.SetHeader("User-Agent", "serviceboot httpclient 0.1")
	client.SetRESTMode()
	client.SetHostURL(baseUrl)
	client.SetDebug(base.IsDevModule())
	if clientSetting != nil {
		clientSetting(client)
	}
	return client, nil
}
