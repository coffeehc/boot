package grpcclient

import (
	"fmt"
	"sync"

	"git.xiagaogao.com/coffee/boot/errors"
	"git.xiagaogao.com/coffee/boot/logs"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const _internalInvoker = "_internal_invoker"

func wapperUnartClientInterceptor(ctx context.Context, errorService errors.Service, logger *zap.Logger) grpc.UnaryClientInterceptor {
	_unaryClientInterceptor := newUnartClientInterceptor(ctx, errorService, logger)
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) (err error) {
		return _unaryClientInterceptor.Interceptor(ctx, method, req, reply, cc, invoker, opts...)
	}
}

func newUnartClientInterceptor(ctx context.Context, errorService errors.Service, logger *zap.Logger) *unartClientInterceptor {
	errorService = errorService.NewService("grpc")
	return &unartClientInterceptor{
		interceptors: make(map[string]*unaryClientInterceptorWapper),
		rootInterceptor: &unaryClientInterceptorWapper{
			interceptor:  newPaincInterceptor(ctx, errorService, logger),
			errorService: errorService,
			logger:       logger,
		},
		mutex:        new(sync.Mutex),
		errorService: errorService,
		logger:       logger,
		ctx:          ctx,
	}
}

type unartClientInterceptor struct {
	interceptors    map[string]*unaryClientInterceptorWapper
	rootInterceptor *unaryClientInterceptorWapper
	mutex           *sync.Mutex
	errorService    errors.Service
	logger          *zap.Logger
	ctx             context.Context
}

func (uci *unartClientInterceptor) Interceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) (err error) {
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
			err = invoker(ctx, method, req, reply, cc, opts...)
			return err
		}
		return uciw.errorService.SystemError("类型错误")
	}
	return uciw.next.interceptor(ctx, method, req, reply, cc, uciw.next.invoker, opts...)
}

func newPaincInterceptor(ctx context.Context, errorService errors.Service, logger *zap.Logger) grpc.UnaryClientInterceptor {
	return func(cxt context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) (err error) {
		defer func() {
			if r := recover(); r != nil {
				if err, ok := r.(error); ok {
					err = adapteError(err, errorService, logger)
					return
				}
				err = errorService.SystemError("无法识别的远程异常", logs.F_ExtendData(r))
			}
		}()
		return adapteError(invoker(cxt, method, req, reply, cc, opts...), errorService, logger)
	}
}

var errCode = codes.Code(18)

func adapteError(err error, errorService errors.Service, logger *zap.Logger) errors.Error {
	if err == nil {
		return nil
	}
	s, ok := status.FromError(err)
	if !ok {
		return errorService.SystemError("无法识别的RPC异常", logs.F_ExtendData(err))
	}
	if s.Code() == errCode {
		return errors.ParseError(s.Message())
	}
	return errorService.SystemError(fmt.Sprintf("RPC异常-%s", err.Error()), logs.F_Error(err), logs.F_ExtendData(s.Message()))
}
