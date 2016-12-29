package restclient

import (
	"fmt"

	"context"

	"github.com/afex/hystrix-go/hystrix"
	"github.com/coffeehc/microserviceboot/base"
)

const errScopeRestClient = "restClient"

type ServiceClient interface {
	GetBaseUrl() string
	GetHttpClient() HttpClient
	Call(cxt context.Context, request Request) (Response, base.Error)
}

func NewServiceClient(serviceInfo base.ServiceInfo, httpClientConfig *HttpClientConfiguration, discoveryConfig interface{}) (ServiceClient, base.Error) {
	if httpClientConfig == nil {
		httpClientConfig = defaultHttpClientConfiguration
	}
	rootCxt := context.TODO()
	var httpClient HttpClient = nil
	var err base.Error
	switch c := discoveryConfig.(type) {
	case string:
		httpClient, err = newHttpClientByHostAddress(c, httpClientConfig, rootCxt)
	case *string:
		httpClient, err = newHttpClientByHostAddress(*c, httpClientConfig, rootCxt)
	case ServiceClientConsulConfig:
		httpClient, err = newHttpClientByConsul(serviceInfo, c, httpClientConfig, rootCxt)
	case *ServiceClientConsulConfig:
		if c == nil {
			c = &ServiceClientConsulConfig{}
		}
		httpClient, err = newHttpClientByConsul(serviceInfo, *c, httpClientConfig, rootCxt)
	default:
		err = base.NewError(base.ErrCodeBaseSystemInit, errScopeRestClient, fmt.Sprintf("无法识别的配置类型,%#v", discoveryConfig))
	}
	if err != nil {
		return nil, base.NewError(base.ErrCodeBaseSystemInit, errScopeRestClient, fmt.Sprintf("不能识别的 config 类型:%#v", discoveryConfig))
	}
	return &_ServiceClient{
		client:      httpClient,
		serviceInfo: serviceInfo,
	}, nil
}

type _ServiceClient struct {
	client      HttpClient
	serviceInfo base.ServiceInfo
}

func (this *_ServiceClient) GetServiceName() string {
	return this.serviceInfo.GetServiceName()
}

func (this *_ServiceClient) GetBaseUrl() string {
	return this.GetHttpClient().GetBaseUrl()
}

func (this *_ServiceClient) GetHttpClient() HttpClient {
	return this.client
}

func (this *_ServiceClient) Call(cxt context.Context, request Request) (Response, base.Error) {
	var response Response
	err := hystrix.Do(request.GetCommand(), func() error {
		var err1 error
		response, err1 = this.client.Do(cxt, request)
		return err1
	}, func(err error) error {
		//logger.Error("请求异常:%#v", err)
		cxt.Done()
		//TODO 处理异常
		return err
	})
	if bErr, ok := err.(base.Error); ok {
		return response, bErr
	}
	return response, base.NewErrorWrapper(errScopeRestClient, err)
}
