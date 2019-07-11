package errors

import (
	"fmt"
	"strings"

	"github.com/pquerna/ffjson/ffjson"
	"go.uber.org/zap"
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
		return ConverError(e, errorService)
	}
	errorService.GetLogger().DPanic("未知异常", zap.Any("err", err))
	return errorService.SystemErrorIgnoreLog(fmt.Sprintf("%#v", err))
}

func ConverError(err error, errorService Service) Error {
	if IsBaseError(err) {
		return err.(Error)
	}
	if strings.HasPrefix(err.Error(), "context ") {
		errorService.GetLogger().WithOptions(zap.AddCallerSkip(1)).Error(err.Error())
		return errorService.SystemErrorIgnoreLog(err.Error())
	}
	errorService.GetLogger().DPanic(err.Error())
	return errorService.WrappedSystemErrorIgnoreLog(err)
}
