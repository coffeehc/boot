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
		_err := authService.Auth(ctx, md)
		if _err != nil {
			return nil, _err
		}
		return handler(ctx, req)
	}
}

func buildAuthStreamServerInterceptor(authService GRPCServerAuth) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		md, ok := metadata.FromIncomingContext(ss.Context())
		if !ok {
			return errors.MessageError("没有认证信息")
		}
		_err := authService.Auth(ss.Context(), md)
		if _err != nil {
			return _err
		}
		ss.Context()
		return handler(srv, ss)
	}
}

type GRPCServerAuth interface {
	Auth(ctx context.Context, md metadata.MD) error
}
