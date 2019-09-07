package errors

import (
	"fmt"
	"strings"

	"github.com/json-iterator/go"
	"go.uber.org/zap"
)

func NamedScope(name string) zap.Field {
	return zap.String("scope", name)
}

func ParseError(jsonStr string) Error {
	err := &baseError{}
	e := jsoniter.Unmarshal([]byte(jsonStr), err)
	if e != nil {
		err.Code = ErrorSystemRPC
		err.Message = fmt.Sprintf("无法解析错误消息[%s],%#v", jsonStr, e)
	}
	return err
}

func ConverUnknowError(err interface{}) Error {
	if e, ok := err.(error); ok {
		return ConverError(e)
	}
	return SystemError("未知异常")
}

func ConverError(err error) Error {
	if IsBaseError(err) {
		return err.(Error)
	}
	if strings.HasPrefix(err.Error(), "context ") || strings.HasPrefix(err.Error(), "rpc error:") {
		return SystemError(err.Error())
	}
	return WrappedSystemError(err)
}
