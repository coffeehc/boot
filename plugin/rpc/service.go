package rpc

import (
	"context"
	"git.xiagaogao.com/coffee/base/log"
	"git.xiagaogao.com/coffee/base/utils"
	"git.xiagaogao.com/coffee/boot/component/grpc/grpcserver"
	"git.xiagaogao.com/coffee/boot/crets"
	"git.xiagaogao.com/coffee/boot/plugin"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"net"
	"sync"
)

var service Service
var mutex = new(sync.RWMutex)
var name = "rpc"
var scope = zap.String("scope", name)

func GetService() Service {
	if service == nil {
		log.Fatal("Service没有初始化", scope)
	}
	return service
}

type Service interface {
	GetGRPCServer() *grpc.Server
	GetRPCServerAddr() string
}

func EnablePlugin(ctx context.Context) {
	if name == "" {
		log.Fatal("插件名称没有初始化")
	}
	mutex.Lock()
	defer mutex.Unlock()
	if service != nil {
		return
	}
	viper.SetDefault("grpc.MaxConcurrentStreams", 100000)
	viper.SetDefault("grpc.MaxMsgSize", 1024*1024*4)
	viper.SetDefault("grpc.RPCServerAddr", "0.0.0.0:0")
	viper.SetDefault("grpc.openTLS", false)
	config := &RpcConfig{}
	_err := viper.UnmarshalKey("grpc", config)
	if _err != nil {
		log.Fatal("加载GRPC配置失败", zap.Error(_err))
	}
	if viper.GetBool("grpc.openTLS") {
		log.Debug("开启TLS")
		grpcserver.SetCerds(ctx, crets.NewServerCreds())
	}
	_server, err := grpcserver.NewServer(ctx, &grpcserver.GRPCServerConfig{
		MaxMsgSize:           config.MaxMsgSize,
		MaxConcurrentStreams: config.MaxConcurrentStreams,
	})
	if err != nil {
		log.Fatal("创建GRPC服务端失败")
	}
	config.RPCServerAddr, err = utils.WarpServiceAddr(config.RPCServerAddr)
	if err != nil {
		log.Fatal("GRPC服务地址处理失败", err.GetFieldsWithCause()...)
	}
	lis, _err := net.Listen("tcp4", config.RPCServerAddr)
	if _err != nil {
		log.Fatal("启动RPC服务端口失败", zap.Error(_err))
	}
	rpcServerAddr := lis.Addr().String()
	lis.Close()
	impl := &serviceImpl{
		server:        _server,
		config:        config,
		rpcServerAddr: rpcServerAddr,
	}
	service = impl
	log.Debug("初始化RPC服务", zap.String("rpcServerAddr", impl.GetRPCServerAddr()))
	plugin.RegisterPlugin("rpcServer", impl)
}
