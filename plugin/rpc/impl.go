package rpc

import (
	"context"
	"fmt"
	"net"
	"time"

	"git.xiagaogao.com/coffee/base/errors"
	"git.xiagaogao.com/coffee/base/log"
	"git.xiagaogao.com/coffee/boot/component/consul"
	"git.xiagaogao.com/coffee/boot/configuration"
	"github.com/hashicorp/consul/api"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

type serviceImpl struct {
	config        *RpcConfig
	server        *grpc.Server
	rpcServerAddr string
	healthServer  *health.Server
}

func (impl *serviceImpl) GetRPCServerAddr() string {
	if impl.server == nil {
		log.Panic("rpc server没有初始化")
	}
	return impl.rpcServerAddr
}

func (impl *serviceImpl) GetGRPCServer() *grpc.Server {
	return impl.server
}

func (impl *serviceImpl) Start(ctx context.Context) errors.Error {
	addr, _ := net.ResolveTCPAddr("tcp4", impl.rpcServerAddr)
	addr.IP = net.IPv4zero
	lis, _err := net.Listen("tcp4", addr.String())
	if _err != nil {
		log.Panic("启动RPC服务端口失败", zap.Error(_err))
	}
	go func() {
		err := impl.server.Serve(lis)
		if err != nil {
			log.Panic("RPC服务异常关闭", zap.Error(err))
		}
		impl.healthServer.SetServingStatus(configuration.GetServiceInfo().ServiceName, grpc_health_v1.HealthCheckResponse_NOT_SERVING)
	}()
	time.Sleep(time.Millisecond * 10)
	impl.healthServer.SetServingStatus(configuration.GetServiceInfo().ServiceName, grpc_health_v1.HealthCheckResponse_SERVING)
	if !impl.config.DisableRegister {
		impl.grpcHealthCheckRegister()
	}
	log.Debug("启动RPC服务", zap.String("rpcServerAddr", impl.rpcServerAddr), zap.String("realAddr", lis.Addr().String()))
	return nil
}
func (impl *serviceImpl) Stop(ctx context.Context) errors.Error {
	impl.healthServer.SetServingStatus(configuration.GetServiceInfo().ServiceName, grpc_health_v1.HealthCheckResponse_NOT_SERVING)
	impl.server.Stop()
	log.Info("RPC服务关闭")
	return nil
}

func (impl *serviceImpl) grpcHealthCheckRegister() {
	agent := consul.GetService().GetConsulClient().Agent()
	serviceId := impl.GetRegisterServiceId()
	err := agent.CheckRegister(&api.AgentCheckRegistration{
		ID:        fmt.Sprintf("%s_grpcHealth", serviceId),
		Name:      fmt.Sprintf("%s_grpcHealth", serviceId),
		ServiceID: serviceId,
		AgentServiceCheck: api.AgentServiceCheck{
			Name:     fmt.Sprintf("%s_grpcHealth", serviceId),
			GRPC:     impl.rpcServerAddr,
			Interval: "10s",
			Timeout:  "2s",
		},
	})
	if err != nil {
		log.Error("注册GRPC失败", zap.Error(err))
		return
	}
	log.Debug("注册GRPCHealthCheck", zap.String("registerServiceId", serviceId))
}

func (impl *serviceImpl) GetRegisterServiceId() string {
	addr, err := net.ResolveTCPAddr("tcp", impl.rpcServerAddr)
	if err != nil {
		log.Panic("RPC服务地址解析失败")
	}
	serviceInfo := configuration.GetServiceInfo()
	return fmt.Sprintf("%s_%s", serviceInfo.ServiceName, addr.IP.String())
}
