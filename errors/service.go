package errors

import (
	"fmt"

	"go.uber.org/zap"
)

type Service interface {
	// 创建一个子ErrorService
	NewService(childScope string) Service
	GetLogger() *zap.Logger
	Error(errorCode int32, message string, fields ...zap.Field) Error
	SystemError(message string, fields ...zap.Field) Error
	MessageError(message string, fields ...zap.Field) Error
	MessageErrorIgnoreLog(message string, fields ...zap.Field) Error
	WrappedError(errorCode int32, err error, fields ...zap.Field) Error
	WrappedSystemError(err error, fields ...zap.Field) Error
	WrappedMessageError(err error, fields ...zap.Field) Error
	WrappedMessageErrorIgnoreLog(err error, fields ...zap.Field) Error
	SystemErrorIgnoreLog(message string, fields ...zap.Field) Error
	WrappedSystemErrorIgnoreLog(err error, fields ...zap.Field) Error
}

func NewService(scope string, logger *zap.Logger) Service {
	if scope == "" {
		panic("没有指定error scope")
	}

	return &serviceImpl{
		rootLogger: logger,
		logger:     logger.WithOptions(zap.AddCaller(), zap.AddCallerSkip(1)),
		scope:      scope,
		child:      make(map[string]Service, 0),
	}
}

type serviceImpl struct {
	rootLogger *zap.Logger
	logger     *zap.Logger
	scope      string
	child      map[string]Service
}

func (impl *serviceImpl) GetLogger() *zap.Logger {
	return impl.rootLogger
}

func (impl *serviceImpl) NewService(childScope string) Service {
	if service, ok := impl.child[childScope]; ok {
		return service
	}
	service := NewService(fmt.Sprintf("%s.%s", impl.scope, childScope), impl.rootLogger)
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
	impl.logger.DPanic(message, fields...)
	return impl.Error(Error_System, message, fields...)
}

func (impl *serviceImpl) SystemErrorIgnoreLog(message string, fields ...zap.Field) Error {
	return impl.Error(Error_System, message, fields...)
}

func (impl *serviceImpl) MessageErrorIgnoreLog(message string, fields ...zap.Field) Error {
	return impl.Error(Error_Message, message, fields...)
}

func (impl *serviceImpl) MessageError(message string, fields ...zap.Field) Error {
	impl.logger.Error(message, fields...)
	return impl.Error(Error_Message, message, fields...)
}

func (impl *serviceImpl) WrappedError(errorCode int32, err error, fields ...zap.Field) Error {
	return &baseError{
		Scope:   impl.scope,
		Code:    errorCode,
		Message: err.Error(),
		Fields:  fields,
	}
}

func (impl *serviceImpl) WrappedSystemError(err error, fields ...zap.Field) Error {
	impl.logger.DPanic(err.Error(), fields...)
	return impl.WrappedError(Error_System, err, fields...)
}

func (impl *serviceImpl) WrappedSystemErrorIgnoreLog(err error, fields ...zap.Field) Error {
	return impl.WrappedError(Error_System, err, fields...)
}

func (impl *serviceImpl) WrappedMessageErrorIgnoreLog(err error, fields ...zap.Field) Error {
	return impl.WrappedError(Error_Message, err, fields...)
}

func (impl *serviceImpl) WrappedMessageError(err error, fields ...zap.Field) Error {
	impl.logger.Error(err.Error(), fields...)
	return impl.WrappedError(Error_Message, err, fields...)
}
