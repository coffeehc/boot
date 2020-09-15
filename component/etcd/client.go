package etcd

import (
	"context"
	"sync"
	"time"

	"git.xiagaogao.com/coffee/base/errors"
	"git.xiagaogao.com/coffee/base/log"
	"github.com/spf13/viper"
	"go.etcd.io/etcd/clientv3"
	"go.uber.org/zap"
)

var defaultEtcdClient *clientv3.Client
var mutex = new(sync.Mutex)
var scope = zap.String("scope", "etcd.client")

func GetEtcdClient() *clientv3.Client {
	if defaultEtcdClient == nil {
		InitEtcdClient()
	}
	return defaultEtcdClient
}

func InitEtcdClient() {
	mutex.Lock()
	defer mutex.Unlock()
	if defaultEtcdClient != nil {
		return
	}
	if !viper.IsSet("etcd") {
		log.Fatal("没有配置etcd")
	}
	config := &Config{}
	err := viper.UnmarshalKey("etcd", config)
	if err != nil {
		log.Fatal("加载etcd配置失败", zap.Error(err), scope)
	}
	if len(config.Endpoints) == 0 {
		log.Fatal("没有设置EtcdServer地址", scope)
	}
	if config.AutoSyncInterval == 0 {
		config.AutoSyncInterval = 5 * time.Second
	}
	if config.DialTimeout == 0 {
		config.DialTimeout = time.Second * 3
	}
	client, err1 := newClient(config)
	if err1 != nil {
		log.Fatal("创建Etcd客户端失败", err1.GetFieldsWithCause(scope)...)
	}
	defaultEtcdClient = client

}

func newClient(config *Config) (*clientv3.Client, errors.Error) {
	ctx := context.TODO()
	if config == nil {
		return nil, errors.SystemError("EtcdConfig为空")
	}
	conf := config.getEtcdConfig()
	etcdClient, _err := clientv3.New(*conf)
	if _err != nil {
		return nil, errors.WrappedSystemError(_err)
	}
	ctx, _ = context.WithTimeout(ctx, time.Second*3)
	_err = etcdClient.Sync(ctx)
	if _err != nil {
		log.Error("同步Etcd失败", zap.Error(_err), scope)
		return nil, errors.SystemError("同步etcd失败")
	}
	log.Debug("初始化EtcdClient", zap.Strings("endpoints", etcdClient.Endpoints()), scope)
	return etcdClient, nil
}
