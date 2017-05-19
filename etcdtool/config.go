package etcdtool

import (
	"crypto/tls"
	"time"

	"context"

	"github.com/coffeehc/microserviceboot/base"
	"github.com/coreos/etcd/clientv3"
)

type Config struct {
	Endpoints        []string `yaml:"endpoints"`
	AutoSyncInterval int64    `yaml:"auto_sync_interval"`
	DialTimeout      int64    `yaml:"dial_timeout"`
	Username         string   `yaml:"username"`
	Password         string   `yaml:"password"`
}

func (config *Config) GetEtcdConfig() (*clientv3.Config, base.Error) {
	if len(config.Endpoints) == 0 {
		return nil, base.NewError(base.ErrCodeBaseSystemInit, "etdc", "没有指定对应的Endpoints")
	}
	return &clientv3.Config{
		Endpoints:        config.Endpoints,
		AutoSyncInterval: config.getAutoSyncInterval(),
		DialTimeout:      config.getDialTimeout(),
		//TLS:              config.getTLS(),
		RejectOldCluster: false,
		DialOptions:      nil,
		Context:          context.Background(),
		Username:         config.Username,
		Password:         config.Password,
	}, nil
}

func (config *Config) getAutoSyncInterval() time.Duration {
	if config.AutoSyncInterval == 0 {
		config.AutoSyncInterval = 5
	}
	return time.Duration(config.AutoSyncInterval) * time.Second
}
func (config *Config) getDialTimeout() time.Duration {
	if config.DialTimeout == 0 {
		config.DialTimeout = 3
	}
	return time.Duration(config.DialTimeout) * time.Second
}

func (config *Config) getTLS() *tls.Config {
	return &tls.Config{
		InsecureSkipVerify: true,
	}
}
