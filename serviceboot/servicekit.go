package serviceboot

// import (
// 	"context"
//
// 	"git.xiagaogao.com/coffee/boot"
// 	"git.xiagaogao.com/coffee/boot/base/errors"
// 	"git.xiagaogao.com/coffee/boot/bootutils"
// 	"git.xiagaogao.com/coffee/boot/logs"
// 	"git.xiagaogao.com/coffee/boot/transport/grpc/grpcclient"
// 	"github.com/coreos/etcd/clientv3"
// 	"go.uber.org/zap"
// )
//
// type ServiceKit interface {
// 	GetEtcdClient() *clientv3.Client
// 	GetServiceInfo() *boot.ServiceInfo
// 	GetServerAddr() string
// 	GetGRPCConnFactory() grpcclient.GRPCConnFactory
// 	GetContext() context.Context
// 	GetConfigPath() string
// 	RPCServiceInitialization(rpcService RPCService) xerror.Error
// 	SetExtentData(data map[string]string)
// 	InitExtConfig(config interface{}) xerror.Error
// }
//
// type serviceKitImpl struct {
// 	etcdClient      *clientv3.Client
// 	serviceInfo     *boot.ServiceInfo
// 	grpcConnFactory grpcclient.GRPCConnFactory
// 	ctx             context.Context
// 	serverAddr      string
// 	configPath      string
// 	extentData      map[string]string
// }
//
// func (impl *serviceKitImpl) SetExtentData(data map[string]string) {
// 	impl.extentData = data
// }
//
// func (impl *serviceKitImpl) GetConfigPath() string {
// 	return impl.configPath
// }
// func (impl *serviceKitImpl) GetEtcdClient() *clientv3.Client {
// 	return impl.etcdClient
// }
// func (impl *serviceKitImpl) GetServiceInfo() *boot.ServiceInfo {
// 	return impl.serviceInfo
// }
// func (impl *serviceKitImpl) GetServerAddr() string {
// 	return impl.serverAddr
// }
// func (impl *serviceKitImpl) GetGRPCConnFactory() grpcclient.GRPCConnFactory {
// 	return impl.grpcConnFactory
// }
// func (impl *serviceKitImpl) GetContext() context.Context {
// 	return impl.ctx
// }
// func (impl *serviceKitImpl) RPCServiceInitialization(rpcService RPCService) xerror.Error {
// 	errorService := xerror.GetErrorService().NewService("rpc")
// 	conn, err := impl.grpcConnFactory.NewClientConn(impl.ctx, rpcService.GetRPCServiceInfo(), false)
// 	if err != nil {
// 		return err
// 	}
// 	err = rpcService.InitRPCService(impl.ctx, conn, errorService)
// 	if err != nil {
// 		return err
// 	}
// 	xlog.Info("初始化RPCService成功", xlog.F_ExtendData(rpcService.GetRPCServiceInfo()))
// 	return nil
// }
//
// func (impl *serviceKitImpl) InitExtConfig(config interface{}) xerror.Error {
// 	return bootutils.LoadConfig(impl.ctx, impl.configPath, config)
// }
