package grpcboot

import (
	"github.com/coffeehc/microserviceboot/serviceboot"
	"github.com/grpc-ecosystem/go-grpc-prometheus"
	"google.golang.org/grpc"
)

type Config struct {
	BaseConfig *serviceboot.ServiceConfig `yaml:"base_config"`
	GrpcConfig *GRpcConfig                `yaml:"grpc_config"`
}

type GRpcConfig struct {
	MaxMsgSize           int    `yaml:"max_msg_size"`
	MaxConcurrentStreams uint32 `yaml:"max_concurrent_streams"`
}

func (this *GRpcConfig) GetGrpcOptions() []grpc.ServerOption {
	return []grpc.ServerOption{
		grpc.MaxConcurrentStreams(this.MaxConcurrentStreams),
		grpc.MaxMsgSize(this.MaxMsgSize),
		grpc.StreamInterceptor(grpc_prometheus.StreamServerInterceptor),
		grpc.UnaryInterceptor(grpc_prometheus.UnaryServerInterceptor),
	}
}

func (this *Config) GetBaseConfig() *serviceboot.ServiceConfig {
	if this.BaseConfig == nil {
		this.BaseConfig = new(serviceboot.ServiceConfig)
	}
	return this.BaseConfig
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
