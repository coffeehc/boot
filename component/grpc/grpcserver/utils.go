package grpcserver

import (
	"fmt"
	"time"

	"git.xiagaogao.com/coffee/boot/base/log"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

func DebufLoggingInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		start := time.Now()
		defer func() {
			log.Debug(fmt.Sprintf("finished %s, took=%s, err=%v", info.FullMethod, time.Since(start), err), scope)
		}()
		resp, err = handler(ctx, req)
		return resp, err
	}
}
