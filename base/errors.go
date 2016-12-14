package base

import (
	"fmt"
	"net/http"
)

type Error interface {
	Error() string
	GetErrorCode() int64
	Scope() string
}

type _ErrorWrapper struct {
	scope string
	err   error
}

func NewErrorWrapper(scope string, err error) Error {
	return &_ErrorWrapper{scope: scope, err: err}
}

func (err _ErrorWrapper) Scope() string {
	return err.scope
}

func (err _ErrorWrapper) Error() string {
	return fmt.Sprintf("[%s] %d:%s", err.scope, ERRCODE_BASE_SYSTEM_UNKNOWN, err.err.Error())
}
func (_ErrorWrapper) GetErrorCode() int64 {
	return ERRCODE_BASE_SYSTEM_UNKNOWN
}

type BaseError struct {
	scope     string
	debugCode int64
	msg       string
}

func (err BaseError) Scope() string {
	return err.scope
}

func (err BaseError) Error() string {
	return fmt.Sprintf("[%s] %d:%s", err.scope, err.debugCode, err.msg)
}

func (err BaseError) GetErrorCode() int64 {
	return err.debugCode
}

func NewError(debugCode int64, scope string, errMsg string) Error {
	return &BaseError{
		scope:     scope,
		msg:       errMsg,
		debugCode: debugCode,
	}
}

type ErrorResponse struct {
	HttpCode        int    `json:"http_code"`
	ErrorCode       int64  `json:"debug_code"`
	Message         string `json:"message"`
	InformationLink string `json:"information_link"`
}

func (err ErrorResponse) Scope() string {
	return "http.response"
}

func (err ErrorResponse) Error() string {
	return fmt.Sprintf("[http.response] %d:%d:%s", err.HttpCode, err.ErrorCode, err.Message)
}

func (err ErrorResponse) GetErrorCode() int64 {
	return err.ErrorCode
}

func (err ErrorResponse) GetHttpCode() int {
	if err.HttpCode == 0 {
		return http.StatusBadRequest
	}
	return err.HttpCode
}

func NewErrorResponse(httpCode int, errorCode int64, message, informationLink string) *ErrorResponse {
	return &ErrorResponse{HttpCode: httpCode, ErrorCode: errorCode, Message: message, InformationLink: informationLink}
}
