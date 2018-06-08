package etcdsd

import (
	"context"
	"crypto/tls"
	"time"

	"os"
	"strings"

	"git.xiagaogao.com/coffee/boot/errors"
	"github.com/coreos/etcd/clientv3"
)

const envETCDEndpoints = "ENV_ETCD_ENDPOINTS"
const envETCDUsername = "ENV_ETCD_USERNAME"
const envETCDPassword = "ENV_ETCD_PASSWORD"

type Config struct {
	Endpoints        []string `yaml:"endpoints"`
	AutoSyncInterval int64    `yaml:"auto_sync_interval"`
	DialTimeout      int64    `yaml:"dial_timeout"`
	Username         string   `yaml:"username"`
	Password         string   `yaml:"password"`
}

func (config *Config) GetEtcdConfig() (*clientv3.Config, errors.Error) {
	if os.Getenv(envETCDUsername) != "" {
		config.Username = os.Getenv(envETCDUsername)
	}
	if os.Getenv(envETCDPassword) != "" {
		config.Password = os.Getenv(envETCDPassword)
	}
	env_endpoints := os.Getenv(envETCDEndpoints)
	if env_endpoints != "" {
		config.Endpoints = strings.Split(env_endpoints, ",")
	}
	if len(config.Endpoints) == 0 {
		return nil, errors.NewError(errors.Error_System, "etdc", "没有指定对应的Endpoints")
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
