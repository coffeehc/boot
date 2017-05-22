package etcdtool

import (
	"context"
	"crypto/tls"
	"io/ioutil"
	"time"

	"github.com/coffeehc/logger"
	"github.com/coffeehc/microserviceboot/base"
	"github.com/coreos/etcd/clientv3"
	yaml "gopkg.in/yaml.v2"
)

func LoadEtcdConfig(configPath string) (*Config, base.Error) {
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, base.NewError(base.ErrCodeBaseSystemInit, "etcd", "加载配置文件失败")
	}
	i := &struct {
		EtcdConfig *Config `yaml:"etcd"`
	}{}
	err = yaml.Unmarshal(data, i)
	if err != nil {
		return nil, base.NewError(base.ErrCodeBaseSystemInit, "etcd", "解析Etcd配置失败")
	}
	if i.EtcdConfig == nil {
		return nil, base.NewError(base.ErrCodeBaseSystemInit, "etcd", "加载的Etcd配置为空")
	}
	logger.Debug("读取配置文件内容为:%#v", i.EtcdConfig)
	return i.EtcdConfig, nil

}

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
