package grpcbase

import (
	"github.com/coffeehc/microserviceboot/base"
	"google.golang.org/grpc"
)

type GRPCService interface {
	base.Service
	GetGRPCOptions() []grpc.ServerOption
	RegisterServer(s *grpc.Server) base.Error
}
