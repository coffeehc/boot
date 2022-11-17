package grpcclient

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/coffeehc/boot/component/grpc/grpcquic"
	"golang.org/x/net/http2"
	"google.golang.org/grpc/credentials/insecure"
	"time"

	"github.com/coffeehc/base/errors"
	"github.com/coffeehc/base/log"
	"github.com/coffeehc/boot/component/grpc/grpcrecovery"
	"github.com/coffeehc/boot/configuration"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/balancer/roundrobin"
	"google.golang.org/grpc/credentials"
	_ "google.golang.org/grpc/encoding/gzip"
	"google.golang.org/grpc/keepalive"
)

func EnableQuic(ctx context.Context, enable bool) context.Context {
	return context.WithValue(ctx, "_EnableQuic", enable)
}

func getEnableQuic(ctx context.Context) bool {
	v := ctx.Value("_EnableQuic")
	if v == nil {
		return false
	}
	return v.(bool)
}

var scope = zap.String("scope", "grpc.client")

func NewClientConnByServiceInfo(ctx context.Context, serviceInfo configuration.ServiceInfo, block bool) (*grpc.ClientConn, error) {
	opts := BuildDialOption(ctx, block)
	target := serviceInfo.Target
	if target.Scheme == "" {
		log.Panic("没有指定需要链接的ServiceInfo的RPC协议，无法创建链接")
	}
	targetUrl := fmt.Sprintf("%s://%s/%s", target.Scheme, target.Authority, target.Endpoint)
	log.Debug("需要获取的客户端地址", zap.String("target", targetUrl))
	ctx, _ = context.WithTimeout(ctx, time.Second*5)
	clientConn, err := grpc.DialContext(ctx, targetUrl, opts...)
	if err != nil {
		log.Error("创建服务端链接失败", zap.Error(err))
		return nil, errors.SystemError("创建grpc客户端")
	}
	return clientConn, nil
}

func NewClientConn(ctx context.Context, block bool, serverAddr string) (*grpc.ClientConn, error) {
	opts := BuildDialOption(ctx, block)
	ctx, _ = context.WithTimeout(ctx, time.Second*5)
	clientConn, err := grpc.DialContext(ctx, serverAddr, opts...)
	// log.Debug("需要链接的服务端地址", zap.String("target", serverAddr))
	if err != nil {
		log.Error("创建客户端链接失败", zap.Error(err))
		return nil, errors.WrappedSystemError(err)
	}
	return clientConn, nil
}

func BuildDialOption(ctx context.Context, block bool) []grpc.DialOption {
	chainUnaryClient := []grpc.UnaryClientInterceptor{
		grpc_prometheus.UnaryClientInterceptor,
		grpcrecovery.UnaryClientInterceptor(),
	}
	chainStreamClient := []grpc.StreamClientInterceptor{
		grpc_prometheus.StreamClientInterceptor,
		grpcrecovery.StreamClientInterceptor(),
	}
	opts := []grpc.DialOption{
		grpc.WithConnectParams(grpc.ConnectParams{
			Backoff: backoff.Config{
				BaseDelay:  time.Second,      // 第一次失败重试前后需等待多久
				Multiplier: 1.5,              // 在失败的重试后乘以的倍数
				Jitter:     0.2,              // 随机抖动因子
				MaxDelay:   time.Second * 30, // backoff上限
			},
			MinConnectTimeout: time.Second,
		}),
		grpc.WithAuthority(configuration.GetRunModel()),
		grpc.WithDefaultCallOptions(
			grpc.UseCompressor("gzip"),
			grpc.WaitForReady(true),
			grpc.MaxCallRecvMsgSize(1024*1024*8),
			grpc.MaxCallSendMsgSize(1024*1024*2),
		),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                time.Second * 3,
			Timeout:             time.Second * 10,
			PermitWithoutStream: false,
		}),
		grpc.WithDefaultServiceConfig(fmt.Sprintf(`{"LoadBalancingPolicy": "%s"}`, roundrobin.Name)),
		grpc.WithUserAgent("coffee's client"),
		grpc.WithChainStreamInterceptor(chainStreamClient...),
		grpc.WithChainUnaryInterceptor(chainUnaryClient...),
		grpc.WithInitialConnWindowSize(1024 * 64),
		grpc.WithInitialWindowSize(1024 * 256),
		//grpc.WithChannelzParentID(&channelz.Identifier{}),
		grpc.FailOnNonTempDialError(false),
		grpc.WithNoProxy(),
		grpc.WithReadBufferSize(1024 * 128),
		grpc.WithWriteBufferSize(1024 * 128),
	}
	perRPCCredentials := ctx.Value(perRPCCredentialsKey)
	if perRPCCredentials != nil {
		if prc, ok := perRPCCredentials.(credentials.PerRPCCredentials); ok {
			opts = append(opts, grpc.WithPerRPCCredentials(prc))
		}
	}
	enableQUiC := getEnableQuic(ctx)
	if !enableQUiC {
		creds := getCerts(ctx)
		if creds == nil {
			opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
		} else {
			log.Error("暂不支持使用自定义证书")
		}
	} else {
		creds := grpcquic.NewCredentials(&tls.Config{
			NextProtos:         []string{"http/1.1", http2.NextProtoTLS, "coffee"},
			InsecureSkipVerify: true,
		})
		opts = append(opts, grpc.WithTransportCredentials(creds))
		tlsConf := &tls.Config{
			NextProtos:         []string{"http/1.1", http2.NextProtoTLS, "coffee"},
			InsecureSkipVerify: true,
		}
		opts = append(opts, grpc.WithContextDialer(grpcquic.NewQuicDialer(tlsConf)))
	}
	if block {
		opts = append(opts, grpc.WithBlock())
	}
	return opts
}
