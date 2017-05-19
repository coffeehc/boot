package etcdtool

import (
	"github.com/coffeehc/microserviceboot/base"
	"github.com/coreos/etcd/clientv3"
)

func NewClient(config *Config) (*clientv3.Client, base.Error) {
	conf, err := config.GetEtcdConfig()
	if err != nil {
		return nil, err
	}
	etcdClient, _err := clientv3.New(*conf)
	if _err != nil {
		return nil, base.NewErrorWrapper(base.ErrCodeBaseSystemInit, "etcd", _err)
	}
	return etcdClient, nil
}
