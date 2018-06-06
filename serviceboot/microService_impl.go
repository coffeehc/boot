package serviceboot

import (
	"context"
	"fmt"
	"net"

	"git.xiagaogao.com/coffee/boot/errors"
	"git.xiagaogao.com/coffee/boot/logs"
	"git.xiagaogao.com/coffee/boot/transport/grpcserver"
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
}

func (ms *grpcMicroServiceImpl) Start(ctx context.Context, serviceConfig *ServiceConfig, configPath string) (err errors.Error) {
	ms.serviceConfig = serviceConfig
	logger := logs.GetLogger(ctx)
	server, err := grpcserver.NewServer(ctx, configPath)
	err = ms.service.Init(ctx, configPath, serviceConfig, server)
	if err != nil {
		return err
	}
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
	logger.Info("服务已正常启动")

	return nil
}

func (ms *grpcMicroServiceImpl) AddCleanFunc(f func()) {
	ms.cleanFuncs = append(ms.cleanFuncs, f)
}

func (ms *grpcMicroServiceImpl) Stop(ctx context.Context) {
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
			stopErr.PrintLog(ctx)
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

func (ms *grpcMicroServiceImpl) GetServiceInfo() ServiceInfo {
	return ms.serviceConfig.ServiceInfo
}
