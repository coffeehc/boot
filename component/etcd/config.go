package etcd

import (
	"context"
	"crypto/tls"
	"time"

	"go.etcd.io/etcd/clientv3"
)

type Config struct {
	Endpoints        []string      `mapstructure:"endpoints,omitempty" json:"endpoints,omitempty"`
	AutoSyncInterval time.Duration `mapstructure:"auto_sync_interval,omitempty" json:"auto_sync_interval,omitempty"`
	DialTimeout      time.Duration `mapstructure:"dial_timeout,omitempty" json:"dial_timeout,omitempty"`
	Username         string        `mapstructure:"username,omitempty" json:"username,omitempty"`
	Password         string        `mapstructure:"password,omitempty" json:"password,omitempty"`
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
