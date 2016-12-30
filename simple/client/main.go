package main

import (
	"time"

	"github.com/coffeehc/logger"
	"github.com/coffeehc/microserviceboot/base"
	"github.com/coffeehc/microserviceboot/consultool"
	"github.com/coffeehc/microserviceboot/serviceclient/grpcclient"
	"github.com/coffeehc/microserviceboot/simple/simplemodel"
	"golang.org/x/net/context"
)

func main() {
	logger.InitLogger()
	serviceInfo := base.NewSimpleServiceInfo("simple_service", "0.0.1", "dev", "https", "", "")
	var e error
	consulClient, err := consultool.NewClient(nil)
	if err != nil {
		e = err
		logger.Error("%s", e)
		time.Sleep(time.Millisecond * 300)
		return
	}
	balancerBuilder := consultool.NewConsulBalancerBuilder(consulClient)
	clientConnFactory := grpcclient.NewClientConnFactory(balancerBuilder)
	clientConn,err := clientConnFactory.GetClientConn(context.Background(),serviceInfo,0)
	if err!=nil{
		logger.Error("error is %s",err)
		time.Sleep(time.Millisecond*300)
		return
	}
	greeterClient := simplemodel.NewGreeterClient(clientConn)
	request := new(simplemodel.Request)
	request.Name = time.Now().String()
	response, err1 := greeterClient.SayHello(context.Background(), request)
	if err1 != nil {
		e = base.NewErrorWrapper("test",err1)
		logger.Error("%s", e)
		time.Sleep(time.Millisecond * 300)
		return
	}
	logger.Debug("response is %s", response.Message)
	time.Sleep(time.Millisecond * 300)
}
