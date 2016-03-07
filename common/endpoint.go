package common

import "github.com/coffeehc/web"

type EndPoint struct {
	Path        string
	Method      web.HttpMethod
	Description string
	HandlerFunc web.RequestHandler
}
