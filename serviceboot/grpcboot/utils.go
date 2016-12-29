package grpcboot

import (
	"time"

	"github.com/coffeehc/logger"
	"github.com/coffeehc/microserviceboot/base"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

func loggingInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	start := time.Now()
	defer func() {
		logger.Debug("finished %s, took=%s, err=%v", info.FullMethod, time.Since(start), err)
	}()
	resp, err = handler(ctx, req)
	return resp, err
}

//NewResponseError create a response error ,code is base.ErrCodeBaseRPCInternal
func NewResponseError(messgae string) base.Error {
	return base.NewError(base.ErrCodeBaseRPCInternal, "response", messgae)
}

//NewServiceProcessError create a service response error ,code is base.ErrCodeBaseRPCAborted
func NewServiceProcessError(service string, message string) base.Error {
	return base.NewError(base.ErrCodeBaseRPCAborted, service, message)
}
