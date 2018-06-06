package etcdtool

import (
	"git.xiagaogao.com/coffee/boot/errors"
	"github.com/coreos/etcd/clientv3"
)

func NewClientByConfigFile(configPath string) (*clientv3.Client, errors.Error) {
	config, err := LoadEtcdConfig(configPath)
	if err != nil {
		return nil, err
	}
	return NewClient(config)
}

func NewClient(config *Config) (*clientv3.Client, errors.Error) {
	conf, err := config.GetEtcdConfig()
	if err != nil {
		return nil, err
	}
	etcdClient, _err := clientv3.New(*conf)
	if _err != nil {
		return nil, errors.NewErrorWrapper(errors.Error_System, "etcd", _err)
	}
	return etcdClient, nil
}
