package grpcclient

import (
	"fmt"
	"sync"

	"github.com/coffeehc/commons/convers"
	"github.com/coffeehc/microserviceboot/base"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var _unaryClientInterceptor = newUnartClientInterceptor()

const _internalInvoker = "_internal_invoker"

//AppendUnartClientInterceptor 追加一个UnartClientInterceptor
func AppendUnartClientInterceptor(name string, unaryClientInterceptor grpc.UnaryClientInterceptor) base.Error {
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
	return uci.rootInterceptor.interceptor(context.WithValue(ctx, _internalInvoker, invoker), method, req, reply, cc, uci.rootInterceptor.invoker, opts...)
}

func (uci *unartClientInterceptor) AppendInterceptor(name string, interceptor grpc.UnaryClientInterceptor) base.Error {
	uci.mutex.Lock()
	defer uci.mutex.Unlock()
	if _, ok := uci.interceptors[name]; ok {
		return base.NewError(base.ErrCodeBaseSystemInit, "grpc interceptor", fmt.Sprintf("%s 已经存在", name))
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
			return base.NewError(base.ErrCodeBaseSystemNil, "grpc handler", "没有 Handler")
		}
		if invoker, ok := realInvoker.(grpc.UnaryInvoker); ok {
			return invoker(ctx, method, req, reply, cc, opts...)
		}
		return base.NewError(base.ErrCodeBaseSystemTypeConversion, "grpc handler", "类型错误")
	}
	return uciw.next.interceptor(ctx, method, req, reply, cc, uciw.next.invoker, opts...)
}

func paincInterceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if _err, ok := r.(error); ok {
				if grpc.Code(_err) == 0xff {
					err = base.ParseErrorFromJSON(convers.StringToBytes(grpc.ErrorDesc(_err)))
					return
				}
				err = _err
				return
			}
			err = base.NewError(base.ErrCodeBaseRPCUnknown, "response", fmt.Sprintf("%s", r))
		}
	}()
	err = invoker(ctx, method, req, reply, cc, opts...)
	if err != nil {
		panic(err)
	}
	return err
}
