// Copyright 2017 David Ackroyd. All Rights Reserved.
// See LICENSE for licensing terms.

package grpcrecovery

import (
	"git.xiagaogao.com/coffee/boot/errors"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// UnaryServerInterceptor returns a new unary server interceptor for panic recovery.
func UnaryServerInterceptor(errorService errors.Service, logger *zap.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (_ interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				err = recoverFrom(r, errorService, logger)
			}
		}()
		resp, err := handler(ctx, req)
		err = recoverFrom(err, errorService, logger)
		return resp, err
	}
}

// StreamServerInterceptor returns a new streaming server interceptor for panic recovery.
func StreamServerInterceptor(errorService errors.Service, logger *zap.Logger) grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
		defer func() {
			if r := recover(); r != nil {
				err = recoverFrom(r, errorService, logger)
			}
		}()
		return recoverFrom(handler(srv, stream), errorService, logger)
	}
}

func recoverFrom(err interface{}, errorService errors.Service, logger *zap.Logger) error {
	if err == nil {
		return nil
	}
	e := errors.ConverUnkonwError(err, errorService)
	if errors.IsSystemError(e) {
		logger.DPanic(e.Error(), e.GetFields()...)
	}
	if errors.IsMessageError(e) {
		logger.Error(e.Error(), e.GetFields()...)
	}
	switch v := err.(type) {
	case errors.Error:
		return status.Errorf(18, v.FormatRPCError())
	case string:
		return status.Errorf(codes.Internal, v)
	case error:
		return status.Errorf(codes.Internal, v.Error())
	default:
		return status.Errorf(codes.Unknown, "%#v", v)
	}
}
