package etcdsd

import (
	"context"
	"time"

	"git.xiagaogao.com/coffee/boot/errors"
	"github.com/coreos/etcd/clientv3"
	"go.uber.org/zap"
)

func NewClient(ctx context.Context, config *Config, errorService errors.Service, logger *zap.Logger) (*clientv3.Client, errors.Error) {
	errorService = errorService.NewService("sd")
	if config == nil {
		logger.Debug("没有配置EtcdConfig,使用默认配置")
		config = &Config{}
	}
	conf, err := config.GetEtcdConfig(errorService)
	if err != nil {
		return nil, err
	}
	etcdClient, _err := clientv3.New(*conf)
	if _err != nil {
		return nil, errorService.WrappedSystemError(_err)
	}
	ctx, _ = context.WithTimeout(ctx, time.Second*3)
	_err = etcdClient.Sync(ctx)
	if _err != nil {
		return nil, errorService.SystemError("同步etcd失败", zap.Error(_err))
	}
	logger.Info("初始化EtcdClient", zap.Strings("endpoints", etcdClient.Endpoints()))
	return etcdClient, nil
}
