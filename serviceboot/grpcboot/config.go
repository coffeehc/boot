package grpcboot

import (
	"github.com/coffeehc/microserviceboot/base"
	"github.com/coffeehc/microserviceboot/serviceboot"
	"github.com/grpc-ecosystem/go-grpc-prometheus"
	"google.golang.org/grpc"
)

//Config grpcboot config
type Config struct {
	ServiceConfig *serviceboot.ServiceConfig `yaml:"service_config"`
	GRPCConfig    struct {
		MaxMsgSize           int    `yaml:"max_msg_size"`
		MaxConcurrentStreams uint32 `yaml:"max_concurrent_streams"`
	} `yaml:"grpc_config"`
}

//GetGRPCOptions 获取 GRPCOption
func (config *Config) GetGRPCOptions() []grpc.ServerOption {
	config.initGRPCConfig()
	grpc.EnableTracing = false
	if base.IsDevModule() {
		grpc.EnableTracing = true
		AppendUnaryServerInterceptor("logger", loggingInterceptor)
	}
	AppendUnaryServerInterceptor("prometheus", grpc_prometheus.UnaryServerInterceptor)
	return []grpc.ServerOption{
		grpc.MaxConcurrentStreams(config.GRPCConfig.MaxConcurrentStreams),
		grpc.MaxMsgSize(config.GRPCConfig.MaxMsgSize),
		grpc.StreamInterceptor(grpc_prometheus.StreamServerInterceptor),
		grpc.UnaryInterceptor(_unaryServerInterceptor.Interceptor),
		grpc.RPCCompressor(grpc.NewGZIPCompressor()),
		grpc.RPCDecompressor(grpc.NewGZIPDecompressor()),
	}
}

//GetServiceConfig 获取 Service Config
func (config *Config) GetServiceConfig() *serviceboot.ServiceConfig {
	if config.ServiceConfig == nil {
		config.ServiceConfig = new(serviceboot.ServiceConfig)
	}
	return config.ServiceConfig
}

func (config *Config) initGRPCConfig() {
	grpcConfig := config.GRPCConfig
	if grpcConfig.MaxConcurrentStreams == 0 {
		grpcConfig.MaxConcurrentStreams = 100000
	}
	if grpcConfig.MaxMsgSize == 0 {
		grpcConfig.MaxMsgSize = 1024 * 1024 * 4
	}
}
