package rpc

import (
	"context"
	"crypto/tls"

	"sync"

	"github.com/coffeehc/base/log"
	"github.com/coffeehc/boot/component/grpc/grpcserver"
	"github.com/coffeehc/boot/plugin"
	"github.com/lucas-clemente/quic-go"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
)

var service Service
var mutex = new(sync.RWMutex)
var name = "rpc"
var scope = zap.String("scope", name)

func GetService() Service {
	if service == nil {
		log.Panic("Service没有初始化", scope)
	}
	return service
}

type Service interface {
	GetGRPCServer() *grpc.Server
	GetRPCServerAddr() string
	GetRegisterServiceId() string
}

func EnablePlugin(ctx context.Context) {
	if name == "" {
		log.Panic("插件名称没有初始化")
	}
	mutex.Lock()
	defer mutex.Unlock()
	if service != nil {
		return
	}
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
	log.Debug("grpc.Config", zap.Any("config", config))
	_server, err := grpcserver.NewServer(ctx, &grpcserver.GRPCServerConfig{
		MaxMsgSize:           config.MaxMsgSize,
		MaxConcurrentStreams: config.MaxConcurrentStreams,
	})
	if err != nil {
		log.Panic("创建GRPC服务端失败")
	}
	//config.RPCServerAddr, err = WarpServiceAddr(config.RPCServerAddr)
	//if err != nil {
	//	log.Panic("GRPC服务地址处理失败", zap.Error(err))
	//}
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
	}
	service = impl
	// reflection.Register(service.GetGRPCServer()) //是否开启远程控制
	log.Debug("初始化RPC服务", zap.String("rpcServerAddr", impl.GetRPCServerAddr()))
	plugin.RegisterPlugin(name, impl)
}
