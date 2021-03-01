package consul

import (
	"git.xiagaogao.com/coffee/base/errors"
	"git.xiagaogao.com/coffee/base/log"
	"github.com/hashicorp/consul/api"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type serviceImpl struct {
	client *api.Client
}

func newService() *serviceImpl {
	config := &Config{}
	err := viper.UnmarshalKey("consul", config)
	if err != nil {
		log.Panic("加载consul配置失败", zap.Error(err))
	}
	if config.Address == "" {
		config.Address = viper.GetString("consul.address")
		if config.Address == "" {
			config.Address = "127.0.0.1:8500"
		}
	}
	if config.Token == "" {
		config.Token = viper.GetString("consul.token")
	}
	if config.Datacenter == "" {
		config.Datacenter = viper.GetString("consul.datacenter")
	}
	if config.Namespace == "" {
		config.Namespace = viper.GetString("consul.namespace")
	}
	log.Info("连接consul服务器", zap.Any("config", config))
	client, err := api.NewClient(&api.Config{
		Datacenter: config.Datacenter,
		Token:      config.Token,
		Address:    config.Address,
		Namespace:  config.Namespace,
		TokenFile:  config.TokenFile,
		WaitTime:   config.WaitTime,
	})
	if err != nil {
		log.Panic("启动consul客户端失败", zap.Error(err))
	}
	impl := &serviceImpl{
		client: client,
	}
	return impl
}

func (impl *serviceImpl) GetConsulClient() *api.Client {
	return impl.client
}

func (impl *serviceImpl) destroy() errors.Error {
	return nil
}
