package base

import "github.com/coffeehc/web"

type EndPointMeta struct {
	Path        string         `json:'path'`
	Method      web.HttpMethod `json:'method'`
	Description string         `json:"description"`
}

type EndPoint struct {
	Metadata    EndPointMeta
	HandlerFunc web.RequestHandler
}
