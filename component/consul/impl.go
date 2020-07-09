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
	config := &api.Config{}
	err := viper.UnmarshalKey("consul", config)
	if err != nil {
		log.Fatal("加载consul配置失败", zap.Error(err))
	}
	client, err := api.NewClient(config)
	if err != nil {
		log.Fatal("启动consul客户端失败", zap.Error(err))
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
