package grpcboot

import (
	"github.com/coffeehc/microserviceboot/serviceboot"
	"google.golang.org/grpc"
)

type Config struct {
	serviceboot.ServiceConfig
	GrpcConfig *GRpcConfig `yaml:"grpc_config"`
}

type GRpcConfig struct {
	MaxMsgSize           int    `yaml:"max_msg_size"`
	MaxConcurrentStreams uint32 `yaml:"max_concurrent_streams"`
}

func (this *GRpcConfig) GetGrpcOptions() []grpc.ServerOption {
	return []grpc.ServerOption{
		grpc.MaxConcurrentStreams(this.MaxConcurrentStreams),
		grpc.MaxMsgSize(this.MaxMsgSize),
	}
}

func (this *Config) GetGRpcServerConfig() *GRpcConfig {
	grpcConfig := this.GrpcConfig
	if grpcConfig == nil {
		grpcConfig = new(GRpcConfig)
	}
	if grpcConfig.MaxConcurrentStreams == 0 {
		grpcConfig.MaxConcurrentStreams = 100000
	}
	if grpcConfig.MaxMsgSize == 0 {
		grpcConfig.MaxMsgSize = 1024 * 1024 * 4
	}
	return grpcConfig
}
