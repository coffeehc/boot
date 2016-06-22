package serviceboot

import (
	"fmt"
	"net/http"

	"encoding/json"
	"github.com/coffeehc/microserviceboot/base"
	"github.com/coffeehc/web"
	"io/ioutil"
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
			reply.With(base.NewErrorResponse(e.ToError())).
				SetStatusCode(int(e.GetHttpCode()))
		case *base.BizErr:
			reply.With(base.NewErrorResponse(e.ToError())).
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

//将 request 的 Json内容解析为 对象
func UnmarshalWhitJson(request *http.Request, data interface{}) {
	dataBytes, err := ioutil.ReadAll(request.Body)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(dataBytes, data)
	if err != nil {
		panic(err)
	}
}

//如果 err 不为空,直接 Panic
func PanicErr(err error) {
	if err != nil {
		panic(err)
	}
}
