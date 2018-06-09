package grpcclient

import (
	"fmt"
	"sync"

	"git.xiagaogao.com/coffee/boot"
	"git.xiagaogao.com/coffee/boot/errors"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

const _internalInvoker = "_internal_invoker"
const context_serviceInfoKey = "__serviceInfo__"

var _unaryClientInterceptor *unartClientInterceptor

func wapperUnartClientInterceptor(serviceInfo boot.ServiceInfo, errorService errors.Service, logger *zap.Logger) grpc.UnaryClientInterceptor {
	_unaryClientInterceptor = newUnartClientInterceptor(errorService, logger)
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) (err error) {
		return _unaryClientInterceptor.Interceptor(context.WithValue(ctx, context_serviceInfoKey, serviceInfo), method, req, reply, cc, invoker, opts...)
	}
}

//AppendUnartClientInterceptor 追加一个UnartClientInterceptor
func AppendUnartClientInterceptor(name string, unaryClientInterceptor grpc.UnaryClientInterceptor) errors.Error {
	return _unaryClientInterceptor.AppendInterceptor(name, unaryClientInterceptor)
}

func newUnartClientInterceptor(errorService errors.Service, logger *zap.Logger) *unartClientInterceptor {
	errorService = errorService.NewService("grpc")
	return &unartClientInterceptor{
		interceptors: make(map[string]*unaryClientInterceptorWapper),
		rootInterceptor: &unaryClientInterceptorWapper{
			interceptor:  newPaincInterceptor(errorService, logger),
			errorService: errorService,
			logger:       logger,
		},
		mutex:        new(sync.Mutex),
		errorService: errorService,
		logger:       logger,
	}
}

type unartClientInterceptor struct {
	interceptors    map[string]*unaryClientInterceptorWapper
	rootInterceptor *unaryClientInterceptorWapper
	mutex           *sync.Mutex
	errorService    errors.Service
	logger          *zap.Logger
}

func (uci *unartClientInterceptor) Interceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) (err error) {
	opts = append(opts, grpc.FailFast(true))
	return uci.rootInterceptor.interceptor(context.WithValue(ctx, _internalInvoker, invoker), method, req, reply, cc, uci.rootInterceptor.invoker, opts...)
}

func (uci *unartClientInterceptor) AppendInterceptor(name string, interceptor grpc.UnaryClientInterceptor) errors.Error {
	uci.mutex.Lock()
	defer uci.mutex.Unlock()
	if _, ok := uci.interceptors[name]; ok {
		return uci.errorService.SystemError(fmt.Sprintf("%s 已经存在", name))
	}
	lastInterceptor := getLastUnaryClientInterceptor(uci.rootInterceptor)
	lastInterceptor.next = &unaryClientInterceptorWapper{interceptor: interceptor}
	uci.interceptors[name] = lastInterceptor.next
	return nil
}

func getLastUnaryClientInterceptor(root *unaryClientInterceptorWapper) *unaryClientInterceptorWapper {
	if root.next == nil {
		return root
	}
	return getLastUnaryClientInterceptor(root.next)
}

type unaryClientInterceptorWapper struct {
	interceptor  grpc.UnaryClientInterceptor
	next         *unaryClientInterceptorWapper
	errorService errors.Service
	logger       *zap.Logger
}

func (uciw *unaryClientInterceptorWapper) invoker(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) (err error) {
	if uciw.next == nil {
		realInvoker := ctx.Value(_internalInvoker)
		if realInvoker == nil {
			return uciw.errorService.SystemError("没有 Handler")
		}
		if invoker, ok := realInvoker.(grpc.UnaryInvoker); ok {
			return invoker(ctx, method, req, reply, cc, opts...)
		}
		return uciw.errorService.SystemError("类型错误")
	}
	return uciw.next.interceptor(ctx, method, req, reply, cc, uciw.next.invoker, opts...)
}

func newPaincInterceptor(errorService errors.Service, logger *zap.Logger) grpc.UnaryClientInterceptor {
	return func(cxt context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) (err error) {
		defer func() {
			if r := recover(); r != nil {
				err = adapteError(cxt, r, errorService)
			}
		}()
		return adapteError(cxt, invoker(cxt, method, req, reply, cc, opts...), errorService)
	}
}

func adapteError(ctx context.Context, err interface{}, errorService errors.Service) errors.Error {
	if err == nil {
		return nil
	}
	serviceName := boot.GetServiceName(ctx)
	if serviceName == "" {
		serviceName = "未知服务"
	}
	serviceName = "grpcserver:" + serviceName
	switch v := err.(type) {
	case errors.Error:
		return v
	case string:
		return errorService.SystemError(v)
	case error:
		if s, ok := status.FromError(v); ok {
			code := int32(s.Code())
			if !errors.IsBaseErrorCode(code) {
				return errorService.WappedSystemError(s.Err())
			}
			return errorService.Error(code, s.Message())
		}
		return errorService.WappedSystemError(v)
	default:
		return errorService.SystemError(fmt.Sprintf("%#v", v))
	}

}
