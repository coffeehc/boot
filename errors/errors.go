package errors

import (
	"fmt"
	"strconv"

	"git.xiagaogao.com/coffee/boot/logs"
	"github.com/pquerna/ffjson/ffjson"
	"go.uber.org/zap"
)

// Error 基础的错误接口
type Error interface {
	error
	GetCode() int32
	GetScopes() string
	GetFields() []zap.Field
	GetFieldsWithCause() []zap.Field
	AddFields(...zap.Field)
	FormatRPCError() string
}

//BaseError Error 接口的实现,可 json 序列化
type baseError struct {
	Scope   string      `json:"scope"`
	Code    int32       `json:"code"`
	Message string      `json:"msg"`
	Fields  []zap.Field `json:"fields,omitempty"`
}

func (err *baseError) FormatRPCError() string {
	return fmt.Sprintf(`{"scope":"rpc","code":%d,"msg":"%s"}`, err.Code, err.Message)
}

func (err *baseError) AddFields(fields ...zap.Field) {
	err.Fields = append(err.Fields, fields...)
}

func (err *baseError) Error() string {
	return err.Message
}

func (err *baseError) GetCode() int32 {
	return err.Code
}
func (err *baseError) GetScopes() string {
	return err.Scope
}

func (err *baseError) GetFields() []zap.Field {
	return append(err.Fields, zap.String(logs.K_ServiceScope, err.GetScopes()), zap.String(logs.K_ErrorCode, strconv.FormatInt(int64(err.GetCode()), 16)))
}
func (err *baseError) GetFieldsWithCause() []zap.Field {
	return append(err.Fields, zap.String(logs.K_ServiceScope, err.GetScopes()), zap.String(logs.K_ErrorCode, strconv.FormatInt(int64(err.GetCode()), 16)), zap.String(logs.K_Cause, err.Message))
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
