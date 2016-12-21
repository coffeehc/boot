package grpcclient

import (
	"github.com/coffeehc/logger"
	"github.com/coffeehc/microserviceboot/base"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
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
	clientConn, err := this.clientConnFactory.GetClientConn(context.Background(), this.serviceInfo, 0)
	if err != nil {
		logger.Warn("grpc ClientConn is Close,%s", err)
		return err
	}
	newClient(clientConn)
	return nil
}
