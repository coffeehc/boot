package main

import (
	"context"
	"fmt"
	"time"

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
	logService, _ := logs.NewService()
	ctx = logs.SetLoggerService(ctx, logService)
	logger := logService.GetLogger()
	ctx = logs.SetLogger(ctx, logger)
	errorService := errors.NewService("simple")
	ctx = errors.SetRootErrorService(ctx, errorService)
	serviceInfo := boot.NewSimpleServiceInfo("simple_service", "0.0.1", "dev", "http", "", "")
	serviceInfo_client := boot.NewSimpleServiceInfo("simple_client", "0.0.1", "dev", "http", "", "")
	etcdClient, err := etcdsd.NewClient(ctx, &etcdsd.Config{
		Endpoints:        []string{"127.0.0.1:2379"},
		AutoSyncInterval: 5,
		DialTimeout:      3,
	})
	if err != nil {
		logger.Error(err.Error(), err.GetFields()...)
	}
	ctx = boot.SetEtcdClient(ctx, etcdClient)
	grpcFactory := grpcclient.NewGRPCConnFactory(ctx, serviceInfo_client)
	grpcConn, err := grpcFactory.NewClientConn(ctx, serviceInfo, false)
	if err != nil {
		logger.Error(err.Error(), err.GetFields()...)
	}
	greeterClient := simplemodel.NewGreeterClient(grpcConn)
	request := new(simplemodel.Request)
	request.Name = time.Now().String()
	response, err1 := greeterClient.SayHello(context.Background(), request)
	if err1 != nil {

		logger.Error(err1.Error())
		time.Sleep(time.Millisecond * 300)
		return
	}
	logger.Debug(fmt.Sprintf("response is %s", response.Message))
	time.Sleep(time.Millisecond * 300)
}
