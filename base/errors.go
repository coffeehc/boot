package base

import (
	"fmt"
	"net/http"

	"github.com/coffeehc/web"
)

type BizErr struct {
	httpCode  int64
	debugCode int64
	msg       string
}

func (this BizErr) Error() string {
	return this.msg
}

func (this BizErr) GetHttpCode() int64 {
	if this.httpCode == 0 {
		return http.StatusBadRequest
	}
	return this.httpCode
}

func (this BizErr) GetDebugCode() int64 {
	return this.debugCode
}

func (this BizErr) ToError() Error {
	return Error{
		Code:     int32(this.GetHttpCode()),
		Debug_id: int64(this.GetDebugCode()),
		Message:  this.Error(),
	}
}

//默认第一个为 httpCode, 第二个为debugCode
func BuildBizErr(err string, codes ...int64) BizErr {
	var httpCode, debugCode int64
	if len(codes) > 0 {
		httpCode = codes[0]
	}
	if len(codes) > 1 {
		debugCode = codes[1]
	}
	return BizErr{
		msg:       err,
		httpCode:  httpCode,
		debugCode: debugCode,
	}
}

type Error struct {
	Code             int32  `json:"code"`
	Debug_id         int64  `json:"debug_id"`
	Message          string `json:"message"`
	Information_link string `json:"information_link"`
}

func ErrorToResponseError(err error) *Error {
	if err == nil {
		return nil
	}
	return NewSimpleError(-1, err.Error())
}

func (this Error) Error() string {
	return fmt.Sprintf("%d:%s", this.Code, this.Message)
}

func NewSimpleError(code int32, message string) *Error {
	return &Error{Code: code, Message: message}
}

type ErrorResponse struct {
	Errors *Error `json:"error"`
}

func NewErrorResponse(errs *Error) ErrorResponse {
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
