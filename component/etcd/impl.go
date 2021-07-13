package etcd

import (
	"context"
	"time"

	"git.xiagaogao.com/coffee/base/errors"
	"git.xiagaogao.com/coffee/base/log"
	"git.xiagaogao.com/coffee/boot/plugin"
	"go.etcd.io/etcd/api/v3/etcdserverpb"
	"go.etcd.io/etcd/client/pkg/v3/transport"
	"go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type Service interface {
	plugin.Plugin
	GetClient() *clientv3.Client
	GetVersion() ([]*etcdserverpb.Member, errors.Error)
}

func newService(ctx context.Context) Service {
	var etcdCert = "./cert/etcd-client.crt"
	var etcdCertKey = "./cert/etcd-client.key"
	var etcdCa = "./cert/etcd-client-ca.crt"
	tlsinfo := &transport.TLSInfo{
		CertFile:      etcdCert,
		KeyFile:       etcdCertKey,
		TrustedCAFile: etcdCa,
	}
	tlsConfig, err := tlsinfo.ClientConfig()
	if err != nil {
		log.Error("构建tlsConfig失败", zap.Error(err))
		return nil
	}
	config := clientv3.Config{
		Endpoints:   []string{"10.11.22.232:2379", "10.11.22.233:2379", "10.11.22.234:2379"},
		DialTimeout: 5 * time.Second,
		Logger:      log.GetLogger(),
		Context:     ctx,
		TLS:         tlsConfig,
		DialOptions: []grpc.DialOption{grpc.WithNoProxy()},
	}
	client, err := clientv3.New(config)
	if err != nil {
		log.Error("初始化Etcd客户端失败", zap.Error(err))
		return nil
	}

	impl := &serviceImpl{
		client: client,
	}
	return impl
}

type serviceImpl struct {
	client *clientv3.Client
}

func (impl *serviceImpl) Start(ctx context.Context) errors.Error {
	log.Debug("service Start")
	return nil
}

func (impl *serviceImpl) Stop(ctx context.Context) errors.Error {
	impl.client.Close()
	return nil
}

func (impl *serviceImpl) GetClient() *clientv3.Client {
	return impl.client
}

func (impl *serviceImpl) GetVersion() ([]*etcdserverpb.Member, errors.Error) {
	listResp, err := impl.client.MemberList(context.TODO())
	if err != nil {
		return nil, errors.ConverError(err)
	}
	return listResp.Members, nil
}
