package serviceclient

import "github.com/coffeehc/microserviceboot/base"

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

type ApiCaller struct {
	Command      string
	EndpointMeta base.EndPointMeta
}
