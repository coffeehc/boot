package main

import (
	"context"
	"time"

	"git.xiagaogao.com/baseservices/baseservice/baseservicecommon"
	"git.xiagaogao.com/coffee/boot"
	"git.xiagaogao.com/coffee/boot/errors"
	"git.xiagaogao.com/coffee/boot/logs"
	"git.xiagaogao.com/coffee/boot/sd/etcdsd"
	"git.xiagaogao.com/coffee/boot/simple/simplemodel"
	"git.xiagaogao.com/coffee/boot/transport/grpcclient"
)

func main() {
	//grpclog.SetLoggerV2(grpclog.NewLoggerV2(os.Stdout,os.Stdout,os.Stdout))
	ctx := context.TODO()
	serviceInfo_client := boot.ServiceInfo{"simple_client", "0.0.1", "", "", "http"}
	logService, _ := logs.NewService(serviceInfo_client)
	ctx = logs.SetLoggerService(ctx, logService)
	logger := logService.GetLogger()
	ctx = logs.SetLogger(ctx, logger)
	errorService := errors.NewService("simple")
	serviceInfo := boot.ServiceInfo{"baseservice", "0.0.1", "", "", "http"}
	serviceInfo_server := boot.ServiceInfo{"simple_service", "0.0.1", "", "", "http"}
	etcdClient, err := etcdsd.NewClient(ctx, &etcdsd.Config{
		Endpoints:        []string{"127.0.0.1:2379"},
		AutoSyncInterval: 5,
		DialTimeout:      3,
	}, errorService, logger)
	if err != nil {
		logger.Error(err.Error(), err.GetFields()...)
	}
	ctx = boot.SetEtcdClient(ctx, etcdClient)
	grpcFactory := grpcclient.NewGRPCConnFactory(ctx, etcdClient, serviceInfo_client, errorService, logger)
	grpcConn, err := grpcFactory.NewClientConn(ctx, serviceInfo, false)
	if err != nil {
		logger.Error(err.Error(), err.GetFields()...)
	}
	sequenceServiceClient := baseservicecommon.NewSequenceServiceClient(grpcConn)
	simplemodelGrpcConn, err := grpcFactory.NewClientConn(ctx, serviceInfo_server, false)
	if err != nil {
		logger.Error(err.Error(), err.GetFields()...)
	}
	resp, err1 := sequenceServiceClient.GenerateSequence(ctx, &baseservicecommon.SequenceGenerate{})
	if err1 != nil {
		logger.Error(err1.Error())
		time.Sleep(time.Millisecond * 300)
		return
	}
	logger.Debug("result", logs.F_ExtendData(resp))
	client1 := simplemodel.NewGreeterClient(simplemodelGrpcConn)
	request := new(simplemodel.Request)
	request.Name = time.Now().String()
	resp1, err1 := client1.SayHello(ctx, request)
	if err1 != nil {
		logger.Error(err1.Error())
		time.Sleep(time.Millisecond * 300)
		return
	}
	logger.Debug("result", logs.F_ExtendData(resp1))
}
