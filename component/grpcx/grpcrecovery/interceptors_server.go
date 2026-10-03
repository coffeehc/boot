// Copyright 2017 David Ackroyd. All Rights Reserved.
// See LICENSE for licensing terms.

package grpcrecovery

import (
	"context"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// UnaryServerInterceptor returns a new unary server interceptor for panic recovery.
func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (_ interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				err = convertRPCError(r, true, zap.String("rpcMethod", info.FullMethod))
			}
		}()
		md, ok := metadata.FromIncomingContext(ctx)
		if ok {
			ctx = ParseMetadataToContext(ctx, md)
		}
		resp, err := handler(ctx, req)
		err = convertRPCError(err, false, zap.String("rpcMethod", info.FullMethod))
		return resp, err
	}
}

// StreamServerInterceptor returns a new streaming server interceptor for panic recovery.
func StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
		defer func() {
			if r := recover(); r != nil {
				err = convertRPCError(r, true, zap.String("rpcMethod", info.FullMethod))
			}
		}()
		ctx := stream.Context()
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			ctx = ParseMetadataToContext(ctx, md)
			stream = &tracedServerStream{ServerStream: stream, traceCtx: ctx}
		}
		return convertRPCError(handler(srv, stream), false, zap.String("rpcMethod", info.FullMethod))
	}
}

// tracedServerStream 将 unary 已支持的追踪 context 传递给流式处理链。
type tracedServerStream struct {
	// ServerStream 承载实际的流收发。
	grpc.ServerStream
	// traceCtx 仅在当前流式 RPC 中使用，继承底层流的取消和 deadline。
	traceCtx context.Context
}

// Context 返回继承流取消和 deadline 的追踪 context。
func (s *tracedServerStream) Context() context.Context {
	return s.traceCtx
}
