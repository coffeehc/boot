package client

import (
	"github.com/coffeehc/resty"
)

type ApiRequest func(request *resty.Request, query map[string]string, body interface{}) (*resty.Response, error)

type ApiCaller struct {
	command    string
	apiRequest ApiRequest
}

func (this *ApiCaller) GetCommand() string {
	return this.command
}
