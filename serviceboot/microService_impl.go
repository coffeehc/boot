package serviceboot

import (
	"context"
	"fmt"
	"net"

	"git.xiagaogao.com/base/cloudcommons/utils"
	"git.xiagaogao.com/coffee/boot"
	"git.xiagaogao.com/coffee/boot/errors"
	"git.xiagaogao.com/coffee/boot/logs"
	"git.xiagaogao.com/coffee/boot/manage"
	"git.xiagaogao.com/coffee/boot/sd/etcdsd"
	"git.xiagaogao.com/coffee/boot/transport/grpcclient"
	"git.xiagaogao.com/coffee/boot/transport/grpcserver"
	"github.com/coreos/etcd/clientv3"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func newMicroService(ctx context.Context, service Service, serviceInfo *boot.ServiceInfo, configPath string, errorService errors.Service, logger *zap.Logger, loggerService logs.Service) (MicroService, errors.Error) {
	errorService = errorService.NewService("boot")
	if service == nil {
		return nil, errorService.MessageError("没有service实例")
	}
	managerService, err := manage.NewManageService(serviceInfo, errorService, logger)
	if err != nil {
		return nil, err
	}
	return &grpcMicroServiceImpl{
		managerService: managerService,
		service:        service,
		cleanFuncs:     make([]func(), 0),
		errorService:   errorService,
		logger:         logger,
		configPath:     configPath,
		loggerService:  loggerService,
		serviceInfo:    serviceInfo,
	}, nil
}

type grpcMicroServiceImpl struct {
	managerService manage.Service
	errorService   errors.Service
	service        Service
	config         *grpcserver.GRPCConfig
	grpcServer     *grpc.Server
	cleanFuncs     []func()
	listener       net.Listener
	serviceConfig  *ServiceConfig
	etcdClient     *clientv3.Client
	logger         *zap.Logger
	configPath     string
	loggerService  logs.Service
	serviceInfo    *boot.ServiceInfo
}

func (ms *grpcMicroServiceImpl) Start(ctx context.Context, serviceConfig *ServiceConfig) (err errors.Error) {
	ms.serviceConfig = serviceConfig
	serviceInfo := ms.serviceInfo
	server, err := grpcserver.NewServer(ctx, serviceConfig.GrpcConfig, serviceInfo, ms.errorService, ms.logger)
	if err != nil {
		return err
	}
	var grpcConnFactory grpcclient.GRPCConnFactory = nil
	var etcdClient *clientv3.Client = nil
	if ms.serviceConfig.SingleService {
		serviceConfig.DisableServiceRegister = true
	} else {
		etcdClient, err = etcdsd.NewClient(ctx, serviceConfig.EtcdConfig, ms.errorService, ms.logger)
		if err != nil {
			return err
		}
		grpcConnFactory = grpcclient.NewGRPCConnFactory(etcdClient, ms.errorService, ms.logger)
	}
	serviceBoot := &serviceKitImpl{
		logger:          ms.logger,
		errorService:    ms.errorService,
		loggerService:   ms.loggerService,
		etcdClient:      etcdClient,
		serviceInfo:     serviceInfo,
		grpcConnFactory: grpcConnFactory,
		ctx:             ctx,
		serverAddr:      serviceConfig.GetServiceEndpoint(),
		configPath:      ms.configPath,
	}
	ms.managerService.Start(nil)
	err = ms.service.Init(ctx, serviceBoot)
	if err != nil {
		return err
	}
	ms.etcdClient = etcdClient
	ms.grpcServer = server
	defer func() {
		if err1 := recover(); err1 != nil {
			if e, ok := err1.(errors.Error); ok {
				err = e
				return
			}
			utils.Panic(ms.errorService.SystemError(fmt.Sprintf("出现严重异常:%#v", err1)), ms.logger)
		}
	}()
	err = ms.service.Run(ctx)
	if err != nil {
		return err
	}
	ms.service.RegisterServer(server)
	lis, err1 := net.Listen("tcp4", serviceConfig.GetServiceEndpoint())
	if err1 != nil {
		return ms.errorService.SystemError("启动RPC服务端口失败，请设置：ENV_SERVICE_ENDPOINT", zap.Error(err1))
	}
	serviceConfig.serviceEndpoint = lis.Addr().String()
	ms.logger.Debug("服务地址", zap.String("serviceEndpoint", lis.Addr().String()))
	ms.listener = lis
	go ms.grpcServer.Serve(lis)
	ms.logger.Debug("服务已正常启动")
	if serviceConfig.DisableServiceRegister {
		return nil
	}
	// 注册服务
	ms.logger.Debug("开始注册服务")
	return etcdsd.RegisterService(ctx, etcdClient, serviceInfo, serviceConfig.GetServiceEndpoint(), ms.managerService.GetEndpoint(), serviceBoot.extentData, ms.errorService, ms.logger)
}

func (ms *grpcMicroServiceImpl) AddCleanFunc(f func()) {
	ms.cleanFuncs = append(ms.cleanFuncs, f)
}

func (ms *grpcMicroServiceImpl) Stop(ctx context.Context) {
	if ms.etcdClient != nil {
		ms.etcdClient.Close()
	}
	if ms.grpcServer != nil {
		ms.grpcServer.GracefulStop()
	}
	if ms.listener != nil {
		ms.listener.Close()
	}
	service := ms.service
	if service != nil && service.Stop != nil {
		stopErr := service.Stop(ctx)
		if stopErr != nil {
			ms.logger.DPanic(stopErr.Error(), stopErr.GetFields()...)
		}
	}
	for _, f := range ms.cleanFuncs {
		func() {
			defer func() {
				if err := recover(); err != nil {
					ms.logger.DPanic("clean func painc", zap.Any(logs.K_Cause, err))
				}
			}()
			f()
		}()
	}
	ms.managerService.GetHttpService().Shutdown()
}

func (ms *grpcMicroServiceImpl) GetService() Service {
	return ms.service
}
