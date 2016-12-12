package grpcclient

import (
	"github.com/coffeehc/logger"
	"github.com/coffeehc/microserviceboot/base"
	"google.golang.org/grpc"
	"golang.org/x/net/context"
	"time"
)

type ServiceClientBase struct {
	serviceInfo       base.ServiceInfo
	clientConnFactory ClientConnFactory
	
}

func (this *ServiceClientBase) Init(serviceInfo base.ServiceInfo, clientConnFactory ClientConnFactory) {
	this.serviceInfo = serviceInfo
	this.clientConnFactory = clientConnFactory
}

func (this *ServiceClientBase) ListenConn(newClient func(conn *grpc.ClientConn)) base.Error {
	cxt,cancel:=context.WithCancel(context.Background())
	clientConn, err := this.clientConnFactory.GetClientConn(cxt,this.serviceInfo, 0)
	go func() {
		<-cxt.Done()
		time.Sleep(time.Second)
		this.ListenConn(newClient)
	}()
	if err != nil {
		cancel()
		logger.Warn("grpc ClientConn is Close,%s", err)
		return err
	}
	newClient(clientConn)
	return nil
}
