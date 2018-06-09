package serviceboot

import (
	"context"

	"git.xiagaogao.com/coffee/boot"
	"git.xiagaogao.com/coffee/boot/errors"
	"git.xiagaogao.com/coffee/boot/logs"
	"git.xiagaogao.com/coffee/boot/transport/grpcclient"
	"github.com/coreos/etcd/clientv3"
	"go.uber.org/zap"
)

type ServiceKit interface {
	GetLogger() *zap.Logger
	GetRootErrorService() errors.Service
	GetLoggerService() logs.Service
	GetEtcdClient() *clientv3.Client
	GetServiceInfo() boot.ServiceInfo
	GetServerAddr() string
	GetGRPCConnFactory() grpcclient.GRPCConnFactory
	GetContext() context.Context
	GetConfigPath() string
	GetRPCServiceInitialization() RPCServiceInitialization
}

type serviceBootImpl struct {
	logger                   *zap.Logger
	errorService             errors.Service
	loggerService            logs.Service
	etcdClient               *clientv3.Client
	serviceInfo              boot.ServiceInfo
	grpcClientFactory        grpcclient.GRPCConnFactory
	ctx                      context.Context
	serverAddr               string
	configPath               string
	rpcServiceInitialization RPCServiceInitialization
}

func (impl *serviceBootImpl) GetConfigPath() string {
	return impl.configPath
}

func (impl *serviceBootImpl) GetLogger() *zap.Logger {
	return impl.logger
}
func (impl *serviceBootImpl) GetRootErrorService() errors.Service {
	return impl.errorService
}
func (impl *serviceBootImpl) GetLoggerService() logs.Service {
	return impl.loggerService
}
func (impl *serviceBootImpl) GetEtcdClient() *clientv3.Client {
	return impl.etcdClient
}
func (impl *serviceBootImpl) GetServiceInfo() boot.ServiceInfo {
	return impl.serviceInfo
}
func (impl *serviceBootImpl) GetServerAddr() string {
	return impl.serverAddr
}
func (impl *serviceBootImpl) GetGRPCConnFactory() grpcclient.GRPCConnFactory {
	return impl.grpcClientFactory
}
func (impl *serviceBootImpl) GetContext() context.Context {
	return impl.ctx
}

func (impl *serviceBootImpl) GetRPCServiceInitialization() RPCServiceInitialization {
	return impl.rpcServiceInitialization
}
