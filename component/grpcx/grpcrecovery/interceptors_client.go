// Copyright 2017 David Ackroyd. All Rights Reserved.
// See LICENSE for licensing terms.

package grpcrecovery

import (
	"context"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// UnaryClientInterceptor 合并追踪 metadata、恢复 panic 并解码 Boot 业务错误。
// 未设置 deadline 的 unary 调用默认限时一分钟，调用方可显式设置其他 deadline。
func UnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) (err error) {
		defer func() {
			if r := recover(); r != nil {
				err = parseRPCError(r, true, zap.String("rpcMethod", method), zap.Any("connState", cc.GetState()), zap.String("target", cc.Target()))
			}
		}()
		_, ok := ctx.Deadline()
		if !ok {
			_ctx, cancelFunc := context.WithTimeout(ctx, time.Minute)
			defer cancelFunc()
			ctx = _ctx
		}
		md := BuildMetadataFromContext(ctx)
		ctx = metadata.NewOutgoingContext(ctx, md)
		err = invoker(ctx, method, req, reply, cc, opts...)
		return parseRPCError(err, false, zap.String("rpcMethod", method), zap.Any("connState", cc.GetState()), zap.String("target", cc.Target()))
	}
}

// StreamClientInterceptor 合并追踪 metadata 并恢复流创建和收发中的 panic。
// 流不设置默认 deadline，由调用方负责取消或结束流；正常 EOF 原样返回。
func StreamClientInterceptor() grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (clientStream grpc.ClientStream, err error) {
		defer func() {
			if r := recover(); r != nil {
				err = parseRPCError(r, true, zap.String("rpcMethod", method), zap.Any("connState", cc.GetState()), zap.String("target", cc.Target()))
			}
		}()
		md := BuildMetadataFromContext(ctx)
		ctx = metadata.NewOutgoingContext(ctx, md)
		clientStream, err = streamer(ctx, desc, cc, method, opts...)
		err = parseRPCError(err, false, zap.String("rpcMethod", method), zap.Any("connState", cc.GetState()), zap.String("target", cc.Target()))
		if err != nil {
			return nil, err
		}
		return &recoveringClientStream{ClientStream: clientStream, method: method}, nil
	}
}

// recoveringClientStream 解码流创建完成后返回的业务错误，保留 EOF 和标准状态码。
type recoveringClientStream struct {
	// ClientStream 承载实际的流收发及 context 生命周期。
	grpc.ClientStream
	// method 用于 panic 日志定位，不保存请求或响应内容。
	method string
}

// Header 取得响应头并解码创建流之后发生的服务错误。
func (s *recoveringClientStream) Header() (md metadata.MD, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = parseRPCError(r, true, zap.String("rpcMethod", s.method))
		}
	}()
	md, err = s.ClientStream.Header()
	return md, parseRPCError(err, false, zap.String("rpcMethod", s.method))
}

// SendMsg 在底层流发送消息，编解码 panic 转换为调用错误。
func (s *recoveringClientStream) SendMsg(m interface{}) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = parseRPCError(r, true, zap.String("rpcMethod", s.method))
		}
	}()
	return parseRPCError(s.ClientStream.SendMsg(m), false, zap.String("rpcMethod", s.method))
}

// RecvMsg 解码远端 Boot 错误，正常流结束仍返回 io.EOF。
func (s *recoveringClientStream) RecvMsg(m interface{}) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = parseRPCError(r, true, zap.String("rpcMethod", s.method))
		}
	}()
	return parseRPCError(s.ClientStream.RecvMsg(m), false, zap.String("rpcMethod", s.method))
}

// CloseSend 结束发送方向，保留正常结束和标准 gRPC 错误语义。
func (s *recoveringClientStream) CloseSend() (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = parseRPCError(r, true, zap.String("rpcMethod", s.method))
		}
	}()
	return parseRPCError(s.ClientStream.CloseSend(), false, zap.String("rpcMethod", s.method))
}
