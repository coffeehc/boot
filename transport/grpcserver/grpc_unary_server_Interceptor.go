package grpcserver

import (
	"fmt"
	"sync"

	"time"

	"runtime/debug"

	"git.xiagaogao.com/coffee/boot"
	"git.xiagaogao.com/coffee/boot/errors"
	"git.xiagaogao.com/coffee/boot/logs"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	spb "google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	_internalUnaryServerInfo = "_internal_UnaryServerInfo"
	_internalHandler         = "_internal_handler"
)

func newUnaryServerInterceptor(ctx context.Context, errorService errors.Service, logger *zap.Logger) *unaryServerInterceptor {
	errorService = errorService.NewService("grpc")
	return &unaryServerInterceptor{
		interceptors: make(map[string]*unaryServerInterceptorWapper),
		rootInterceptor: &unaryServerInterceptorWapper{
			interceptor: catchPanicInterceptor,
		},
		mutex:        new(sync.Mutex),
		logger:       logger,
		errorService: errorService,
	}
}

type unaryServerInterceptor struct {
	interceptors    map[string]*unaryServerInterceptorWapper
	rootInterceptor *unaryServerInterceptorWapper
	mutex           *sync.Mutex
	errorService    errors.Service
	logger          *zap.Logger
}

func (usi *unaryServerInterceptor) Interceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	ctx = logs.SetLogger(ctx, usi.logger)
	ctx = context.WithValue(ctx, _internalUnaryServerInfo, info)
	ctx = context.WithValue(ctx, _internalHandler, handler)
	return usi.rootInterceptor.interceptor(ctx, req, info, usi.rootInterceptor.handler)
}

func (usi *unaryServerInterceptor) AppendInterceptor(name string, interceptor grpc.UnaryServerInterceptor) errors.Error {
	usi.mutex.Lock()
	defer usi.mutex.Unlock()
	if _, ok := usi.interceptors[name]; ok {
		return usi.errorService.SystemError(fmt.Sprintf("%s 已经存在", name))
	}
	lastInterceptor := getLastUnaryServerInterceptor(usi.rootInterceptor)
	lastInterceptor.next = &unaryServerInterceptorWapper{interceptor: interceptor, errorService: usi.errorService, logger: usi.logger}
	usi.interceptors[name] = lastInterceptor.next
	return nil
}

func getLastUnaryServerInterceptor(root *unaryServerInterceptorWapper) *unaryServerInterceptorWapper {
	if root.next == nil {
		return root
	}
	return getLastUnaryServerInterceptor(root.next)
}

type unaryServerInterceptorWapper struct {
	interceptor  grpc.UnaryServerInterceptor
	next         *unaryServerInterceptorWapper
	errorService errors.Service
	logger       *zap.Logger
}

func (usiw *unaryServerInterceptorWapper) handler(ctx context.Context, req interface{}) (interface{}, error) {
	if usiw.next == nil {
		realHandler := ctx.Value(_internalHandler)
		if realHandler == nil {
			return nil, usiw.errorService.SystemError("没有 Handler")
		}
		if handler, ok := realHandler.(grpc.UnaryHandler); ok {
			return handler(ctx, req)
		}
		return nil, usiw.errorService.SystemError("类型错误")
	}
	info := ctx.Value(_internalUnaryServerInfo)
	if info == 0 {
		return nil, usiw.errorService.SystemError("没有 ServerInfo")
	}
	if unaryServerInfo, ok := info.(*grpc.UnaryServerInfo); ok {
		return usiw.next.interceptor(ctx, req, unaryServerInfo, usiw.next.handler)
	}
	return nil, usiw.errorService.SystemError("类型错误")
}

func catchPanicInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = adapteError(ctx, r)
		}
	}()
	resp, err = handler(ctx, req)
	if err != nil {
		return nil, adapteError(ctx, err)
	}
	return resp, nil
}

func adapteError(ctx context.Context, err interface{}) error {
	if err == nil {
		return nil
	}
	if boot.IsDevModule() {
		logger := logs.GetLogger(ctx)
		if e, ok := err.(errors.Error); ok && errors.IsMessageError(e) {
			logger.Error(e.Error(), e.GetFields()...)
		} else {
			logger.Error("rpc内部异常", zap.Any(logs.K_Cause, err))
		}
		time.Sleep(time.Millisecond * 100)
		debug.PrintStack()
	}
	switch v := err.(type) {
	case errors.Error:
		return status.ErrorProto(&spb.Status{
			Code:    v.GetCode(),
			Message: v.Error(),
		})
	case string:
		return status.Errorf(codes.Internal, v)
	case error:
		return status.Errorf(codes.Internal, v.Error())
	default:
		return status.Errorf(codes.Unknown, "%#v", v)
	}

}
