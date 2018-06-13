package main

import (
	"context"
	"fmt"
	"time"

	"git.xiagaogao.com/coffee/boot"
	"git.xiagaogao.com/coffee/boot/errors"
	"git.xiagaogao.com/coffee/boot/logs"
	"git.xiagaogao.com/coffee/boot/serviceboot"
	"git.xiagaogao.com/coffee/boot/simple"
	"git.xiagaogao.com/coffee/boot/simple/simplemodel"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func main() {
	serviceboot.ServiceLaunch(context.TODO(), &serviceImpl{})
}

type serviceImpl struct {
	rpcService *simple.RPCService
	logger     *zap.Logger
}

func (impl *serviceImpl) Init(cxt context.Context, serviceKit serviceboot.ServiceKit) errors.Error {
	impl.rpcService = &simple.RPCService{}
	serviceKit.RPCServiceInitialization(impl.rpcService)
	impl.logger = serviceKit.GetLogger()
	return nil
}
func (impl *serviceImpl) Run(cxt context.Context) errors.Error {
	go func() {
		client := impl.rpcService.GetClient()
		for i := int64(0); i < 100000000; i++ {
			resp, err := client.SayHello(cxt, &simplemodel.Request{fmt.Sprintf("coffee-%d", i), i})
			if err != nil {
				impl.logger.Error(fmt.Sprintf("全程调用异常%#v", err), logs.F_ExtendData(err))
				continue
			}
			impl.logger.Debug(resp.GetMessage())
			time.Sleep(time.Second * 5)
		}
	}()
	return nil
}
func (impl *serviceImpl) Stop(cxt context.Context) errors.Error {
	return nil
}
func (impl *serviceImpl) RegisterServer(s *grpc.Server) errors.Error {
	return nil
}
func (impl *serviceImpl) GetServiceInfo() boot.ServiceInfo {
	return boot.ServiceInfo{
		ServiceName: "simple_client",
		Version:     "0.0.1",
		Descriptor:  "测试客户端",
		APIDefine:   "",
		Scheme:      "http",
	}
}
