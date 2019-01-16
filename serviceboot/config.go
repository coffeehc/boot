package serviceboot

import (
	"context"
	"fmt"
	"os"

	"git.xiagaogao.com/coffee/boot"
	"git.xiagaogao.com/coffee/boot/bootutils"
	"git.xiagaogao.com/coffee/boot/errors"
	"git.xiagaogao.com/coffee/boot/sd/etcdsd"
	"git.xiagaogao.com/coffee/boot/transport/grpcserver"
	"go.uber.org/zap"
)

// ServiceConfig 服务配置
type ServiceConfig struct {
	DisableServiceRegister bool                   `yaml:"disable_service_register"`
	SingleService          bool                   `yaml:"single_service"`
	GrpcConfig             *grpcserver.GRPCConfig `yaml:"grpc_config"`
	EtcdConfig             *etcdsd.Config         `yaml:"etcd_config"`
	serviceEndpoint        string
}

func (s *ServiceConfig) GetServiceEndpoint() string {
	return s.serviceEndpoint
}

// LoadConfig 加载ServiceConfiguration的配置
func loadServiceConfig(ctx context.Context, errorService errors.Service, logger *zap.Logger) (*ServiceConfig, string, errors.Error) {
	config := &ServiceConfig{}
	loaded := false
	if !boot.IsProductModel() && *configPath == "./config.yml" {
		confPath := fmt.Sprintf("./config-%s.yml", boot.RunModel())
		err := bootutils.LoadConfig(ctx, confPath, config, errorService, logger)
		if err == nil {
			*configPath = confPath
			loaded = true
		}
	}
	if !loaded {
		err := bootutils.LoadConfig(ctx, *configPath, config, errorService, logger)
		if err != nil {
			// return nil, "", err
			logger.Warn("没有找到配置文件")
			*configPath = ""
		} else {
			loaded = true
		}
	}
	if loaded {
		logger.Debug("加载配置完成", zap.Any("config", config))
	}
	serviceEndpoint, ok := os.LookupEnv("ENV_SERVICE_ENDPOINT")
	if !ok {
		serviceEndpoint = *serviceEndpointFlag
	}
	serviceEndpoint, err := bootutils.WarpServerAddr(serviceEndpoint, errorService)
	if err != nil {
		return nil, "", err
	}
	config.serviceEndpoint = serviceEndpoint
	if !config.SingleService {
		config.SingleService = *singleService
	}
	return config, *configPath, nil
}
