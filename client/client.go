package client

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/afex/hystrix-go/hystrix"
	"github.com/coffeehc/microserviceboot/common"
	"github.com/coffeehc/resty"
	"github.com/miekg/dns"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"strings"
)

type ServiceClient struct {
	client      *resty.Client
	serviceInfo common.ServiceInfo
	apiCallers  map[string]*ApiCaller
	config      *ServiceClientConfig
	dnsClient   *dns.Client
}

func NewServiceClient(config *ServiceClientConfig, clietnSetting func(client *resty.Client)) (*ServiceClient, error) {
	if config.Info == nil {
		return nil, errors.New("没有实现 ServiceInfo接口")
	}
	client := resty.New()
	loadbalancer := config.GetLoadBalancer()
	client.SetTransport(loadbalancer.getTransport())
	serviceClient := &ServiceClient{
		config:      config,
		client:      client,
		serviceInfo: config.Info,
		apiCallers:  make(map[string]*ApiCaller, 0),
	}
	client.SetHostURL(serviceClient.GetBaseUrl())
	client.SetDebug(common.IsDevModule())
	if clietnSetting != nil {
		clietnSetting(client)
	}
	return serviceClient, nil
}

func (this *ServiceClient) GetServiceName() string {
	return this.serviceInfo.GetServiceName()
}

func (this *ServiceClient) GetBaseUrl() string {
	tag := "pro"
	if common.IsDevModule() {
		tag = "dev"
	}
	return fmt.Sprintf("http://%s.%s.service.%s.%s", tag, this.serviceInfo.GetServiceName(), this.config.DataCenter, this.config.Domain)
}

func (this *ServiceClient) ApiRegiter(command string, apiRequest ApiRequest) error {
	if this.apiCallers[command] == nil {
		apiCaller := &ApiCaller{
			command:    command,
			apiRequest: apiRequest,
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
		request := this.client.R()
		resp, err := caller.apiRequest(request, query, body)
		response = resp
		return err
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
		request := this.client.R()
		resp, err := caller.apiRequest(request, query, body)
		response <- resp
		return err
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
