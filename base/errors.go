package base

import (
	"fmt"

	"github.com/pquerna/ffjson/ffjson"
)

// Error 基础的错误接口
type Error interface {
	error
	GetCode() int32
	GetScopes() string
}

//BaseError Error 接口的实现,可 json 序列化
type baseError struct {
	Scope     string `json:"scope"`
	DebugCode int32  `json:"debug_code"`
	Message   string `json:"message"`
}

func (err *baseError) Error() string {
	return err.Message
}

func (err *baseError) GetCode() int32 {
	return err.DebugCode
}
func (err *baseError) GetScopes() string {
	return err.Scope
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

func ErrorToJson(err Error) string {
	data, _ := ffjson.Marshal(err)
	return string(data)
}

//NewError 构建一个新的 Error
func NewError(debugCode int32, scope string, errMsg string) Error {
	return &baseError{
		Scope:     scope,
		DebugCode: debugCode,
		Message:   errMsg,
	}
}

//NewErrorWrapper 创建一个对普通的 error的封装
func NewErrorWrapper(debugCode int32, scope string, err error) Error {
	if _err, ok := err.(Error); ok {
		scope = fmt.Sprintf("%s-%s", scope, _err.GetScopes())
		debugCode = debugCode | _err.GetCode()
	}
	return &baseError{Scope: scope, DebugCode: debugCode, Message: err.Error()}
}
