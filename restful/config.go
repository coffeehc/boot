package restful

import "git.xiagaogao.com/coffee/httpx"

//EndpointMeta endpoint meta define
type EndpointMeta struct {
	Path        string              `json:"path"`
	Method      httpx.RequestMethod `json:"method"`
	Description string              `json:"description"`
}

//Endpoint endpoint define
type Endpoint struct {
	Metadata    EndpointMeta
	HandlerFunc httpx.RequestHandler
}
