package boottools

import (
	"context"

	"git.xiagaogao.com/coffee/boot/errors"
	"git.xiagaogao.com/coffee/boot/sd/etcdsd"
	"git.xiagaogao.com/coffee/boot/serviceboot"
	"git.xiagaogao.com/coffee/boot/transport/grpcclient"
	"github.com/coreos/etcd/clientv3"
	"go.uber.org/zap"
)

func NewEtcdClientByHost(endpoints []string, errorService errors.Service, logger *zap.Logger) (*clientv3.Client, errors.Error) {
	return NewEtcdClient(&etcdsd.Config{
		Endpoints: endpoints,
	}, errorService, logger)
}

func NewEtcdClient(config *etcdsd.Config, errorService errors.Service, logger *zap.Logger) (*clientv3.Client, errors.Error) {
	if config == nil {
		config = &etcdsd.Config{
			Endpoints: []string{"127.0.0.1:2379"},
		}
	}
	return etcdsd.NewClient(context.TODO(), config, errorService, logger)
}

func RPCServiceInitialization(rpcService serviceboot.RPCService, serviceAddr string, errorService errors.Service, logger *zap.Logger) errors.Error {
	conn, err := grpcclient.NewClientConn(context.TODO(), errorService, logger, serviceAddr)
	if err != nil {
		return err
	}
	rpcService.InitRPCService(context.TODO(), conn, errorService, logger)
	return nil
}
