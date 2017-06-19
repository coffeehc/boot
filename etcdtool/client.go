package etcdtool

import (
	"github.com/coffeehc/microserviceboot/base"
	"github.com/coreos/etcd/clientv3"
)

func NewClientByConfigFile(configPath string) (*clientv3.Client, base.Error) {
	config, err := LoadEtcdConfig(configPath)
	if err != nil {
		return nil, err
	}
	return NewClient(config)
}

func NewClient(config *Config) (*clientv3.Client, base.Error) {
	conf, err := config.GetEtcdConfig()
	if err != nil {
		return nil, err
	}
	etcdClient, _err := clientv3.New(*conf)
	if _err != nil {
		return nil, base.NewErrorWrapper(base.ErrCode_System, "etcd", _err)
	}
	return etcdClient, nil
}
