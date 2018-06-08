package etcdsd

import (
	"context"

	"git.xiagaogao.com/coffee/boot/errors"
	"github.com/coreos/etcd/clientv3"
)

func NewClient(ctx context.Context, config *Config) (*clientv3.Client, errors.Error) {
	rootErrorService := errors.GetRootErrorService(ctx)
	errorService := rootErrorService.NewService("sd")
	conf, err := config.GetEtcdConfig()
	if err != nil {
		return nil, err
	}
	etcdClient, _err := clientv3.New(*conf)
	if _err != nil {
		return nil, errorService.BuildWappedSystemError(err)
	}
	return etcdClient, nil
}