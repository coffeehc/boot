package restclient

import (
	"net"

	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/afex/hystrix-go/hystrix"
	"github.com/coffeehc/logger"
	"github.com/coffeehc/microserviceboot/base"
	"github.com/coffeehc/microserviceboot/base/restbase"
)

func init() {
	hystrix.DefaultMaxConcurrent = 2000
	hystrix.DefaultVolumeThreshold = 4000
	hystrix.DefaultTimeout = 30000
}

type HttpClient interface {
	GetBaseUrl() string
	BuildRequest(*restbase.EndPointMeta) (Request, base.Error)
	Do(cxt context.Context, req Request) (Response, base.Error)
}

type _HttpClientWithBase struct {
	baseUrl string
	cxt     context.Context
	client  *http.Client
}

func (this *_HttpClientWithBase) GetBaseUrl() string {
	return this.baseUrl
}

func (this *_HttpClientWithBase) BuildRequest(endPointMeta *restbase.EndPointMeta) (Request, base.Error) {
	request := AcquireRequest()
	err := request.Init(this.baseUrl, endPointMeta)
	if err != nil {
		return nil, err
	}
	return request, nil
}

func (this *_HttpClientWithBase) Do(cxt context.Context, req Request) (Response, base.Error) {
	if header, ok := cxt.Value("defaultHeader").(http.Header); ok {
		for key, values := range header {
			for _, v := range values {
				req.GetHeader().Add(key, v)
			}
		}
	}
	if request, ok := req.(*_Request); ok {
		response, err := this.client.Do(request.request)
		if err != nil {
			return nil, base.NewErrorWrapper(err)
		}
		return buildResponse(response), nil
	} else {
		return nil, base.NewError(-1, "not support Resqust implement")
	}
}

func buildTransport(config *HttpClientConfiguration) *http.Transport {
	disableKeepAlives := false
	if config.GetKeepAlive() == 0 {
		disableKeepAlives = true
	}
	transport := &http.Transport{
		Proxy: nil, //http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   config.GetTimeout(),
			KeepAlive: config.GetKeepAlive(),
		}).DialContext,
		MaxIdleConns:          config.GetMaxIdleConns(),
		MaxIdleConnsPerHost:   config.GetMaxIdleConnsPerHost(),
		IdleConnTimeout:       config.GetIdleConnTimeout(),
		TLSHandshakeTimeout:   config.GetTLSHandshakeTimeout(),
		ExpectContinueTimeout: config.GetExpectContinueTimeout(),
		DisableKeepAlives:     disableKeepAlives,
	}
	return transport
	//TODO 还差 Dial 和 DialTLS的处理
}

func newHttpClientByHostAddress(baseUrl string, config *HttpClientConfiguration, cxt context.Context) (HttpClient, base.Error) {
	_url, err := url.Parse(baseUrl)
	if err != nil {
		return nil, base.NewErrorWrapper(err)
	}
	logger.Debug("client host is %s", _url.Host)
	logger.Debug("client baseUrl is %s", baseUrl)
	transport := buildTransport(config)
	httpClient := &http.Client{
		Timeout:   config.GetTimeout(),
		Transport: transport,
	}
	return &_HttpClientWithBase{
		cxt:     cxt,
		baseUrl: baseUrl,
		client:  httpClient,
	}, nil

}

func newHttpClientByConsul(serviceInfo base.ServiceInfo, serviceClientDNSConfig ServiceClientConsulConfig, httpConfig *HttpClientConfiguration, cxt context.Context) (HttpClient, base.Error) {
	logger.Info("dataCenter is [%s]", serviceClientDNSConfig.GetDataCenter())
	logger.Info("dnsAddress is [%s]", serviceClientDNSConfig.GetNameServer())
	ip, port, err := net.SplitHostPort(serviceClientDNSConfig.GetNameServer())
	if err != nil {
		return nil, base.NewError(base.ERROR_CODE_BASE_INIT_ERROR, err.Error())
	}
	loadBalancer := newLoadBalancer(ip, port, serviceClientDNSConfig.GetLoadBalanceType())
	host := buildServiceDomain(serviceInfo.GetServiceName(), serviceClientDNSConfig)
	transport := buildTransport(httpConfig)
	transport.Dial = loadBalancer.Dial
	httpClient := &http.Client{
		Timeout:   httpConfig.GetTimeout(),
		Transport: transport,
	}
	return &_HttpClientWithBase{
		cxt:     cxt,
		baseUrl: fmt.Sprintf("http://%s/", host),
		client:  httpClient,
	}, nil
}
