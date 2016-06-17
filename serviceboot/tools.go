package serviceboot

import (
	"fmt"
	"net/http"

	"github.com/coffeehc/microserviceboot/base"
	"github.com/coffeehc/web"
)

func ErrorRecover(reply web.Reply) {
	if err := recover(); err != nil {
		switch e := err.(type) {
		case base.Error:
			reply.With(base.NewErrorResponse(&e)).
				SetStatusCode(http.StatusBadRequest)
		case *base.Error:
			reply.With(base.NewErrorResponse(e)).
				SetStatusCode(http.StatusBadRequest)
		case base.BizErr:
			reply.With(e.ToError()).
				SetStatusCode(int(e.GetHttpCode()))
		case *base.BizErr:
			reply.With(e.ToError()).
				SetStatusCode(int(e.GetHttpCode()))
		case string:
			reply.With(base.NewErrorResponse(base.NewSimpleError(http.StatusBadRequest, e))).
				SetStatusCode(http.StatusBadRequest)
		case error:
			reply.With(base.NewErrorResponse(base.NewSimpleError(http.StatusInternalServerError, e.Error()))).
				SetStatusCode(http.StatusInternalServerError)
		default:
			reply.With(base.NewErrorResponse(base.NewSimpleError(http.StatusInternalServerError, fmt.Sprintf("%#v", err)))).
				SetStatusCode(http.StatusInternalServerError)
		}
		//暂时统一按照400处理
	}
}
