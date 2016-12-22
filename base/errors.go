package base

import (
	"github.com/pquerna/ffjson/ffjson"
	"net/http"
)

type Error interface {
	error
	GetErrorCode() int64
	Scopes() string
	Message() string
}

type BaseError struct {
	Scope     string `json:"scope"`
	DebugCode int64  `json:"debug_code"`
	Msg       string `json:"msg"`
}

func (err *BaseError) Error() string {
	data, _ := ffjson.Marshal(err)
	return string(data)
}

func (err *BaseError) GetErrorCode() int64 {
	return err.DebugCode
}
func (err *BaseError) Scopes() string {
	return err.Scope
}

func (err *BaseError)Message() string{
	return err.Msg
}

func NewError(debugCode int64, scope string, errMsg string) Error {
	return &BaseError{
		Scope:     scope,
		Msg:       errMsg,
		DebugCode: debugCode,
	}
}

type _ErrorWrapper struct {
	BaseError
	Err error `json:"err"`
}

func NewErrorWrapper(scope string, err error) Error {
	return &_ErrorWrapper{BaseError: BaseError{Scope: scope, DebugCode: ERRCODE_BASE_SYSTEM_UNKNOWN, Msg: err.Error()}, Err: err}
}

type ErrorResponse struct {
	BaseError
	HttpCode        int    `json:"http_code"`
	InformationLink string `json:"information_link"`
}

func (err ErrorResponse) GetHttpCode() int {
	if err.HttpCode == 0 {
		return http.StatusBadRequest
	}
	return err.HttpCode
}

func NewErrorResponse(httpCode int, errorCode int64, message, informationLink string) *ErrorResponse {
	return &ErrorResponse{BaseError: BaseError{Scope: "response", DebugCode: errorCode, Msg: message}, HttpCode: httpCode, InformationLink: informationLink}
}
