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
	GetServiceInfo() *boot.ServiceInfo
	GetServerAddr() string
	GetGRPCConnFactory() grpcclient.GRPCConnFactory
	GetContext() context.Context
	GetConfigPath() string
	RPCServiceInitialization(rpcService RPCService) errors.Error
	SetExtentData(data map[string]string)
}

type serviceKitImpl struct {
	logger          *zap.Logger
	errorService    errors.Service
	loggerService   logs.Service
	etcdClient      *clientv3.Client
	serviceInfo     *boot.ServiceInfo
	grpcConnFactory grpcclient.GRPCConnFactory
	ctx             context.Context
	serverAddr      string
	configPath      string
	extentData      map[string]string
}

func (impl *serviceKitImpl) SetExtentData(data map[string]string) {
	impl.extentData = data
}

func (impl *serviceKitImpl) GetConfigPath() string {
	return impl.configPath
}

func (impl *serviceKitImpl) GetLogger() *zap.Logger {
	return impl.logger
}
func (impl *serviceKitImpl) GetRootErrorService() errors.Service {
	return impl.errorService
}
func (impl *serviceKitImpl) GetLoggerService() logs.Service {
	return impl.loggerService
}
func (impl *serviceKitImpl) GetEtcdClient() *clientv3.Client {
	return impl.etcdClient
}
func (impl *serviceKitImpl) GetServiceInfo() *boot.ServiceInfo {
	return impl.serviceInfo
}
func (impl *serviceKitImpl) GetServerAddr() string {
	return impl.serverAddr
}
func (impl *serviceKitImpl) GetGRPCConnFactory() grpcclient.GRPCConnFactory {
	return impl.grpcConnFactory
}
func (impl *serviceKitImpl) GetContext() context.Context {
	return impl.ctx
}
func (impl *serviceKitImpl) RPCServiceInitialization(rpcService RPCService) errors.Error {
	errorService := impl.errorService.NewService("rpc")
	conn, err := impl.grpcConnFactory.NewClientConn(impl.ctx, rpcService.GetRPCServiceInfo(), false)
	if err != nil {
		return err
	}
	impl.logger.Debug("初始化RPCService", logs.F_ExtendData(rpcService.GetRPCServiceInfo()))
	return rpcService.InitRPCService(impl.ctx, conn, errorService, impl.logger)

}
