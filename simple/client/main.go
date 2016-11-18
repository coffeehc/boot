package main

import (
	"github.com/coffeehc/logger"
	"github.com/coffeehc/microserviceboot/base"
	"github.com/coffeehc/microserviceboot/consultool"
	"github.com/coffeehc/microserviceboot/serviceclient/grpcclient"
	"github.com/coffeehc/microserviceboot/simple/simplemodel"
	"golang.org/x/net/context"
	"time"
)

func main() {
	logger.InitLogger()
	serviceInfo := base.NewSimpleServiceInfo("simple_service", "0.0.1", "https", "dev", "", "")
	var e error
	consulClient, err := consultool.NewConsulClient(nil)
	if err != nil {
		e = err
		logger.Error("%s", e)
		time.Sleep(time.Millisecond * 300)
		return
	}
	balanacer, err := consultool.NewConsulBalancer(consulClient, serviceInfo)
	if err != nil {
		e = err
		logger.Error("%s", e)
		time.Sleep(time.Millisecond * 300)
		return
	}
	client := grpcclient.NewGrpcClient()
	clientConn, err := client.NewClientConn(serviceInfo, balanacer)
	if err != nil {
		e = err
		logger.Error("%s", e)
		time.Sleep(time.Millisecond * 300)
		return
	}
	greeterClient := simplemodel.NewGreeterClient(clientConn)
	request := new(simplemodel.Request)
	request.Name = time.Now().String()
	response, err1 := greeterClient.SayHello(context.Background(), request)
	if err1 != nil {
		e = base.NewErrorWrapper(err1)
		logger.Error("%s", e)
		time.Sleep(time.Millisecond * 300)
		return
	}
	logger.Debug("response is %s", response.Message)
}
