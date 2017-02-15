package restclient

import (
	"fmt"
	
	"context"
	
	"github.com/coffeehc/microserviceboot/base"
	"github.com/coffeehc/commons/httpcommons/client"
	"github.com/coffeehc/microserviceboot/loadbalancer"
	"github.com/coffeehc/microserviceboot/consultool"
	"github.com/hashicorp/consul/api"
	"github.com/coffeehc/microserviceboot/base/restbase"
	"io/ioutil"
)

type ServiceClient interface {
	GetBaseUrl() string
}

func NewServiceClient(serviceInfo base.ServiceInfo, httpClientConfig *client.HTTPClientOptions, discoveryConfig interface{}) (ServiceClient, base.Error) {
	if serviceInfo == nil {
		return nil, base.NewError(base.ErrCodeBaseSystemNil, "rest client", "serviceInfo is nil")
	}
	if discoveryConfig == nil {
		return nil, base.NewError(base.ErrCodeBaseSystemNil, "rest client", "discoveryConfig is nil")
	}
	if httpClientConfig == nil {
		httpClientConfig = *client.HTTPClientOptions{}
	}
	rootCxt := context.Background()
	var balancer loadbalancer.Balancer
	var baseURL string
	var err base.Error
	switch c := discoveryConfig.(type) {
	case string: //host
		if c == "" {
			return nil, base.NewError(base.ErrCodeBaseSystemNil, "rest client", "discoveryConfig is a addrs")
		}
		balancer, err = loadbalancer.NewAddrArrayBalancer([]string{c})
		if err != nil {
			return nil, base.NewErrorWrapper("rest client", err)
		}
		baseURL = fmt.Sprintf("%s://%s", serviceInfo.GetScheme(), c)
	case *api.Client:
		balancer, err = consultool.NewConsulBalancer(rootCxt, c, serviceInfo)
		if err != nil {
			return nil, err
		}
		baseURL = fmt.Sprintf("%s://%s.%s.service", serviceInfo.GetScheme(), serviceInfo.GetServiceTag(), serviceInfo.GetServiceName())
	}
	restClient := newHttpClient(rootCxt, serviceInfo, balancer, httpClientConfig)
	return &_ServiceClient{
		client:      restClient,
		serviceInfo: serviceInfo,
		baseURL:baseURL,
	}, nil
}

type _ServiceClient struct {
	client      client.HTTPClient
	serviceInfo base.ServiceInfo
	baseURL     string
}

func (sc *_ServiceClient) GetServiceName() string {
	return sc.serviceInfo.GetServiceName()
}

func (sc *_ServiceClient) GetBaseUrl() string {
	return sc.baseURL
}

func (sc *_ServiceClient) GetHttpClient() client.HTTPClient {
	return sc.client
}

func (sc *_ServiceClient)Do(endpintMeta restbase.EndpointMeta, requestData, responstData interface{}, contentType ContentType) base.Error {
	if contentType == nil {
		contentType = JsonContentType
	}
	req := client.NewHTTPRequest()
	method := string(endpintMeta.Method)
	req.SetMethod(method)
	uri := fmt.Sprintf("%s/%s", sc.baseURL, endpintMeta.Path)
	if method != "POST" && method != "PUT" && method != "PATCH" && method != "OPTIONS" {
		if query, ok := requestData.(string); ok {
			uri = fmt.Sprintf("%s?%s", uri, query)
		} else {
			return base.NewError(base.ErrCodeBaseRPCInvalidArgument, "rest client", "requestData is not query")
		}
	} else {
		if requestData != nil {
			req.SetContentType(contentType.GetContenType())
			data, err := contentType.Encode(requestData)
			if err != nil {
				return base.NewErrorWrapper("rest client", err)
			}
			req.SetBody(data)
		}
	}
	req.SetURI(uri)
	response,err:=sc.client.Do(req,false)
	if err!=nil{
		return base.NewErrorWrapper("rest client",err)
	}
	//response.GetContentType() == ""
	body :=response.GetBody()
	defer body.Close()
	if response.SetStatusCode() != 200{
		return base.NewError(base.ErrCodeBaseRPCAborted,"rest client",fmt.Sprintf("response code is ",response.SetStatusCode()))
	}
	data ,err:= ioutil.ReadAll(body)
	if err!=nil{
		return base.NewErrorWrapper("rest client", err)
	}
	err = contentType.Decoder(data,responstData)
	if err!=nil{
		return base.NewErrorWrapper("rest client", err)
	}
	return nil
}


