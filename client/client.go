package client

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/afex/hystrix-go/hystrix"
	"github.com/coffeehc/microserviceboot/base"
	"github.com/coffeehc/resty"
	"github.com/miekg/dns"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"strings"
)

type ServiceClient struct {
	client      *resty.Client
	serviceInfo base.ServiceInfo
	apiCallers  map[string]*ApiCaller
	config      *ServiceClientConfig
	dnsClient   *dns.Client
}

func NewServiceClient(config *ServiceClientConfig, clientSetting func(client *resty.Client)) (*ServiceClient, error) {
	if config.Info == nil {
		return nil, errors.New("没有实现 ServiceInfo接口")
	}
	client := resty.New()
	if config.DNSAddress != "" {
		loadbalancer := config.GetLoadBalancer()
		client.SetTransport(loadbalancer.getTransport())
	}
	client.SetHeader("User-Agent", "micorserviceboot httpclient 0.1")
	client.SetRESTMode()
	serviceClient := &ServiceClient{
		config:      config,
		client:      client,
		serviceInfo: config.Info,
		apiCallers:  make(map[string]*ApiCaller, 0),
	}
	client.SetHostURL(serviceClient.GetBaseUrl())
	client.SetDebug(base.IsDevModule())
	if clientSetting != nil {
		clientSetting(client)
	}
	return serviceClient, nil
}

func (this *ServiceClient) GetServiceName() string {
	return this.serviceInfo.GetServiceName()
}

func (this *ServiceClient) GetBaseUrl() string {
	config := this.config
	if config.DirectBaseUrl != "" {
		return config.DirectBaseUrl
	}
	tag := "pro"
	if base.IsDevModule() {
		tag = "dev"
	}
	return fmt.Sprintf("http://%s.%s.service.%s.%s", tag, this.serviceInfo.GetServiceName(), this.config.DataCenter, this.config.Domain)
}

func (this *ServiceClient) ApiRegiter(command string, method RequestMethod, uri string, apiRequestSetting ApiRequestSetting) error {
	if this.apiCallers[command] == nil {
		apiCaller := &ApiCaller{
			command:           command,
			apiRequestSetting: apiRequestSetting,
			method:            method,
			uri:               uri,
		}
		this.apiCallers[command] = apiCaller
		return nil
	}
	return fmt.Errorf("command[%s]已经存在,不能再次注册", command)

}

func (this *ServiceClient) SyncCallApiExt(command string, query map[string]string, body interface{}, result interface{}) error {
	resp, err := this.SyncCallApi(command, query, body)
	if err != nil {
		return err
	}
	contentType := resp.Header().Get("Content-Type")
	switch {
	case strings.HasPrefix(contentType, "application/json"):
		return json.Unmarshal(resp.Body(), result)
	case strings.HasPrefix(contentType, "text/xml"):
		return xml.Unmarshal(resp.Body(), result)
	default:
		return errors.New("can't decode response data")
	}
}

func (this *ServiceClient) SyncCallApi(command string, query map[string]string, body interface{}) (*resty.Response, error) {
	caller, ok := this.apiCallers[command]
	if !ok {
		return nil, fmt.Errorf("没有注册过cmmand[%s]", command)
	}
	var response *resty.Response
	err := hystrix.Do(command, func() error {
		var err1 error
		response, err1 = doCommand(this.client, caller, query, body)
		return err1
	}, func(err error) error {
		//TODO 处理异常
		return err
	})
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (this *ServiceClient) AsyncCallApi(command string, query map[string]string, body interface{}) (chan<- *resty.Response, error) {
	caller, ok := this.apiCallers[command]
	if !ok {
		return nil, fmt.Errorf("没有注册过cmmand[%s]", command)
	}
	response := make(chan *resty.Response)
	err := hystrix.Go(command, func() error {
		var err1 error
		res, err1 := doCommand(this.client, caller, query, body)
		response <- res
		return err1
	}, func(err error) error {
		//TODO 处理异常
		close(response)
		return err
	})
	if err != nil {
		close(response)
		return response, <-err
	}
	return response, nil
}

func doCommand(client *resty.Client, caller *ApiCaller, query map[string]string, body interface{}) (*resty.Response, error) {
	request := client.R()
	if caller.apiRequestSetting != nil {
		caller.apiRequestSetting(request)
	}
	request.SetQueryParams(query)
	request.SetBody(body)
	return request.Execute(string(caller.method), caller.uri)

}
