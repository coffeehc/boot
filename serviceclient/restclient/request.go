package restclient

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/coffeehc/microserviceboot/base"
	"github.com/coffeehc/microserviceboot/base/restbase"
	"github.com/golang/protobuf/proto"
)

const err_scope_rest_request = "rest request"

type Request interface {
	GetCommand() string
	Init(baseUrl string, endpointMeta *restbase.EndpointMeta) base.Error
	SetPathParam(map[string]string) base.Error
	SetQueryParam(values url.Values) base.Error
	EncodeBody(data interface{}, handler RequestBodyEncoder) base.Error
	GetHeader() http.Header
	//Release() //用于复用对象
}

func AcquireRequest() Request {
	return &_Request{
		request: &http.Request{},
	}
}

type _Request struct {
	command string
	request *http.Request
}

func (this *_Request) GetCommand() string {
	return this.command
}

func (this *_Request) Init(baseUrl string, endpointMeta *restbase.EndpointMeta) base.Error {
	_url, err := url.Parse(baseUrl + endpointMeta.Path)
	if err != nil {
		return base.NewErrorWrapper(err_scope_rest_request, err)
	}
	_url.Opaque = "rest"
	this.request.URL = _url
	this.request.Header.Set("User-Agent", "coffee client")
	this.request.Host = _url.Host
	this.command = _url.Host + "." + _url.Path
	return nil

}
func (this *_Request) SetPathParam(pathParams map[string]string) base.Error {
	restUri := this.request.URL.Path
	for k, v := range pathParams {
		restUri = strings.Replace(restUri, "{"+k+"}", v, -1)
	}
	this.request.URL.Path = restUri
	return nil
}
func (this *_Request) SetQueryParam(values url.Values) base.Error {
	this.request.URL.RawQuery = values.Encode()
	return nil
}
func (this *_Request) EncodeBody(data interface{}, handler RequestBodyEncoder) base.Error {
	return handler(data, this.request)
}
func (this *_Request) GetHeader() http.Header {
	return this.request.Header
}

type RequestBodyEncoder func(data interface{}, request *http.Request) base.Error

func RequestFormBodyEncoder(data interface{}, request *http.Request) base.Error {
	values, ok := data.(url.Values)
	if !ok {
		return base.NewError(base.ErrCodeBaseSystemTypeConversion, err_scope_rest_request, "request body 参数不是 url.Values类型")
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Body = ioutil.NopCloser(strings.NewReader(values.Encode()))
	return nil
}

func RequestJsonBodyEncoder(data interface{}, request *http.Request) base.Error {
	request.Header.Set("Content-Type", "application/json")
	d, err := json.Marshal(data)
	if err != nil {
		return base.NewError(base.ErrCodeBaseSystemMarshal, err_scope_rest_request, err.Error())
	}
	request.Body = ioutil.NopCloser(bytes.NewReader(d))
	return nil
}

func RequestPBBodyEncoder(data interface{}, request *http.Request) base.Error {
	message, ok := data.(proto.Message)
	if !ok {
		return base.NewError(-1, err_scope_rest_request, "data not implement proto.Message")
	}
	request.Header.Set("Content-Type", "application/x-protobuf")
	d, err := proto.Marshal(message)
	if err != nil {
		return base.NewError(base.ErrCodeBaseSystemMarshal, err_scope_rest_request, err.Error())
	}
	request.Body = ioutil.NopCloser(bytes.NewReader(d))
	return nil
}
