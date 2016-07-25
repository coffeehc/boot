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
	"github.com/coffeehc/logger"
)

func NewServiceClient(serviceInfo base.ServiceInfo, discoveryConfig interface{}) (*ServiceClient, base.Error) {
	var restClient *resty.Client
	var err base.Error
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
		err = base.NewError(base.ERROR_CODE_BASE_INIT_ERROR,fmt.Sprintf("无法识别的配置类型,%#v", discoveryConfig))
	}
	if err != nil {
		return nil, base.NewError(base.ERROR_CODE_BASE_INIT_ERROR,fmt.Sprintf("不能识别的 config 类型:%#v", discoveryConfig))
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

func (this *ServiceClient) ApiRegister(command string, endpointMeta base.EndPointMeta)  base.Error {
	if this.apiCallers[command] == nil {
		apiCaller := &ApiCaller{
			Command:      command,
			EndpointMeta: endpointMeta,
		}
		this.apiCallers[command] = apiCaller
		return nil
	}
	return base.NewError(base.ERROR_CODE_BASE_API_COMMAND_REGISTERED,fmt.Sprintf("command[%s]已经存在,不能再次注册", command))

}

func (this *ServiceClient) SyncCallApiExt(command string, pathParam map[string]string, query url.Values, body RequestBody, result interface{}) base.Error {
	resp, err := this.SyncCallApi(command, pathParam, query, body)
	if err != nil {
		return err
	}
	if resp.StatusCode() >= 400 {
		response := &base.ErrorResponse{}
		err := json.Unmarshal(resp.Body(), response)
		if err != nil {
			return base.NewError(base.ERROR_CODE_BASE_DECODE_ERROR, err.Error())
		}
		return response
	}
	//TODO 300+的处理需要考虑
	if result == nil {
		return nil
	}
	var parseError error
	contentType := resp.Header().Get("Content-Type")
	switch {
	case strings.HasPrefix(contentType, "application/json"):
		parseError = json.Unmarshal(resp.Body(), result)
	case strings.HasPrefix(contentType, "text/xml"):
		parseError = xml.Unmarshal(resp.Body(), result)
	case strings.HasPrefix(contentType, "application/x-protobuf"):
		if message, ok := result.(proto.Message); ok {
			parseError = proto.Unmarshal(resp.Body(), message)
			break
		}
		fallthrough
	default:
		return base.NewError(base.ERROR_CODE_BASE_API_NOSUPPORT_CONTENTTYPE, fmt.Sprintf("content type %s no support",contentType))
	}
	if parseError != nil {
		return base.NewError(base.ERROR_CODE_BASE_DECODE_ERROR, err.Error())
	}
	return nil
}

func (this *ServiceClient) SyncCallApi(command string, pathParam map[string]string, query url.Values, body RequestBody) (*resty.Response, base.Error) {
	caller, ok := this.apiCallers[command]
	if !ok {
		return nil, base.NewError(base.ERROR_CODE_BASE_API_COMMAND_NOREGISTER,fmt.Sprintf("没有注册过cmmand[%s]", command))
	}
	var response *resty.Response
	err := hystrix.Do(command, func() error {
		var reqErr base.Error
		response, reqErr = doCommand(this.client, caller, query, body, pathParam)
		return reqErr
	}, func(err error) error {
		logger.Error("请求异常:%#v",err)
		//TODO 处理异常
		return err
	})
	if err != nil {
		if e,ok:=err.(base.Error);ok{
			return nil,e
		}
		return nil, base.NewError(base.ERROR_CODE_BASE_SYSTEM_ERROR,err.Error())
	}
	return response, nil
}


func doCommand(client *resty.Client, caller *ApiCaller, query url.Values, body RequestBody, pathParam map[string]string) (*resty.Response, base.Error) {
	request := client.R()
	request.QueryParam = query
	if body != nil {
		body.SetBody(request)
	}
	endPointMeta := caller.EndpointMeta
	uri := endPointMeta.Path
	if pathParam == nil || len(pathParam) == 0 {
		return httpExecute(request,string(endPointMeta.Method),uri)
	}
	uri = WarpUrl(uri, pathParam)
	return httpExecute(request,string(endPointMeta.Method),uri)
}

func httpExecute(request *resty.Request,method,uri string)(*resty.Response, base.Error){
	resp,err :=request.Execute(method, uri)
	if err!=nil{
		return nil,base.NewError(base.ERROR_CODE_BASE_API_REQUEST_BAD,err.Error())
	}
	return resp,nil
}
