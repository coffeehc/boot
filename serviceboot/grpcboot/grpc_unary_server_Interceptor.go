package grpcboot

import (
	"fmt"
	"github.com/coffeehc/microserviceboot/base"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"sync"
)

var _unartServerInterceptor = newUnartServerInterceptor()

func AppendUnartServerInterceptor(name string, unaryServerInterceptor grpc.UnaryServerInterceptor) base.Error {
	return _unartServerInterceptor.AppendInterceptor(name, unaryServerInterceptor)
}

func newUnartServerInterceptor() *unartServerInterceptor {
	return &unartServerInterceptor{
		interceptors: make(map[string]*wapperUnartServerInterceptor),
		rootInterceptor: &wapperUnartServerInterceptor{
			interceptor: paincInterceptor,
		},
		mutex: new(sync.Mutex),
	}
}

type unartServerInterceptor struct {
	interceptors    map[string]*wapperUnartServerInterceptor
	rootInterceptor *wapperUnartServerInterceptor
	mutex           *sync.Mutex
}

func (this *unartServerInterceptor) Interceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	return this.rootInterceptor.interceptor(context.WithValue(context.WithValue(ctx, "_internal_UnaryServerInfo", info), "_internal_handler", handler), req, info, this.rootInterceptor.handler)
}

func (this *unartServerInterceptor) AppendInterceptor(name string, interceptor grpc.UnaryServerInterceptor) base.Error {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	if _, ok := this.interceptors[name]; ok {
		return base.NewError(base.ERRCODE_BASE_SYSTEM_INIT_ERROR, "grpc interceptor", fmt.Sprintf("%s 已经存在", name))
	}
	lastInterceptor := getLastUnaryServerInterceptor(this.rootInterceptor)
	lastInterceptor.next = &wapperUnartServerInterceptor{interceptor: interceptor}
	this.interceptors[name] = lastInterceptor.next
	return nil
}

func getLastUnaryServerInterceptor(root *wapperUnartServerInterceptor) *wapperUnartServerInterceptor {
	if root.next == nil {
		return root
	}
	return getLastUnaryServerInterceptor(root.next)
}

type wapperUnartServerInterceptor struct {
	interceptor grpc.UnaryServerInterceptor
	next        *wapperUnartServerInterceptor
}

func (this *wapperUnartServerInterceptor) handler(ctx context.Context, req interface{}) (interface{}, error) {
	if this.next == nil {
		realHandler := ctx.Value("_internal_handler")
		if realHandler == nil {
			return nil, base.NewError(base.ERRCODE_BASE_SYSTEM_NIL, "grpc handler", "没有 Handler")
		}
		if handler, ok := realHandler.(grpc.UnaryHandler); ok {
			return handler(ctx, req)
		}
		return nil, base.NewError(base.ERRCODE_BASE_SYSTEM_TYPE_CONV_ERROR, "grpc handler", "类型错误")
	}
	info := ctx.Value("_internal_UnaryServerInfo")
	if info == 0 {
		return nil, base.NewError(base.ERRCODE_BASE_SYSTEM_NIL, "grpc interceptor", "没有 ServerInfo")
	}
	if unaryServerInfo, ok := info.(*grpc.UnaryServerInfo); ok {
		return this.next.interceptor(ctx, req, unaryServerInfo, this.next.handler)
	}
	return nil, base.NewError(base.ERRCODE_BASE_SYSTEM_TYPE_CONV_ERROR, "grpc interceptor", "类型错误")
}

func paincInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			if _err, ok := r.(base.Error); ok {
				err = grpc.Errorf(255, _err.Error())
				return
			}
			if _err, ok := r.(error); ok {
				err = grpc.Errorf(codes.Internal, _err.Error())
				return
			}
			err = grpc.Errorf(codes.Unknown, "%s", r)
		}
	}()
	resp, err = handler(ctx, req)
	if err != nil {
		panic(err)
	}
	return resp, nil
}
