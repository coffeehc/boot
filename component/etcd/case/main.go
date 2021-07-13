package main

import (
	"context"
	"time"

	"git.xiagaogao.com/coffee/base/log"
	"git.xiagaogao.com/coffee/boot/component/grpc/grpcrecovery"
	"git.xiagaogao.com/coffee/boot/testutils"
	"go.etcd.io/etcd/client/pkg/v3/transport"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
)

func main() {
	testutils.InitTestConfig()
	grpcrecovery.SetLogLevel(zapcore.ErrorLevel)
	grpclog.SetLoggerV2(grpcrecovery.NewZapLogger())
	tlsInfo := &transport.TLSInfo{
		CertFile:      "./cert/etcd-client.crt",
		KeyFile:       "./cert/etcd-client.key",
		TrustedCAFile: "./cert/etcd-client-ca.crt",
	}
	tlsConfig, err := tlsInfo.ClientConfig()
	if err != nil {
		log.Error("构建tlsConfig失败", zap.Error(err))
		return
	}
	log.Debug("config", zap.Any("tlsConfig", tlsConfig))
	ctx := context.TODO()
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"10.11.22.232:2379"},
		DialTimeout: 5 * time.Second,
		Logger:      log.GetLogger(),
		Context:     ctx,
		TLS:         tlsConfig,
		DialOptions: []grpc.DialOption{
			grpc.WithNoProxy(),
		},
	})
	if err != nil {
		log.Error("初始化Etcd客户端失败", zap.Error(err))
		return
	}
	log.Debug("构建成功")
	members, err := client.MemberList(ctx)
	if err != nil {
		log.Error("错误", zap.Error(err))
		return
	}
	for _, member := range members.Members {
		log.Debug("member", zap.Any("member", member))
	}
}
