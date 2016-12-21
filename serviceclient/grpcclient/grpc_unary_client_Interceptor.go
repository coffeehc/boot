package grpcclient

import (
	"fmt"
	"github.com/coffeehc/microserviceboot/base"
	"github.com/pquerna/ffjson/ffjson"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"sync"
)

var _unartClientInterceptor = newUnartClientInterceptor()

func AppendUnartClientInterceptor(name string, unaryClientInterceptor grpc.UnaryClientInterceptor) base.Error {
	return _unartClientInterceptor.AppendInterceptor(name, unaryClientInterceptor)
}

func newUnartClientInterceptor() *unartClientInterceptor {
	return &unartClientInterceptor{
		interceptors: make(map[string]*wapperUnartClientInterceptor),
		rootInterceptor: &wapperUnartClientInterceptor{
			interceptor: paincInterceptor,
		},
		mutex: new(sync.Mutex),
	}
}

type unartClientInterceptor struct {
	interceptors    map[string]*wapperUnartClientInterceptor
	rootInterceptor *wapperUnartClientInterceptor
	mutex           *sync.Mutex
}

func (this *unartClientInterceptor) Interceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) (err error) {
	return this.rootInterceptor.interceptor(context.WithValue(ctx, "_internal_invoker", invoker), method, req, reply, cc, this.rootInterceptor.invoker, opts...)
}

func (this *unartClientInterceptor) AppendInterceptor(name string, interceptor grpc.UnaryClientInterceptor) base.Error {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	if _, ok := this.interceptors[name]; ok {
		return base.NewError(base.ERRCODE_BASE_SYSTEM_INIT_ERROR, "grpc interceptor", fmt.Sprintf("%s 已经存在", name))
	}
	lastInterceptor := getLastUnaryClientInterceptor(this.rootInterceptor)
	lastInterceptor.next = &wapperUnartClientInterceptor{interceptor: interceptor}
	this.interceptors[name] = lastInterceptor.next
	return nil
}

func getLastUnaryClientInterceptor(root *wapperUnartClientInterceptor) *wapperUnartClientInterceptor {
	if root.next == nil {
		return root
	}
	return getLastUnaryClientInterceptor(root.next)
}

type wapperUnartClientInterceptor struct {
	interceptor grpc.UnaryClientInterceptor
	next        *wapperUnartClientInterceptor
}

func (this *wapperUnartClientInterceptor) invoker(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) (err error) {
	if this.next == nil {
		realInvoker := ctx.Value("_internal_invoker")
		if realInvoker == nil {
			return base.NewError(base.ERRCODE_BASE_SYSTEM_NIL, "grpc handler", "没有 Handler")
		}
		if invoker, ok := realInvoker.(grpc.UnaryInvoker); ok {
			return invoker(ctx, method, req, reply, cc, opts...)
		}
		return base.NewError(base.ERRCODE_BASE_SYSTEM_TYPE_CONV_ERROR, "grpc handler", "类型错误")
	}
	return this.next.interceptor(ctx, method, req, reply, cc, this.next.invoker, opts...)
}

func paincInterceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if _err, ok := r.(error); ok {
				if grpc.Code(_err) == 0xff {
					baseError := &base.BaseError{}
					__err := ffjson.Unmarshal([]byte(grpc.ErrorDesc(_err)), baseError)
					if __err != nil {
						err = base.NewError(base.ERRCODE_BASE_SYSTEM_DECODE_ERROR, "response", __err.Error())
						return
					}
					err = baseError
					return
				}
				err = _err
				return
			}
			err = base.NewError(base.ERRCODE_BASE_RPC_UNKNOWN, "response", fmt.Sprintf("%s", r))
		}
	}()
	err = invoker(ctx, method, req, reply, cc, opts...)
	if err != nil {
		panic(err)
	}
	return err
}
