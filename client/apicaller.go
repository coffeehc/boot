package client

import (
	"github.com/coffeehc/resty"
)

type RequestMethod string

const (
	RequestMethod_GET     = RequestMethod("GET")
	RequestMethod_POST    = RequestMethod("POST")
	RequestMethod_PUT     = RequestMethod("PUT")
	RequestMethod_DELETE  = RequestMethod("DELETE")
	RequestMethod_PATCH   = RequestMethod("PATCH")
	RequestMethod_HEAD    = RequestMethod("HEAD")
	RequestMethod_OPTIONS = RequestMethod("OPTIONS")
)

type ApiRequestSetting func(request *resty.Request)

type ApiCaller struct {
	command           string
	apiRequestSetting ApiRequestSetting
	method            RequestMethod
	uri               string
}

func (this *ApiCaller) GetCommand() string {
	return this.command
}
