package common

import "github.com/coffeehc/web"

type EndPoint struct {
	Path        string
	Method      string
	Description string
	HandlerFunc web.RequestHandler
}
