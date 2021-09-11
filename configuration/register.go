package configuration

import (
	"context"

	"github.com/coffeehc/base/errors"
	"google.golang.org/grpc"
)

type ServiceRegisterInfo struct {
	Info           ServiceInfo       `json:"info"`
	ServiceAddr    string            `json:"service_addr"`
	ManageEndpoint string            `json:"manage_endpoint"`
	Metadata       map[string]string `json:"metadata,omitempty"`
}

type RPCService interface {
	GetRPCServiceInfo() ServiceInfo
	InitRPCService(ctx context.Context, grpcConn *grpc.ClientConn) errors.Error
}
