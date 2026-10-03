package grpcserver

import (
	"context"

	"github.com/coffeehc/base/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func buildAuthUnaryServerInterceptor(authService GRPCServerAuth) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, errors.MessageError("没有认证信息")
		}
		_ctx, _err := authService.Auth(ctx, md)
		if _err != nil {
			return nil, _err
		}
		return handler(_ctx, req)
	}
}

func buildAuthStreamServerInterceptor(authService GRPCServerAuth) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		md, ok := metadata.FromIncomingContext(ss.Context())
		if !ok {
			return errors.MessageError("没有认证信息")
		}
		authCtx, _err := authService.Auth(ss.Context(), md)
		if _err != nil {
			return _err
		}
		return handler(srv, &authenticatedStream{ServerStream: ss, authCtx: authCtx})
	}
}

// authenticatedStream 将鉴权后的 context 绑定到本次 RPC，其他流操作保持原有实现。
type authenticatedStream struct {
	// ServerStream 是本次请求的底层流。
	grpc.ServerStream
	// authCtx 包含鉴权结果，只在本次流式 RPC 的生命周期内使用。
	authCtx context.Context
}

// Context 返回当前流鉴权后的身份 context。
func (s *authenticatedStream) Context() context.Context {
	return s.authCtx
}

// GRPCServerAuth 验证 RPC metadata，并返回包含身份信息的 context。
// 实现必须支持并发调用；失败时返回错误，成功时返回非 nil context。
type GRPCServerAuth interface {
	// Auth 拒绝未授权请求；返回的 context 会传递给 unary 和 streaming handler。
	Auth(ctx context.Context, md metadata.MD) (context.Context, error)
}
