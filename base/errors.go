package base

import (
	"net/http"

	"github.com/coffeehc/commons/convers"
	"github.com/pquerna/ffjson/ffjson"
)

// Error 基础的错误接口
type Error interface {
	error
	GetErrorCode() int64
	Scopes() string
	Message() string
}

//BaseError Error 接口的实现,可 json 序列化
type baseError struct {
	Scope     string `json:"scope"`
	DebugCode int64  `json:"debug_code"`
	Msg       string `json:"msg"`
}

func (err *baseError) Error() string {
	data, _ := ffjson.Marshal(err)
	return convers.BytesToString(data)
}

func (err *baseError) GetErrorCode() int64 {
	return err.DebugCode
}
func (err *baseError) Scopes() string {
	return err.Scope
}

func (err *baseError) Message() string {
	return err.Msg
}

//ParseErrorFromJSON 从 Jons数据解析出 Error 对象
func ParseErrorFromJSON(data []byte) Error {
	err := &baseError{}
	e := ffjson.Unmarshal(data, err)
	if e != nil {
		return nil
	}
	return err
}

//NewError 构建一个新的 Error
func NewError(debugCode int64, scope string, errMsg string) Error {
	return &baseError{
		Scope:     scope,
		Msg:       errMsg,
		DebugCode: debugCode,
	}
}

type errorWrapper struct {
	baseError
	Err error `json:"err"`
}

//NewErrorWrapper 创建一个对普通的 error的封装
func NewErrorWrapper(scope string, err error) Error {
	return &errorWrapper{baseError: baseError{Scope: scope, DebugCode: ErrCodeBaseSystemUnknown, Msg: err.Error()}, Err: err}
}

//ErrorResponse 对http error的封装
type ErrorResponse interface {
	Error
	GetHTTPCode() int
}

type errorResponse struct {
	baseError
	HTTPCode        int    `json:"http_code"`
	InformationLink string `json:"information_link"`
}

func (err *errorResponse) GetHTTPCode() int {
	if err.HTTPCode == 0 {
		return http.StatusBadRequest
	}
	return err.HTTPCode
}

//NewErrorResponse 创建一个 Response 的 Error
func NewErrorResponse(httpCode int, errorCode int64, message, informationLink string) ErrorResponse {
	return &errorResponse{baseError: baseError{Scope: "response", DebugCode: errorCode, Msg: message}, HTTPCode: httpCode, InformationLink: informationLink}
}
