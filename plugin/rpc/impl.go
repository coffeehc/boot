package rpc

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"sync"

	"github.com/coffeehc/base/log"
	"github.com/coffeehc/boot/component/grpcx/grpcquic"
	"github.com/coffeehc/boot/component/grpcx/grpcserver"
	"github.com/coffeehc/boot/configuration"
	"github.com/quic-go/quic-go"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"golang.org/x/net/http2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

// SetRPCServerAddr 设置监听地址，须在 EnablePlugin 前调用；监听 IP 会原样保留。
func SetRPCServerAddr(addr string) {
	viper.Set("grpc.rpc_server_addr", addr)
}

// newService 只装配 server 和健康服务，端口绑定由 Start 负责。
func newService(ctx context.Context) Service {
	viper.SetDefault("grpc.max_concurrent_streams", 100)
	viper.SetDefault("grpc.max_msg_size", 8*1024*1024)
	viper.SetDefault("grpc.rpc_server_addr", "0.0.0.0:8888")
	viper.SetDefault("grpc.disable_quic_server", true)
	config := &RpcConfig{}
	if err := viper.UnmarshalKey("grpc", config); err != nil {
		log.Panic("加载 gRPC 配置失败", zap.Error(err))
	}
	if config.DisableTCPServer && config.DisableQUICServer {
		log.Panic("RPC 服务至少需要启用一种监听协议")
	}
	addr, err := net.ResolveTCPAddr("tcp", config.RPCServerAddr)
	if err != nil {
		log.Panic("解析 RPC 监听地址失败", zap.Error(err))
	}
	server, err := grpcserver.NewServer(ctx, nil)
	if err != nil {
		log.Panic("创建 gRPC 服务端失败", zap.Error(err))
	}
	healthServer := health.NewServer()
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_NOT_SERVING)
	healthServer.SetServingStatus(configuration.GetServiceInfo().ServiceName, grpc_health_v1.HealthCheckResponse_NOT_SERVING)
	grpc_health_v1.RegisterHealthServer(server, healthServer)
	return &serviceImpl{
		server:        server,
		config:        config,
		rpcServerAddr: addr.String(),
		healthServer:  healthServer,
	}
}

// serviceImpl 拥有 gRPC 监听、健康状态和关闭流程；生命周期由 plugin 串行驱动。
type serviceImpl struct {
	// config 是启动装配时加载的监听配置，之后只读。
	config *RpcConfig
	// server 由调用方在 Start 前注册业务服务。
	server *grpc.Server
	// addressMutex 保护实际绑定地址的发布和读取。
	addressMutex sync.RWMutex
	// rpcServerAddr 在 Start 前为配置地址，之后为实际绑定地址。
	rpcServerAddr string
	// healthServer 为服务注册和客户端健康检查提供状态。
	healthServer *health.Server
	// serveWorkers 只等待监听 goroutine，不等待忽略取消的业务 handler。
	serveWorkers sync.WaitGroup
}

// GetRPCServerAddr 返回监听地址；端口为 0 时，Start 成功后才能取得实际端口。
func (impl *serviceImpl) GetRPCServerAddr() string {
	impl.addressMutex.RLock()
	defer impl.addressMutex.RUnlock()
	return impl.rpcServerAddr
}

// GetGRPCServer 返回用于注册业务服务的 server，注册必须在 Start 前完成。
func (impl *serviceImpl) GetGRPCServer() *grpc.Server {
	return impl.server
}

// Start 同步绑定全部启用的端口，成功后交由 Serve goroutine 接收 RPC。
// 绑定失败由本实例释放已打开的监听器，再将错误交给 plugin 启动流程。
func (impl *serviceImpl) Start(ctx context.Context) (startErr error) {
	if err := ctx.Err(); err != nil {
		return err
	}
	listeners := make([]net.Listener, 0, 2)
	defer func() {
		if startErr != nil {
			for _, listener := range listeners {
				startErr = errors.Join(startErr, listener.Close())
			}
		}
	}()
	address := impl.GetRPCServerAddr()
	if !impl.config.DisableTCPServer {
		listener, err := net.Listen("tcp", address)
		if err != nil {
			return fmt.Errorf("监听 RPC TCP 地址 %s 失败: %w", address, err)
		}
		listeners = append(listeners, listener)
		address = listener.Addr().String()
	}
	if !impl.config.DisableQUICServer {
		cert, err := grpcquic.GenerateTlsSelfSignedCert()
		if err != nil {
			return fmt.Errorf("创建 RPC QUIC 证书失败: %w", err)
		}
		tlsConfig := &tls.Config{
			Certificates: []tls.Certificate{cert},
			NextProtos:   []string{http2.NextProtoTLS},
			MinVersion:   tls.VersionTLS13,
		}
		listener, err := quic.ListenAddr(address, tlsConfig, nil)
		if err != nil {
			return fmt.Errorf("监听 RPC QUIC 地址 %s 失败: %w", address, err)
		}
		listeners = append(listeners, grpcquic.Listen(*listener))
		address = listener.Addr().String()
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	impl.addressMutex.Lock()
	impl.rpcServerAddr = address
	impl.addressMutex.Unlock()
	impl.healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	impl.healthServer.SetServingStatus(configuration.GetServiceInfo().ServiceName, grpc_health_v1.HealthCheckResponse_SERVING)
	for _, listener := range listeners {
		impl.serveWorkers.Add(1)
		go impl.serve(listener)
	}
	log.Info("RPC 服务监听已启动", zap.String("address", address))
	return nil
}

// serve 记录监听异常并停止整个 RPC server，使健康状态和监听状态一致。
func (impl *serviceImpl) serve(listener net.Listener) {
	defer impl.serveWorkers.Done()
	if err := impl.server.Serve(listener); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
		log.Error("RPC 服务监听异常结束", zap.String("address", listener.Addr().String()), zap.Error(err))
		impl.healthServer.Shutdown()
		impl.server.Stop()
	}
}

// Stop 先停止健康报告并等待在途 RPC；shutdown context 到期后关闭连接并返回取消原因。
func (impl *serviceImpl) Stop(ctx context.Context) error {
	impl.healthServer.Shutdown()
	done := make(chan struct{})
	go impl.gracefulStop(done)
	select {
	case <-done:
		impl.serveWorkers.Wait()
		log.Info("RPC 服务已关闭")
		return nil
	case <-ctx.Done():
		impl.server.Stop()
		impl.serveWorkers.Wait()
		return fmt.Errorf("RPC 服务等待在途请求结束超时或取消: %w", ctx.Err())
	}
}

// gracefulStop 与 Stop 的超时分支并发；业务 handler 仍需响应连接取消。
func (impl *serviceImpl) gracefulStop(done chan<- struct{}) {
	impl.server.GracefulStop()
	close(done)
}

// GetRegisterServiceId 以服务名和监听 IP 构造注册标识。
func (impl *serviceImpl) GetRegisterServiceId() string {
	addr, err := net.ResolveTCPAddr("tcp", impl.GetRPCServerAddr())
	if err != nil {
		log.Panic("RPC 服务地址解析失败", zap.Error(err))
	}
	return fmt.Sprintf("%s_%s", configuration.GetServiceInfo().ServiceName, addr.IP.String())
}
