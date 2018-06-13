package serviceboot

import (
	"context"
	"flag"

	"git.xiagaogao.com/coffee/boot"
	"git.xiagaogao.com/coffee/boot/bootutils"
	"git.xiagaogao.com/coffee/boot/errors"
	"git.xiagaogao.com/coffee/boot/sd/etcdsd"
	"git.xiagaogao.com/coffee/boot/transport/grpcserver"
	"go.uber.org/zap"
)

var configPath = flag.String("config", "./config.yml", "配置文件路径")

// ServiceConfig 服务配置
type ServiceConfig struct {
	DisableServiceRegister bool                   `yaml:"disable_service_register"`
	ServerAddr             string                 `yaml:"server_addr"`
	GrpcConfig             *grpcserver.GRPCConfig `yaml:"grpc_config"`
	EtcdConfig             *etcdsd.Config         `yaml:"etcd_config"`
}

//LoadConfig 加载ServiceConfiguration的配置
func loadServiceConfig(ctx context.Context, errorService errors.Service, logger *zap.Logger) (*ServiceConfig, string, errors.Error) {
	config := &ServiceConfig{}
	loaded := false
	if boot.IsDevModule() && *configPath == "./config.yml" {
		err := bootutils.LoadConfig(ctx, "./config-dev.yml", config, errorService, logger)
		if err == nil {
			*configPath = "./config-dev.yml"
			loaded = true
		}
	}
	if !loaded {
		err := bootutils.LoadConfig(ctx, *configPath, config, errorService, logger)
		if err != nil {
			return nil, "", err
		}
	}
	if config.ServerAddr == "" {
		return nil, "", errorService.MessageError("没有配置ServiceAddr")
	}
	serverAddr, err := bootutils.WarpServerAddr(config.ServerAddr, errorService)
	if err != nil {
		return nil, "", err
	}
	config.ServerAddr = serverAddr
	return config, *configPath, nil
}
