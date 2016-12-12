package grpcboot

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"runtime"
	"github.com/coffeehc/logger"
	"google.golang.org/grpc/codes"
	"github.com/coffeehc/microserviceboot/base"
	"fmt"
	"sync"
)

var _unartServerInterceptor = newUnartServerInterceptor()

func AppendUnartServerInterceptor(name string,unaryServerInterceptor grpc.UnaryServerInterceptor)base.Error{
	return _unartServerInterceptor.AppendInterceptor(name,unaryServerInterceptor)
}

func newUnartServerInterceptor()*unartServerInterceptor{
	return &unartServerInterceptor{
		interceptors: map[string]*wapperUnartServerInterceptor{
			"_root":&wapperUnartServerInterceptor{
				interceptor:paincInterceptor,
			},
		},
		mutex:new(sync.Mutex),
	}
}

type unartServerInterceptor struct {
	interceptors map[string]*wapperUnartServerInterceptor
	mutex *sync.Mutex
}

func (this *unartServerInterceptor)Interceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error){
	context.WithValue(context.WithValue(ctx,"_internal_UnaryServerInfo",info),"_internal_handler",handler)
	defer func() {
		if r := recover(); r != nil {
			stack := make([]byte, 1024)
			stack = stack[:runtime.Stack(stack, false)]
			logger.Error("panic grpc invoke: %s, err=%v, stack:\n", info.FullMethod, r, string(stack))
			err = grpc.Errorf(codes.Internal, "panic error: %v", r)
		}
	}()
	return handler(ctx, req)
}

func (this *unartServerInterceptor)AppendInterceptor(name string,interceptor grpc.UnaryServerInterceptor)base.Error{
	this.mutex.Lock()
	defer this.mutex.Unlock()
	if _,ok:=this.interceptors[name];ok{
		return base.NewError(base.ERRCODE_BASE_SYSTEM_INIT_ERROR,"grpc interceptor",fmt.Sprintf("%s 已经存在",name))
	}
	lastInterceptor := getLastUnaryServerInterceptor(this.interceptors["_root"])
	lastInterceptor.next = &wapperUnartServerInterceptor{interceptor:interceptor}
	this.interceptors[name] = lastInterceptor.next
	return nil
}

func getLastUnaryServerInterceptor(root *wapperUnartServerInterceptor)*wapperUnartServerInterceptor{
	if root.next == nil{
		return root
	}
	return getLastUnaryServerInterceptor(root.next)
}

type wapperUnartServerInterceptor struct {
	interceptor grpc.UnaryServerInterceptor
	next *wapperUnartServerInterceptor
}

func (this *wapperUnartServerInterceptor)handler(ctx context.Context, req interface{}) (interface{}, error){
	if this.next == nil{
		realHandler := ctx.Value("_internal_handler")
		if realHandler == nil{
			return nil,base.NewError(base.ERRCODE_BASE_SYSTEM_NIL,"grpc handler","没有 Handler")
		}
		if handler,ok:=realHandler.(grpc.UnaryHandler);ok{
			return handler(ctx,req)
		}
		return nil,base.NewError(base.ERRCODE_BASE_SYSTEM_TYPE_CONV_ERROR,"grpc handler","类型错误")
	}
	info := ctx.Value("_internal_UnaryServerInfo")
	if info == 0{
		return nil,base.NewError(base.ERRCODE_BASE_SYSTEM_NIL,"grpc interceptor","没有 ServerInfo")
	}
	if unaryServerInfo,ok:=info.(*grpc.UnaryServerInfo);ok{
		return this.next.interceptor(ctx,req,unaryServerInfo,this.handler)
	}
	return nil,base.NewError(base.ERRCODE_BASE_SYSTEM_TYPE_CONV_ERROR,"grpc interceptor","类型错误")
}

func paincInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			stack := make([]byte, 1024)
			stack = stack[:runtime.Stack(stack, false)]
			logger.Error("panic grpc invoke: %s, err=%v, stack:\n", info.FullMethod, r, string(stack))
			err = grpc.Errorf(codes.Internal, "panic error: %v", r)
		}
	}()
	return handler(ctx, req)
}

