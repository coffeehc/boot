package discovery

import (
	"context"
	"time"

	"git.xiagaogao.com/coffee/boot/base/errors"
	"git.xiagaogao.com/coffee/boot/base/log"
	"git.xiagaogao.com/coffee/boot/component/grpc/codec/proxycodec"
	"git.xiagaogao.com/coffee/boot/component/grpc/grpcclient"
	"git.xiagaogao.com/coffee/boot/component/grpc/grpcrecovery"
	"git.xiagaogao.com/coffee/boot/configuration"
	"git.xiagaogao.com/coffee/boot/crets"
	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-prometheus"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
	"google.golang.org/grpc/keepalive"
)

func RPCServiceInitializationByAddress(ctx context.Context, rpcService configuration.RPCService, serverAddr string) errors.Error {
	chainUnaryClient := grpc_middleware.ChainUnaryClient(
		grpc_prometheus.UnaryClientInterceptor,
		grpcrecovery.UnaryClientInterceptor(),
	)
	chainStreamClient := grpc_middleware.ChainStreamClient(
		grpc_prometheus.StreamClientInterceptor,
		grpcrecovery.StreamClientInterceptor(),
	)
	opts := []grpc.DialOption{
		grpc.WithBackoffMaxDelay(time.Second * 10),
		grpc.WithAuthority(configuration.GetModel()),
		grpc.WithDefaultCallOptions(grpc.UseCompressor("gzip"), grpc.FailFast(true)),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{Time: time.Second * 5, Timeout: time.Second * 10, PermitWithoutStream: false}),
		grpc.WithBalancerName(roundrobin.Name),
		grpc.WithUserAgent("coffee's client"),
		grpc.WithUnaryInterceptor(chainUnaryClient),
		grpc.WithStreamInterceptor(chainStreamClient),
		grpc.WithInitialConnWindowSize(10),
		grpc.WithInitialWindowSize(1024),
		grpc.WithChannelzParentID(0),
		grpc.FailOnNonTempDialError(true),
		grpc.WithInsecure(),
	}
	ctx, _ = context.WithTimeout(ctx, time.Second*10)
	conn, err := grpc.DialContext(ctx, serverAddr, opts...)
	if err != nil {
		return errors.ConverError(err)
	}
	log.Debug("需要链接的服务端地址", zap.String("target", serverAddr))
	_err := rpcService.InitRPCService(ctx, conn)
	if _err != nil {
		return _err
	}
	log.Info("初始化RPCService成功", zap.Any("RPCService", rpcService.GetRPCServiceInfo()))
	return nil
}

func RPCServiceInitialization(ctx context.Context, rpcService configuration.RPCService) errors.Error {
	var conn *grpc.ClientConn = nil
	var err errors.Error = nil
	v := ctx.Value(contextKeyProxyGateway)
	if v == nil || v.(string) == "" {
		conn, err = grpcclient.NewClientConnByRegister(ctx, rpcService.GetRPCServiceInfo(), false)
	} else {
		proxyAddr := v.(string)
		chainUnaryClient := grpc_middleware.ChainUnaryClient(
			grpc_prometheus.UnaryClientInterceptor,
			grpcrecovery.UnaryClientInterceptor(),
		)
		chainStreamClient := grpc_middleware.ChainStreamClient(
			grpc_prometheus.StreamClientInterceptor,
			grpcrecovery.StreamClientInterceptor(),
		)
		opts := []grpc.DialOption{
			grpc.WithBackoffMaxDelay(time.Second * 10),
			grpc.WithAuthority(configuration.GetModel()),
			grpc.WithDefaultCallOptions(grpc.UseCompressor("gzip"), grpc.FailFast(true), grpc.CallContentSubtype(proxycodec.Name)),
			grpc.WithKeepaliveParams(keepalive.ClientParameters{Time: time.Second * 5, Timeout: time.Second * 10, PermitWithoutStream: false}),
			grpc.WithBalancerName(roundrobin.Name),
			grpc.WithUserAgent("coffee's client"),
			grpc.WithUnaryInterceptor(chainUnaryClient),
			grpc.WithStreamInterceptor(chainStreamClient),
			grpc.WithInitialConnWindowSize(10),
			grpc.WithInitialWindowSize(1024),
			grpc.WithChannelzParentID(0),
			grpc.FailOnNonTempDialError(true),
			grpc.WithTransportCredentials(crets.NewClientCreds("proxy.51apis.com")),
		}
		ctx, _ = context.WithTimeout(ctx, time.Second*5)
		_conn, _err := grpc.DialContext(ctx, proxyAddr, opts...)
		if _err != nil {
			return errors.ConverError(_err)
		}
		conn = _conn
		log.Debug("需要链接的服务端地址", zap.String("target", proxyAddr))

	}
	if err != nil {
		return err
	}
	err = rpcService.InitRPCService(ctx, conn)
	if err != nil {
		return err
	}
	log.Info("初始化RPCService成功", zap.Any("RPCService", rpcService.GetRPCServiceInfo()))
	return nil
}

const (
	contextKeyProxyGateway = "_grpc.ProxyGatewayAddr"
)

func SetProxyGatewayAddr(ctx context.Context, addr string) context.Context {
	return context.WithValue(ctx, contextKeyProxyGateway, addr)
}
