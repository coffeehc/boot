package base

import (
	"fmt"
	"net/http"
)

type Error interface {
	Error() string
	GetErrorCode() int64
}

type _ErrorWrapper struct {
	err error
}

func NewErrorWrapper(err error) Error {
	return &_ErrorWrapper{err: err}
}

func (err _ErrorWrapper) Error() string {
	return err.err.Error()
}
func (_ErrorWrapper) GetErrorCode() int64 {
	return ERROR_CODE_BASE_SYSTEM_ERROR
}

type BaseError struct {
	debugCode int64
	msg       string
}

func (err BaseError) Error() string {
	return err.msg
}

func (err BaseError) GetErrorCode() int64 {
	return err.debugCode
}

//默认第一个为 httpCode, 第二个为debugCode
func NewError(debugCode int64, errMsg string) Error {
	return &BaseError{
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

func (err ErrorResponse) Error() string {
	return fmt.Sprintf("%d:%d:%s", err.HttpCode, err.ErrorCode, err.Message)
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
