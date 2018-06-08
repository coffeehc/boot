package etcdsd

import (
	"context"
	"fmt"
	"strings"

	"git.xiagaogao.com/coffee/boot"
	"git.xiagaogao.com/coffee/boot/errors"
	"git.xiagaogao.com/coffee/boot/logs"
	"git.xiagaogao.com/coffee/boot/sd"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/pquerna/ffjson/ffjson"
	"go.uber.org/zap"
	"google.golang.org/grpc/resolver"
)

var logger *zap.Logger

const MicorScheme = "micor"

func RegisterResolver(ctx context.Context, client *clientv3.Client, serviceInfo boot.ServiceInfo, defaultSrvAddr ...string) errors.Error {
	logger = logs.GetLogger(ctx)
	rb, err := newResolver(ctx, client, serviceInfo, defaultSrvAddr...)
	if err != nil {
		return err
	}
	resolver.Register(rb)
	return nil
}

func newResolver(ctx context.Context, client *clientv3.Client, serviceInfo boot.ServiceInfo, defaultSrvAddr ...string) (resolver.Builder, errors.Error) {
	rb := &etcdResolverBuilder{
		client:         client,
		ctx:            ctx,
		defaultSrvAddr: defaultSrvAddr,
		scheme:         sd.BuildServiceKeyPrefix(serviceInfo),
	}
	return rb, nil
}

type etcdResolverBuilder struct {
	client         *clientv3.Client
	ctx            context.Context
	defaultSrvAddr []string
	scheme         string
}

func (impl *etcdResolverBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOption) (resolver.Resolver, error) {
	ctx, cancel := context.WithCancel(impl.ctx)
	r := &etcdResolver{
		cc:             cc,
		client:         impl.client,
		ctx:            ctx,
		cancel:         cancel,
		defaultSrvAddr: impl.defaultSrvAddr,
		logger:         logs.GetLogger(ctx),
		keyPrefix:      fmt.Sprintf("/ms/registers/%s/%s/", target.Endpoint, target.Authority),
		target:         target,
	}
	addrList := r.initServerAddr()
	go r.watch(addrList)
	return r, nil
}

func (impl *etcdResolverBuilder) Scheme() string {
	return MicorScheme
}

type etcdResolver struct {
	cc             resolver.ClientConn
	client         *clientv3.Client
	ctx            context.Context
	cancel         context.CancelFunc
	defaultSrvAddr []string
	logger         *zap.Logger
	keyPrefix      string
	target         resolver.Target
}

func (impl *etcdResolver) ResolveNow(ro resolver.ResolveNowOption) {
}

// Close closes the resolver.
func (impl *etcdResolver) Close() {
	impl.cancel()
}

func (r *etcdResolver) initServerAddr() []resolver.Address {
	var addrList []resolver.Address
	for _, addr := range r.defaultSrvAddr {
		addrList = append(addrList, resolver.Address{Addr: addr})
	}
	getResp, err := r.client.Get(context.Background(), r.keyPrefix, clientv3.WithPrefix())
	if err != nil {
		logger.Error("etcd获取服务节点信息失败:%s", zap.Any(logs.K_Cause, err))
	} else {
		if getResp.Count == 0 {
			logger.Warn(fmt.Sprintf("服务[%s]没有足够的节点使用", r.target.Endpoint))
		}
		for _, kv := range getResp.Kvs {
			addrList = append(addrList, *r.getServiceAddr(kv))
		}
	}
	r.cc.NewAddress(addrList)
	return addrList
}

func (r *etcdResolver) watch(addrList []resolver.Address) {
	rch := r.client.Watch(context.Background(), r.keyPrefix, clientv3.WithPrefix())
	for n := range rch {
		for _, ev := range n.Events {
			addr := strings.TrimPrefix(string(ev.Kv.Key), r.keyPrefix)
			switch ev.Type {
			case mvccpb.PUT:
				if !exist(addrList, addr) {
					addrList = append(addrList, resolver.Address{Addr: addr})
					r.cc.NewAddress(addrList)
				}
			case mvccpb.DELETE:
				if s, ok := remove(addrList, addr); ok {
					addrList = s
					r.cc.NewAddress(addrList)
				}
			}
		}
	}
}

func exist(l []resolver.Address, addr string) bool {
	for i := range l {
		if l[i].Addr == addr {
			return true
		}
	}
	return false
}

func remove(s []resolver.Address, addr string) ([]resolver.Address, bool) {
	for i := range s {
		if s[i].Addr == addr {
			s[i] = s[len(s)-1]
			return s[:len(s)-1], true
		}
	}
	return nil, false
}

func (r *etcdResolver) getServiceAddr(kv *mvccpb.KeyValue) *resolver.Address {
	info := &sd.ServiceRegisterInfo{}
	err := ffjson.Unmarshal(kv.Value, info)
	if err != nil {
		logger.Error("Unmarshal err", zap.Any(logs.K_Cause, err))
		return nil
	}
	if info.ServiceInfo.GetServiceTag() == r.target.Authority {
		if info.ServerAddr == "" {
			return &resolver.Address{Addr: string(kv.Key[len(r.keyPrefix):])}
		}
		return &resolver.Address{Addr: info.ServerAddr}
	}
	return nil
}
