package rpc

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/coffeehc/boot/component/grpc/grpcquic"
	"github.com/coffeehc/boot/component/grpc/grpcserver"
	"github.com/lucas-clemente/quic-go"
	"github.com/spf13/viper"
	"golang.org/x/net/http2"
	"net"

	"github.com/coffeehc/base/log"
	"github.com/coffeehc/boot/configuration"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

func newService(ctx context.Context) Service {
	viper.SetDefault("grpc.max_concurrent_streams", 100000)
	viper.SetDefault("grpc.max_msg_size", 1024*1024*4)
	viper.SetDefault("grpc.rpc_server_addr", "0.0.0.0:8888")
	config := &RpcConfig{}
	_err := viper.UnmarshalKey("grpc", config)
	if _err != nil {
		log.Panic("加载GRPC配置失败", zap.Error(_err))
	}
	// if viper.GetBool("grpc.openTLS") {
	// 	log.Debug("开启TLS")
	// 	grpcserver.SetCerds(ctx, crets.NewServerCreds())
	// }
	cert, err := grpcquic.GenerateTlsSelfSignedCert()
	if err != nil {
		return nil
	}
	log.Debug("grpc.Config", zap.Any("config", config))
	_server, err := grpcserver.NewServer(ctx, &grpcserver.GRPCServerConfig{
		MaxMsgSize:           config.MaxMsgSize,
		MaxConcurrentStreams: config.MaxConcurrentStreams,
	})
	if err != nil {
		log.Panic("创建GRPC服务端失败")
	}
	lis, _err := quic.ListenAddr(config.RPCServerAddr, &tls.Config{InsecureSkipVerify: true}, nil)
	//lis, _err := net.Listen("tcp4", config.RPCServerAddr)
	if _err != nil {
		log.Panic("启动RPC服务端口失败", zap.Error(_err))
	}
	rpcServerAddr := lis.Addr().String()
	lis.Close()
	impl := &serviceImpl{
		server:        _server,
		config:        config,
		rpcServerAddr: rpcServerAddr,
		healthServer:  health.NewServer(),
		cert:          cert,
	}
	return impl
}

type serviceImpl struct {
	config        *RpcConfig
	server        *grpc.Server
	rpcServerAddr string
	healthServer  *health.Server
	cert          tls.Certificate
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
	tlsConfig := &tls.Config{
		Certificates:       []tls.Certificate{impl.cert},
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
	//tcpListener, _err := tls.Listen("tcp4", addr.String(), tlsConfig)
	tcpListener, _err := net.Listen("tcp4", addr.String())
	if _err != nil {
		log.Panic("启动RPC服务端口失败", zap.Error(_err))
	}
	grpc_health_v1.RegisterHealthServer(impl.server, impl.healthServer)
	go func() {
		impl.healthServer.SetServingStatus(configuration.GetServiceInfo().ServiceName, grpc_health_v1.HealthCheckResponse_SERVING)
		err := impl.server.Serve(tcpListener)
		if err != nil {
			log.Panic("RPC服务异常关闭", zap.Error(err))
		}
		impl.healthServer.SetServingStatus(configuration.GetServiceInfo().ServiceName, grpc_health_v1.HealthCheckResponse_NOT_SERVING)
		log.Debug("启动RPCServer完成", zap.String("rpcServerAddr", impl.rpcServerAddr))
	}()
	go func() {
		impl.healthServer.SetServingStatus(configuration.GetServiceInfo().ServiceName, grpc_health_v1.HealthCheckResponse_SERVING)
		err := impl.server.Serve(udpListener)
		if err != nil {
			log.Panic("RPC服务异常关闭", zap.Error(err))
		}
		log.Debug("QUIC协议支持开启")
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
