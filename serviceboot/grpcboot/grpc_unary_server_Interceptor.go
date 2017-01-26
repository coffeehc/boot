package grpcboot

import (
	"fmt"
	"sync"

	"github.com/coffeehc/microserviceboot/base"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

var (
	_internalUnaryServerInfo = "_internal_UnaryServerInfo"
	_internalHandler         = "_internal_handler"
)

var _unaryServerInterceptor = newUnaryServerInterceptor()

//AppendUnaryServerInterceptor 追加新的UnaryServerInterceptor
func AppendUnaryServerInterceptor(name string, unaryServerInterceptor grpc.UnaryServerInterceptor) base.Error {
	return _unaryServerInterceptor.AppendInterceptor(name, unaryServerInterceptor)
}

func newUnaryServerInterceptor() *unaryServerInterceptor {
	return &unaryServerInterceptor{
		interceptors: make(map[string]*unaryServerInterceptorWapper),
		rootInterceptor: &unaryServerInterceptorWapper{
			interceptor: catchPanicInterceptor,
		},
		mutex: new(sync.Mutex),
	}
}

type unaryServerInterceptor struct {
	interceptors    map[string]*unaryServerInterceptorWapper
	rootInterceptor *unaryServerInterceptorWapper
	mutex           *sync.Mutex
}

func (usi *unaryServerInterceptor) Interceptor(cxt context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	cxt = context.WithValue(cxt, _internalUnaryServerInfo, info)
	cxt = context.WithValue(cxt, _internalHandler, handler)
	return usi.rootInterceptor.interceptor(cxt, req, info, usi.rootInterceptor.handler)
}

func (usi *unaryServerInterceptor) AppendInterceptor(name string, interceptor grpc.UnaryServerInterceptor) base.Error {
	usi.mutex.Lock()
	defer usi.mutex.Unlock()
	if _, ok := usi.interceptors[name]; ok {
		return base.NewError(base.ErrCodeBaseSystemInit, "grpc interceptor", fmt.Sprintf("%s 已经存在", name))
	}
	lastInterceptor := getLastUnaryServerInterceptor(usi.rootInterceptor)
	lastInterceptor.next = &unaryServerInterceptorWapper{interceptor: interceptor}
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
	interceptor grpc.UnaryServerInterceptor
	next        *unaryServerInterceptorWapper
}

func (usiw *unaryServerInterceptorWapper) handler(ctx context.Context, req interface{}) (interface{}, error) {
	if usiw.next == nil {
		realHandler := ctx.Value(_internalHandler)
		if realHandler == nil {
			return nil, base.NewError(base.ErrCodeBaseSystemNil, "grpc handler", "没有 Handler")
		}
		if handler, ok := realHandler.(grpc.UnaryHandler); ok {
			return handler(ctx, req)
		}
		return nil, base.NewError(base.ErrCodeBaseSystemTypeConversion, "grpc handler", "类型错误")
	}
	info := ctx.Value(_internalUnaryServerInfo)
	if info == 0 {
		return nil, base.NewError(base.ErrCodeBaseSystemNil, "grpc interceptor", "没有 ServerInfo")
	}
	if unaryServerInfo, ok := info.(*grpc.UnaryServerInfo); ok {
		return usiw.next.interceptor(ctx, req, unaryServerInfo, usiw.next.handler)
	}
	return nil, base.NewError(base.ErrCodeBaseSystemTypeConversion, "grpc interceptor", "类型错误")
}

func catchPanicInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			if _err, ok := r.(base.Error); ok {
				err = grpc.Errorf(0xff, _err.Error())
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
