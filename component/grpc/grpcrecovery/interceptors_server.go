// Copyright 2017 David Ackroyd. All Rights Reserved.
// See LICENSE for licensing terms.

package grpcrecovery

import (
	"go.uber.org/zap"
	"golang.org/x/net/context"
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
		err = convertRPCError(err, false)
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
		return convertRPCError(handler(srv, stream), false)
	}
}
