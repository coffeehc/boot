package errors

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

type Service interface {
	//创建一个子ErrorService
	NewService(childScope string) Service
	BuildError(errorCode int32, message string, fields ...zap.Field) Error
	BuildSystemError(message string, fields ...zap.Field) Error
	BuildMessageError(message string, fields ...zap.Field) Error
	BuildWappedError(errorCode int32, err error, fields ...zap.Field) Error
	BuildWappedSystemError(err error, fields ...zap.Field) Error
	BuildWappedMessageError(err error, fields ...zap.Field) Error
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

func (impl *serviceImpl) BuildError(errorCode int32, message string, fields ...zap.Field) Error {
	return NewError(errorCode, impl.scope, message, fields...)
}
func (impl *serviceImpl) BuildSystemError(message string, fields ...zap.Field) Error {
	return impl.BuildError(Error_System, message, fields...)
}
func (impl *serviceImpl) BuildMessageError(message string, fields ...zap.Field) Error {
	return impl.BuildError(Error_Message, message, fields...)
}

func (impl *serviceImpl) BuildWappedError(errorCode int32, err error, fields ...zap.Field) Error {
	return NewErrorWrapper(errorCode, impl.scope, err, fields...)
}

func (impl *serviceImpl) BuildWappedSystemError(err error, fields ...zap.Field) Error {
	return impl.BuildWappedError(Error_System, err, fields...)
}
func (impl *serviceImpl) BuildWappedMessageError(err error, fields ...zap.Field) Error {
	return impl.BuildWappedError(Error_Message, err, fields...)
}

const ctx_Service_Key = "_root_errorService"

func SetRootErrorService(ctx context.Context, errorService Service) context.Context {
	return context.WithValue(ctx, ctx_Service_Key, errorService)
}

func GetRootErrorService(ctx context.Context) Service {
	return ctx.Value(ctx_Service_Key).(Service)
}
