package grpcbase

import (
	"github.com/coffeehc/microserviceboot/base"
	"google.golang.org/grpc"
)

type GRpcService interface {
	base.Service
	GetGrpcOptions() []grpc.ServerOption
	RegisterServer(s *grpc.Server) base.Error
}
