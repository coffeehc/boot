package errors

import (
	"fmt"

	"github.com/pquerna/ffjson/ffjson"
)

func ParseError(jsonStr string) Error {
	err := &baseError{}
	e := ffjson.Unmarshal([]byte(jsonStr), err)
	if e != nil {
		err.Code = Error_System_RPC
		err.Message = fmt.Sprintf("无法解析错误消息[%s],%#v", jsonStr, e)
	}
	return err
}

func ConverUnkonwError(err interface{}, errorService Service) Error {
	if e, ok := err.(error); ok {
		if IsBaseError(e) {
			return e.(Error)
		}
		return errorService.WappedSystemError(e)
	}
	return errorService.SystemError(fmt.Sprintf("%#v", err))
}

func ConverError(err error, errorService Service) Error {
	if IsBaseError(err) {
		return err.(Error)
	}
	return errorService.WappedSystemError(err)
}
