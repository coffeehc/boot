package errors

import (
	"fmt"

	"go.uber.org/zap"
)

type Service interface {
	//创建一个子ErrorService
	NewService(childScope string) Service
	Error(errorCode int32, message string, fields ...zap.Field) Error
	SystemError(message string, fields ...zap.Field) Error
	MessageError(message string, fields ...zap.Field) Error
	WappedError(errorCode int32, err error, fields ...zap.Field) Error
	WappedSystemError(err error, fields ...zap.Field) Error
	WappedMessageError(err error, fields ...zap.Field) Error
}

func NewService(scope string) Service {
	if scope == "" {
		panic("没有指定error scope")
	}
	return &serviceImpl{
		scope: scope,
		child: make(map[string]Service, 0),
	}
}

type serviceImpl struct {
	scope string
	child map[string]Service
}

func (impl *serviceImpl) NewService(childScope string) Service {
	if service, ok := impl.child[childScope]; ok {
		return service
	}
	service := NewService(fmt.Sprintf("%s.%s", impl.scope, childScope))
	impl.child[childScope] = service
	return service
}

func (impl *serviceImpl) Error(errorCode int32, message string, fields ...zap.Field) Error {
	return &baseError{
		Scope:   impl.scope,
		Code:    errorCode,
		Message: message,
		Fields:  fields,
	}
}
func (impl *serviceImpl) SystemError(message string, fields ...zap.Field) Error {
	return impl.Error(Error_System, message, fields...)
}
func (impl *serviceImpl) MessageError(message string, fields ...zap.Field) Error {
	return impl.Error(Error_Message, message, fields...)
}

func (impl *serviceImpl) WappedError(errorCode int32, err error, fields ...zap.Field) Error {
	return &baseError{
		Scope:   impl.scope,
		Code:    errorCode,
		Message: err.Error(),
		Fields:  fields,
	}
}

func (impl *serviceImpl) WappedSystemError(err error, fields ...zap.Field) Error {
	return impl.WappedError(Error_System, err, fields...)
}
func (impl *serviceImpl) WappedMessageError(err error, fields ...zap.Field) Error {
	return impl.WappedError(Error_Message, err, fields...)
}
