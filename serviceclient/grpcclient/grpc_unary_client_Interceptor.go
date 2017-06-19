package grpcclient

import (
	"fmt"
	"sync"

	"reflect"

	"github.com/coffeehc/logger"
	"github.com/coffeehc/microserviceboot/base"
	"github.com/coffeehc/microserviceboot/pb"
	"golang.org/x/net/context"
	spb "google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

var (
	_errDetailTypeURL       = "grpc.errdetail"
	_unaryClientInterceptor = newUnartClientInterceptor()
)

func init() {
	pb.RegisterType(_errDetailTypeURL, reflect.TypeOf(spb.Status{}))
}

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
	opts = append(opts, grpc.FailFast(false))
	return uci.rootInterceptor.interceptor(context.WithValue(ctx, _internalInvoker, invoker), method, req, reply, cc, uci.rootInterceptor.invoker, opts...)
}

func (uci *unartClientInterceptor) AppendInterceptor(name string, interceptor grpc.UnaryClientInterceptor) base.Error {
	uci.mutex.Lock()
	defer uci.mutex.Unlock()
	if _, ok := uci.interceptors[name]; ok {
		return base.NewError(base.ErrCode_System, "grpc interceptor", fmt.Sprintf("%s 已经存在", name))
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
			return base.NewError(base.ErrCode_System, "grpc", "没有 Handler")
		}
		if invoker, ok := realInvoker.(grpc.UnaryInvoker); ok {
			return invoker(ctx, method, req, reply, cc, opts...)
		}
		return base.NewError(base.ErrCode_System, "grpc", "类型错误")
	}
	return uciw.next.interceptor(ctx, method, req, reply, cc, uciw.next.invoker, opts...)
}

func paincInterceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if status, ok := r.(status.Status); ok {
				pbStatus := status.Proto()

				logger.Debug("status is %#v", status)
				if pbStatus.Code == 0xff {
					//status.Message
					//err = base.ParseErrorFromJSON([]byte(grpc.ErrorDesc(_err)))
					//err = pbStatus
					logger.Debug("出现错误")
					return
				}
				//err = pbStatus
				return
			}
			err = base.NewError(base.ErrCode_System, "response", fmt.Sprintf("%s", r))
		}
	}()
	err = invoker(ctx, method, req, reply, cc, opts...)
	if err != nil {
		panic(err)
	}
	return err
}
