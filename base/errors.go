package base

import (
	"fmt"

	"github.com/pquerna/ffjson/ffjson"
)

// Error 基础的错误接口
type Error interface {
	error
	GetErrorCode() int64
	GetScopes() string
	GetRootError() error
}

type strError string

func (m strError) Error() string {
	return string(m)
}

//BaseError Error 接口的实现,可 json 序列化
type baseError struct {
	Scope     string `json:"scope"`
	DebugCode int64  `json:"debug_code"`
	RootError error  `json:"root_error"`
	Message   string `json:"message"`
}

func (err *baseError) Error() string {
	return err.Message
}

func (err *baseError) GetErrorCode() int64 {
	return err.DebugCode
}
func (err *baseError) GetScopes() string {
	return err.Scope
}

func (err *baseError) GetRootError() error {
	return err.RootError
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
	data, _ := ffjson.Marshal(err.GetRootError())
	return fmt.Sprintf(`{"scope":"%s","debug_code":"%d","root_error":{%s},"message":"%s"}`, err.GetScopes(), err.GetErrorCode(), data, err.Error())
}

//NewError 构建一个新的 Error
func NewError(debugCode int64, scope string, errMsg string) Error {
	return &baseError{
		Scope:     scope,
		RootError: strError(errMsg),
		DebugCode: debugCode,
		Message:   errMsg,
	}
}

//NewErrorWrapper 创建一个对普通的 error的封装
func NewErrorWrapper(debugCode int64, scope string, err error) Error {
	return &baseError{Scope: scope, DebugCode: ErrCodeBaseSystemUnknown, RootError: err, Message: err.Error()}
}
