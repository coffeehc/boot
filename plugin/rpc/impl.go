package rpc

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/coffeehc/boot/component/grpc/grpcquic"
	"github.com/lucas-clemente/quic-go"
	"golang.org/x/net/http2"
	"net"

	"github.com/coffeehc/base/log"
	"github.com/coffeehc/boot/configuration"
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

func (impl *serviceImpl) Start(ctx context.Context) error {
	log.Debug("开始启动RPC服务", zap.String("rpcServerAddr", impl.rpcServerAddr))
	cert, err := grpcquic.GenerateTlsSelfSignedCert()
	if err != nil {
		return err
	}
	tlsConfig := &tls.Config{
		Certificates:       []tls.Certificate{cert},
		NextProtos:         []string{"http/1.1", http2.NextProtoTLS, "coffee"},
		InsecureSkipVerify: true,
	}
	addr, _ := net.ResolveTCPAddr("tcp4", impl.rpcServerAddr)
	udpAddr, _ := net.ResolveUDPAddr("udp4", impl.rpcServerAddr)
	addr.IP = net.IPv4zero
	udplis, _err := quic.ListenAddr(udpAddr.String(), tlsConfig, nil)
	if _err != nil {
		log.Panic("启动RPC服务端口失败", zap.Error(_err))
	}
	udpListener := grpcquic.Listen(udplis)
	//tcpListener, _err := net.Listen("tcp4", addr.String())
	//if _err != nil {
	//	log.Panic("启动RPC服务端口失败", zap.Error(_err))
	//}
	grpc_health_v1.RegisterHealthServer(impl.server, impl.healthServer)
	go func() {
		impl.healthServer.SetServingStatus(configuration.GetServiceInfo().ServiceName, grpc_health_v1.HealthCheckResponse_SERVING)
		err := impl.server.Serve(udpListener)
		if err != nil {
			log.Panic("RPC服务异常关闭", zap.Error(err))
		}
		//err = impl.server.Serve(tcpListener)
		//if err != nil {
		//	log.Panic("RPC服务异常关闭", zap.Error(err))
		//}
		impl.healthServer.SetServingStatus(configuration.GetServiceInfo().ServiceName, grpc_health_v1.HealthCheckResponse_NOT_SERVING)
		log.Debug("启动RPCServer完成", zap.String("rpcServerAddr", impl.rpcServerAddr))
	}()
	return nil
}
func (impl *serviceImpl) Stop(ctx context.Context) error {
	impl.healthServer.SetServingStatus(configuration.GetServiceInfo().ServiceName, grpc_health_v1.HealthCheckResponse_NOT_SERVING)
	impl.server.Stop()
	log.Info("RPC服务关闭")
	return nil
}

func (impl *serviceImpl) GetRegisterServiceId() string {
	addr, err := net.ResolveTCPAddr("tcp", impl.rpcServerAddr)
	if err != nil {
		log.Panic("RPC服务地址解析失败")
	}
	serviceInfo := configuration.GetServiceInfo()
	return fmt.Sprintf("%s_%s", serviceInfo.ServiceName, addr.IP.String())
}
