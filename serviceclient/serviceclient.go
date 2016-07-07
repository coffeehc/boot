package serviceclient

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/url"
	"strings"

	"github.com/afex/hystrix-go/hystrix"
	"github.com/coffeehc/microserviceboot/base"
	"github.com/go-resty/resty"
	"github.com/golang/protobuf/proto"
)

func NewServiceClient(serviceInfo base.ServiceInfo, discoveryConfig interface{}) (*ServiceClient, error) {
	var restClient *resty.Client
	var err error
	switch c := discoveryConfig.(type) {
	case string:
		restClient, err = newServiceClientByDirectBaseUrl(serviceInfo, c)
	case *string:
		restClient, err = newServiceClientByDirectBaseUrl(serviceInfo, *c)
	case ServiceClientConsulConfig:
		restClient, err = newServiceClientByConsul(serviceInfo, c)
	case *ServiceClientConsulConfig:
		if c == nil {
			c = &ServiceClientConsulConfig{}
		}
		restClient, err = newServiceClientByConsul(serviceInfo, *c)
	default:
		err = fmt.Errorf("无法识别的配置类型,%#v", discoveryConfig)
	}
	if err != nil {
		return nil, fmt.Errorf("不能识别的 config 类型:%#v", discoveryConfig)
	}
	return &ServiceClient{
		client:      restClient,
		serviceInfo: serviceInfo,
		apiCallers:  make(map[string]*ApiCaller, 0),
	}, nil
}

type ServiceClient struct {
	client      *resty.Client
	serviceInfo base.ServiceInfo
	apiCallers  map[string]*ApiCaller
}

func (this *ServiceClient) GetServiceName() string {
	return this.serviceInfo.GetServiceName()
}

func (this *ServiceClient) GetRestClient() *resty.Client {
	return this.client
}

func (this *ServiceClient) ApiRegister(command string, endpointMeta base.EndPointMeta) error {
	if this.apiCallers[command] == nil {
		apiCaller := &ApiCaller{
			Command:      command,
			EndpointMeta: endpointMeta,
		}
		this.apiCallers[command] = apiCaller
		return nil
	}
	return fmt.Errorf("command[%s]已经存在,不能再次注册", command)

}

func (this *ServiceClient) SyncCallApiExt(command string, pathParam map[string]string, query url.Values, body RequestBody, result interface{}) *base.Error {
	resp, err := this.SyncCallApi(command, pathParam, query, body)
	if err != nil {
		return base.NewSimpleError(-1, fmt.Sprintf("%s", err.Error()))
	}
	if resp.StatusCode() >= 400 {
		response := &base.ErrorResponse{}
		json.Unmarshal(resp.Body(), response)
		if response.Errors != nil {
			return response.Errors
		}
		return base.NewSimpleError(-1, fmt.Sprintf("%s", resp.Body()))
	}
	if result == nil {
		return nil
	}
	contentType := resp.Header().Get("Content-Type")
	switch {
	case strings.HasPrefix(contentType, "application/json"):
		return base.ErrorToResponseError(json.Unmarshal(resp.Body(), result))
	case strings.HasPrefix(contentType, "text/xml"):
		return base.ErrorToResponseError(xml.Unmarshal(resp.Body(), result))
	case strings.HasPrefix(contentType, "application/x-protobuf"):
		if message, ok := result.(proto.Message); ok {
			return base.ErrorToResponseError(proto.Unmarshal(resp.Body(), message))
		}
		fallthrough
	default:
		return base.NewSimpleError(-1, "can't decode response data")
	}
}

func (this *ServiceClient) SyncCallApi(command string, pathParam map[string]string, query url.Values, body RequestBody) (*resty.Response, error) {
	caller, ok := this.apiCallers[command]
	if !ok {
		return nil, fmt.Errorf("没有注册过cmmand[%s]", command)
	}
	var response *resty.Response
	err := hystrix.Do(command, func() error {
		var err1 error
		response, err1 = doCommand(this.client, caller, query, body, pathParam)
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

func (this *ServiceClient) AsyncCallApi(command string, pathParam map[string]string, query url.Values, body RequestBody) (chan<- *resty.Response, error) {
	caller, ok := this.apiCallers[command]
	if !ok {
		return nil, fmt.Errorf("没有注册过cmmand[%s]", command)
	}
	response := make(chan *resty.Response)
	err := hystrix.Go(command, func() error {
		var err1 error
		res, err1 := doCommand(this.client, caller, query, body, pathParam)
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

func doCommand(client *resty.Client, caller *ApiCaller, query url.Values, body RequestBody, pathParam map[string]string) (*resty.Response, error) {
	request := client.R()
	request.QueryParam = query
	if body != nil {
		body.SetBody(request)
	}
	endPointMeta := caller.EndpointMeta
	uri := endPointMeta.Path
	if pathParam == nil || len(pathParam) == 0 {
		return request.Execute(string(endPointMeta.Method), uri)
	}
	uri = WarpUrl(uri, pathParam)
	return request.Execute(string(endPointMeta.Method), uri)
}
