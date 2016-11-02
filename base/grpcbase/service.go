package grpcbase

import (
	"github.com/coffeehc/microserviceboot/base"
	"google.golang.org/grpc"
)

type GRpcService interface {
	base.Service
	GetGrpcOptions() []grpc.ServerOption
	Init(configPath string) base.Error
	RegisterServer(s *grpc.Server) base.Error
}
