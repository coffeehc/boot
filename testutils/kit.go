package testutils

import (
	"context"

	"git.xiagaogao.com/coffee/boot/base/errors"
	"git.xiagaogao.com/coffee/boot/bootutils"
	"git.xiagaogao.com/coffee/boot/discover/etcdsd"
	"git.xiagaogao.com/coffee/boot/serviceboot"
	"git.xiagaogao.com/coffee/boot/transport/grpc/grpcclient"
	"git.xiagaogao.com/coffee/boot/xerror"
	"git.xiagaogao.com/coffee/boot/xlog"
	"github.com/coreos/etcd/clientv3"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func BuildServiceKit(testName string, etcdEndPoints []string) (serviceboot.ServiceKit, errors.Error) {
	ctx := context.TODO()
	boot.InitModel()
	serviceInfo := &boot.ServiceInfo{
		ServiceName: testName,
		Version:     "0.0.1",
		Descriptor:  "测试服务-" + testName,
		APIDefine:   "",
		Scheme:      "http",
	}
	logService, err1 := log.NewService(serviceInfo)
	if err1 != nil {
		panic(err1.Error())
	}
	logger, _ := zap.NewDevelopment(zap.AddStacktrace(zapcore.DPanicLevel))
	errorService := errors.NewService(testName, logger)
	etcdClient, err := etcdsd.newClient(ctx, &etcdsd.Config{
		Endpoints:   etcdEndPoints,
		DialTimeout: int64(3),
	}, errorService, logger)
	if err != nil {
		logger.Error("初始化注册中心失败", log.F_Error(err))
		return nil, err
	}
	logger.Debug("Etcd客户端初始化完成")
	grpcConnFactory := grpcclient.NewGRPCConnFactory(etcdClient, errorService, logger)

	return &serviceKitImpl{
		logger:          logger,
		errorService:    errorService,
		loggerService:   logService,
		etcdClient:      etcdClient,
		serviceInfo:     serviceInfo,
		grpcConnFactory: grpcConnFactory,
		ctx:             ctx,
		serverAddr:      "",
		configPath:      "",
	}, nil
}

type serviceKitImpl struct {
	logger          *zap.Logger
	errorService    errors.Service
	loggerService   log.Service
	etcdClient      *clientv3.Client
	serviceInfo     *boot.ServiceInfo
	grpcConnFactory grpcclient.GRPCConnFactory
	ctx             context.Context
	serverAddr      string
	configPath      string
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
func (impl *serviceKitImpl) GetLoggerService() log.Service {
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
func (impl *serviceKitImpl) RPCServiceInitialization(rpcService serviceboot.RPCService) errors.Error {
	errorService := impl.errorService.NewService("rpc")
	conn, err := impl.grpcConnFactory.NewClientConn(impl.ctx, rpcService.GetRPCServiceInfo(), false)
	if err != nil {
		return err
	}
	impl.logger.Debug("初始化RPCService", log.F_ExtendData(rpcService.GetRPCServiceInfo()))
	return rpcService.InitRPCService(impl.ctx, conn, errorService, impl.logger)
}

func (impl *serviceKitImpl) SetExtentData(data map[string]string) {
}

func (impl *serviceKitImpl) InitExtConfig(config interface{}) errors.Error {
	return bootutils.LoadConfig(impl.ctx, impl.configPath, config, impl.errorService, impl.logger)
}
