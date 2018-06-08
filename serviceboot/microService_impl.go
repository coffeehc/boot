package serviceboot

import (
	"context"
	"fmt"
	"net"

	"git.xiagaogao.com/coffee/boot"
	"git.xiagaogao.com/coffee/boot/errors"
	"git.xiagaogao.com/coffee/boot/logs"
	"git.xiagaogao.com/coffee/boot/sd/etcdsd"
	"git.xiagaogao.com/coffee/boot/transport/grpcserver"
	"github.com/coreos/etcd/clientv3"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func newMicroService(ctx context.Context, service Service) (MicroService, errors.Error) {
	errorService := errors.GetRootErrorService(ctx).NewService("boot")
	if service == nil {
		return nil, errorService.BuildMessageError("没有service实例")
	}
	return &grpcMicroServiceImpl{
		service:      service,
		cleanFuncs:   make([]func(), 0),
		errorService: errorService,
	}, nil
}

type grpcMicroServiceImpl struct {
	errorService  errors.Service
	service       Service
	config        *grpcserver.GRPCConfig
	grpcServer    *grpc.Server
	cleanFuncs    []func()
	listener      net.Listener
	serviceConfig *ServiceConfig
	etcdClient    *clientv3.Client
}

func (ms *grpcMicroServiceImpl) Start(ctx context.Context, serviceConfig *ServiceConfig, configPath string) (err errors.Error) {
	ms.serviceConfig = serviceConfig
	logger := logs.GetLogger(ctx)
	server, err := grpcserver.NewServer(ctx, configPath)
	if err != nil {
		return err
	}
	ctx = boot.SetGRPCServer(ctx, server)
	etcdClient, err := etcdsd.NewClient(ctx, serviceConfig.EtcdConfig)
	if err != nil {
		return err
	}
	ctx = boot.SetEtcdClient(ctx, etcdClient)
	err = ms.service.Init(ctx, configPath, serviceConfig)
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
			err = ms.errorService.BuildSystemError(fmt.Sprintf("出现严重异常:%#v", err1))
		}
	}()
	err = ms.service.Run(ctx)
	if err != nil {
		return err
	}
	if err != nil {
		return err
	}
	ms.service.RegisterServer(server)
	lis, err1 := net.Listen("tcp", serviceConfig.ServerAddr)
	if err1 != nil {
		return ms.errorService.BuildWappedSystemError(err1)
	}
	ms.listener = lis
	go ms.grpcServer.Serve(lis)
	logger.Debug("服务已正常启动")
	//注册服务
	return etcdsd.RegisterService(ctx, etcdClient, serviceConfig.ServiceInfo, serviceConfig.ServerAddr)
}

func (ms *grpcMicroServiceImpl) AddCleanFunc(f func()) {
	ms.cleanFuncs = append(ms.cleanFuncs, f)
}

func (ms *grpcMicroServiceImpl) Stop(ctx context.Context) {
	if ms.etcdClient != nil {
		ms.etcdClient.Close()
	}
	logger := logs.GetLogger(ctx)
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
			logger.Error(stopErr.Error(), stopErr.GetFields()...)
		}
	}
	for _, f := range ms.cleanFuncs {
		func() {
			defer func() {
				if err := recover(); err != nil {
					logger.Error("clean func painc", zap.Any(logs.K_Cause, err))
				}
			}()
			f()
		}()
	}
}

func (ms *grpcMicroServiceImpl) GetService() Service {
	return ms.service
}

func (ms *grpcMicroServiceImpl) GetServiceInfo() boot.ServiceInfo {
	return ms.serviceConfig.ServiceInfo
}
