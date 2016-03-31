package base

import (
	"fmt"
	"net/http"

	"github.com/coffeehc/web"
)

type Error struct {
	Code             int32  `json:"code"`
	Debug_id         int64  `json:"debug_id"`
	Message          string `json:"message"`
	Information_link string `json:"information_link"`
}

func NewSimpleError(code int32, message string) Error {
	return Error{Code: code, Message: message}
}

type ErrorResponse struct {
	Errors []Error `json:"errors"`
}

func NewErrorResponse(errs ...Error) ErrorResponse {
	return ErrorResponse{errs}
}

func RegeditRestFilter(server *web.Server) {
	server.AddFirstFilter("/*", restFilter)
}

func restFilter(request *http.Request, reply web.Reply, chain web.FilterChain) {
	defer func() {
		if err := recover(); err != nil {
			var httpErr *web.HttpError
			var ok bool
			if httpErr, ok = err.(*web.HttpError); !ok {
				httpErr = web.HTTPERR_500(fmt.Sprintf("%#s", err))
			}
			reply.SetStatusCode(httpErr.Code)
		}
	}()

}
