// Copyright 2017 David Ackroyd. All Rights Reserved.
// See LICENSE for licensing terms.

package grpcrecovery

import (
	"go.uber.org/zap"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// UnaryServerInterceptor returns a new unary server interceptor for panic recovery.
func UnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) (err error) {
		defer func() {
			if r := recover(); r != nil {
				err = parseRPCError(r, true, zap.String("rpcMethod", method))
			}
		}()
		err = invoker(ctx, method, req, reply, cc, opts...)
		return parseRPCError(err, false)
	}
}

// StreamServerInterceptor returns a new streaming server interceptor for panic recovery.
func StreamClientInterceptor() grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (clientStream grpc.ClientStream, err error) {
		defer func() {
			if r := recover(); r != nil {
				err = parseRPCError(r, true, zap.String("rpcMethod", method))
			}
		}()
		clientStream, err = streamer(ctx, desc, cc, method, opts...)
		err = parseRPCError(err, false)
		return clientStream, err
	}
}
