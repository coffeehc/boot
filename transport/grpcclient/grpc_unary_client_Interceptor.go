package grpcclient

import (
	"fmt"
	"sync"

	"git.xiagaogao.com/coffee/boot"
	"git.xiagaogao.com/coffee/boot/errors"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

var (
	_unaryClientInterceptor = newUnartClientInterceptor()
)

const _internalInvoker = "_internal_invoker"
const context_serviceInfoKey = "__serviceInfo__"

func wapperUnartClientInterceptor(serviceInfo boot.ServiceInfo) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) (err error) {
		return _unaryClientInterceptor.Interceptor(context.WithValue(ctx, context_serviceInfoKey, serviceInfo), method, req, reply, cc, invoker, opts...)
	}
}

//AppendUnartClientInterceptor 追加一个UnartClientInterceptor
func AppendUnartClientInterceptor(name string, unaryClientInterceptor grpc.UnaryClientInterceptor) errors.Error {
	return _unaryClientInterceptor.AppendInterceptor(name, unaryClientInterceptor)
}

func newUnartClientInterceptor() *unartClientInterceptor {
	return &unartClientInterceptor{
		interceptors: make(map[string]*unaryClientInterceptorWapper),
		rootInterceptor: &unaryClientInterceptorWapper{
			interceptor: paincInterceptor,
		},
		mutex: new(sync.Mutex),
	}
}

type unartClientInterceptor struct {
	interceptors    map[string]*unaryClientInterceptorWapper
	rootInterceptor *unaryClientInterceptorWapper
	mutex           *sync.Mutex
}

func (uci *unartClientInterceptor) Interceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) (err error) {
	opts = append(opts, grpc.FailFast(true))
	return uci.rootInterceptor.interceptor(context.WithValue(ctx, _internalInvoker, invoker), method, req, reply, cc, uci.rootInterceptor.invoker, opts...)
}

func (uci *unartClientInterceptor) AppendInterceptor(name string, interceptor grpc.UnaryClientInterceptor) errors.Error {
	uci.mutex.Lock()
	defer uci.mutex.Unlock()
	if _, ok := uci.interceptors[name]; ok {
		return errors.NewError(errors.Error_System, "grpcserver interceptor", fmt.Sprintf("%s 已经存在", name))
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
	interceptor grpc.UnaryClientInterceptor
	next        *unaryClientInterceptorWapper
}

func (uciw *unaryClientInterceptorWapper) invoker(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) (err error) {
	if uciw.next == nil {
		realInvoker := ctx.Value(_internalInvoker)
		if realInvoker == nil {
			return errors.NewError(errors.Error_System, "grpcserver", "没有 Handler")
		}
		if invoker, ok := realInvoker.(grpc.UnaryInvoker); ok {
			return invoker(ctx, method, req, reply, cc, opts...)
		}
		return errors.NewError(errors.Error_System, "grpcserver", "类型错误")
	}
	return uciw.next.interceptor(ctx, method, req, reply, cc, uciw.next.invoker, opts...)
}

func paincInterceptor(cxt context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = adapteError(cxt, r)
		}
	}()
	return adapteError(cxt, invoker(cxt, method, req, reply, cc, opts...))
}

func adapteError(cxt context.Context, err interface{}) errors.Error {
	if err == nil {
		return nil
	}
	serviceName := "未知服务"
	serviceInfo, ok := cxt.Value(context_serviceInfoKey).(boot.ServiceInfo)
	if ok {
		serviceName = serviceInfo.GetServiceName()
	}
	serviceName = "grpcserver:" + serviceName
	switch v := err.(type) {
	case errors.Error:
		return v
	case string:
		return errors.NewError(errors.Error_System, serviceName, v)
	case error:
		if s, ok := status.FromError(v); ok {
			code := int32(s.Code())
			if !errors.IsBaseErrorCode(code) {
				return errors.NewErrorWrapper(errors.Error_System, serviceName, s.Err())
			}
			return errors.NewError(code, serviceName, s.Message())
		}
		return errors.NewErrorWrapper(errors.Error_System, serviceName, v)
	default:
		return errors.NewError(errors.Error_System, serviceName, fmt.Sprintf("未知异常:%#v", v))
	}

}
