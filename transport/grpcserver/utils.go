package grpcserver

import (
	"fmt"
	"time"

	"git.xiagaogao.com/coffee/boot/logs"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

func loggingInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	logger := logs.GetLogger(ctx)
	start := time.Now()
	defer func() {
		logger.Debug(fmt.Sprintf("finished %s, took=%s, err=%v", info.FullMethod, time.Since(start), err))
	}()
	resp, err = handler(ctx, req)
	return resp, err
}
