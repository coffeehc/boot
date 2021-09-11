package grpcserver

import (
	"fmt"
	"time"

	"github.com/coffeehc/base/log"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var EnableAccessLog bool = false

func DebugLoggingInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		start := time.Now()
		defer func() {
			log.Debug(fmt.Sprintf("FullMethod %s, took=%s, err=%v", info.FullMethod, time.Since(start), err), scope)
		}()
		resp, err = handler(ctx, req)
		return resp, err
	}
}

const (
	contextkeyServerCerds = "_grpc.serverCredentials"
	serverGrpcAuthKey     = "__ServerGrpcAuthKey"
)

func SetGrpcAuth(ctx context.Context, auth GRPCServerAuth) context.Context {
	return context.WithValue(ctx, serverGrpcAuthKey, auth)
}

func SetCerds(ctx context.Context, creds credentials.TransportCredentials) context.Context {
	return context.WithValue(ctx, contextkeyServerCerds, creds)
}

func getCerts(ctx context.Context) credentials.TransportCredentials {
	v := ctx.Value(contextkeyServerCerds)
	if v == nil {
		return nil
	}
	if cerds, ok := v.(credentials.TransportCredentials); ok {
		return cerds
	}
	return nil
}
