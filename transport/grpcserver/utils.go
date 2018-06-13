package grpcserver

import (
	"fmt"
	"time"

	"git.xiagaogao.com/coffee/boot/errors"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

func BuildLoggingInterceptor(errorService errors.Service, logger *zap.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		start := time.Now()
		defer func() {
			logger.Debug(fmt.Sprintf("finished %s, took=%s, err=%v", info.FullMethod, time.Since(start), err))
		}()
		resp, err = handler(ctx, req)
		return resp, err
	}
}
