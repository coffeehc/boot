package grpcboot

import ()
import (
	"github.com/coffeehc/logger"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"time"
)

func loggingInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	start := time.Now()
	logger.Debug("calling %s", info.FullMethod)
	resp, err = handler(ctx, req)
	logger.Debug("finished %s, took=%s, err=%v", info.FullMethod, time.Since(start), err)
	return resp, err
}


