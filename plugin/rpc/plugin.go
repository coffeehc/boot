package rpc

import (
	"context"
	"net"
	"sync"
	"time"

	"git.xiagaogao.com/coffee/boot/base/errors"
	"git.xiagaogao.com/coffee/boot/base/log"
	"git.xiagaogao.com/coffee/boot/base/utils"
	"git.xiagaogao.com/coffee/boot/component/grpc/grpcserver"
	"git.xiagaogao.com/coffee/boot/crets"
	"git.xiagaogao.com/coffee/boot/plugin"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type RpcConfig struct {
	MaxMsgSize           int    // `yaml:"max_msg_size"`
	MaxConcurrentStreams uint32 // `yaml:"max_concurrent_streams"`
	RPCServerAddr        string
}

var rpcServerAddr string
var server *grpc.Server
var mutex = new(sync.Mutex)

func EnablePlugin(ctx context.Context) {
	mutex.Lock()
	defer mutex.Unlock()
	if server != nil {
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
	rpcServerAddr = lis.Addr().String()
	lis.Close()
	server = _server
	log.Debug("初始化RPC服务", zap.String("rpcServerAddr", rpcServerAddr))
	plugin.RegisterPlugin("rpcServer", &pluginImpl{
		config: config,
	})
}

type pluginImpl struct {
	config *RpcConfig
}

func (impl *pluginImpl) Start(ctx context.Context) errors.Error {
	addr, _ := net.ResolveTCPAddr("tcp4", rpcServerAddr)
	addr.IP = net.IPv4zero
	lis, _err := net.Listen("tcp4", addr.String())
	if _err != nil {
		log.Fatal("启动RPC服务端口失败", zap.Error(_err))
	}
	go func() {
		err := server.Serve(lis)
		if err != nil {
			log.Fatal("RPC服务异常关闭", zap.Error(err))
		}
	}()
	time.Sleep(time.Millisecond * 10)
	log.Debug("启动RPC服务", zap.String("rpcServerAddr", rpcServerAddr))
	return nil
}
func (impl *pluginImpl) Stop(ctx context.Context) errors.Error {
	server.Stop()
	log.Info("RPC服务关闭")
	return nil
}

func GetRPCServer() *grpc.Server {
	return server
}

func GetRPCServerAddr() string {
	if server == nil {
		log.Fatal("rpc server没有初始化")
	}
	return rpcServerAddr
}

func RegisterRPCMotheds(register func(server *grpc.Server)) {
	if server == nil {
		log.Error("GRPC服务没有初始化,不能注册")
	}
	log.Debug("开始注册RPC接口方法")
	register(server)
}
