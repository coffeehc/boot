package etcd

import (
	"context"
	"crypto/tls"
	"time"

	"github.com/coreos/etcd/clientv3"
)

type Config struct {
	Endpoints        []string
	AutoSyncInterval time.Duration
	DialTimeout      time.Duration
	Username         string
	Password         string
}

func (config *Config) getEtcdConfig() *clientv3.Config {
	return &clientv3.Config{
		Endpoints:        config.Endpoints,
		AutoSyncInterval: config.getAutoSyncInterval(),
		DialTimeout:      config.getDialTimeout(),
		// TLS:              config.getTLS(),
		DialKeepAliveTime:    time.Second * 60,
		DialKeepAliveTimeout: time.Second * 90,
		RejectOldCluster:     false,
		DialOptions:          nil,
		Context:              context.Background(),
		Username:             config.Username,
		Password:             config.Password,
	}
}

func (config *Config) getAutoSyncInterval() time.Duration {
	if config.AutoSyncInterval == 0 {
		config.AutoSyncInterval = 5 * time.Second
	}
	return config.AutoSyncInterval
}
func (config *Config) getDialTimeout() time.Duration {
	if config.DialTimeout == 0 {
		config.DialTimeout = 3 * time.Second
	}
	return config.DialTimeout
}

func (config *Config) getTLS() *tls.Config {
	return &tls.Config{
		InsecureSkipVerify: true,
	}
}
