package grpcrecovery

import (
	"git.xiagaogao.com/coffee/boot/errors"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func recoverFrom(err interface{}, errorService errors.Service, logger *zap.Logger) error {
	if err == nil {
		return nil
	}
	switch v := err.(type) {
	case errors.Error:
		if errors.IsSystemError(v) {
			logger.WithOptions(zap.AddCallerSkip(2)).DPanic(v.Error(), v.GetFields()...)
		}
		return status.Errorf(18, v.FormatRPCError())
	case string:
		return status.Errorf(codes.Internal, v)
	case error:
		return status.Errorf(codes.Internal, v.Error())
	default:
		return status.Errorf(codes.Unknown, "%#v", v)
	}
}
