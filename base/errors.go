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

func (this _ErrorWrapper) Error() string {
	return this.err.Error()
}
func (this _ErrorWrapper) GetErrorCode() int64 {
	return 0x500
}

type BizErr struct {
	debugCode int64
	msg       string
}

func (this BizErr) Error() string {
	return this.msg
}

func (this BizErr) GetErrorCode() int64 {
	return this.debugCode
}

//默认第一个为 httpCode, 第二个为debugCode
func NewBizErr(debugCode int64, errMsg string) *BizErr {
	return &BizErr{
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

func (this ErrorResponse) Error() string {
	return fmt.Sprintf("%d:%d:%s", this.HttpCode, this.ErrorCode, this.Message)
}

func (this ErrorResponse) GetErrorCode() int64 {
	return this.ErrorCode
}

func (this ErrorResponse) GetHttpCode() int {
	if this.HttpCode == 0 {
		return http.StatusBadRequest
	}
	return this.HttpCode
}

func NewErrorResponse(httpCode int, errorCode int64, message, informationLink string) *ErrorResponse {
	return &ErrorResponse{HttpCode: httpCode, ErrorCode: errorCode, Message: message, InformationLink: informationLink}
}
