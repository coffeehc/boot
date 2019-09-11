package etcdsd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"git.xiagaogao.com/coffee/boot/base/errors"
	"git.xiagaogao.com/coffee/boot/base/log"
	"git.xiagaogao.com/coffee/boot/configuration"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/pquerna/ffjson/ffjson"
	"go.uber.org/zap"
	"google.golang.org/grpc/resolver"
)

func Resolver(ctx context.Context) errors.Error {
	rb, err := newResolver(ctx)
	if err != nil {
		return err
	}
	resolver.Register(rb)
	return nil
}

func newResolver(ctx context.Context, defaultSrvAddr ...string) (resolver.Builder, errors.Error) {
	rb := &etcdResolverBuilder{
		ctx:            ctx,
		defaultSrvAddr: defaultSrvAddr,
		scheme:         BuildServiceKeyPrefix(),
	}
	return rb, nil
}

type etcdResolverBuilder struct {
	ctx            context.Context
	defaultSrvAddr []string
	scheme         string
}

func (impl *etcdResolverBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOption) (resolver.Resolver, error) {
	log.Debug("Build Resolver", zap.Any("target", target))
	ctx, cancel := context.WithCancel(impl.ctx)
	r := &etcdResolver{
		cc:             cc,
		ctx:            ctx,
		cancel:         cancel,
		defaultSrvAddr: impl.defaultSrvAddr,
		keyPrefix:      fmt.Sprintf("/ms/registers/%s/%s/", target.Endpoint, target.Authority),
		target:         target,
		ServerName:     target.Endpoint,
		client:         GetEtcdClient(),
	}
	addrList := r.initServerAddr()
	r.cc.NewAddress(addrList)
	go r.watch(addrList)
	return r, nil
}

func (impl *etcdResolverBuilder) Scheme() string {
	return configuration.MicroServiceProtocolScheme
}

type etcdResolver struct {
	cc             resolver.ClientConn
	client         *clientv3.Client
	ctx            context.Context
	cancel         context.CancelFunc
	defaultSrvAddr []string
	keyPrefix      string
	target         resolver.Target
	ServerName     string
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
		addrList = append(addrList, resolver.Address{Addr: addr, ServerName: r.ServerName})
	}
	// log.Debug("Get service endpoints", zap.String("prefix", r.keyPrefix))
	getResp, err := r.client.Get(context.Background(), r.keyPrefix, clientv3.WithPrefix())
	if err != nil {
		log.Error("etcd获取服务节点信息失败:%s", zap.Any("cause", err))
	} else {
		// log.Debug("获取到节点数据", zap.Strings("endpoints",r.client.Endpoints()))//zap.String("prefix", r.keyPrefix),zap.Any("nodes",getResp))
		if getResp.Count == 0 {
			log.Warn(fmt.Sprintf("服务[%s]没有足够的节点使用", r.target.Endpoint))
		}
		for _, kv := range getResp.Kvs {
			addrList = append(addrList, *r.getServiceAddr(kv))
		}
	}
	return addrList
}

func (r *etcdResolver) watch(addrList []resolver.Address) {
	rch := r.client.Watch(context.Background(), r.keyPrefix, clientv3.WithPrefix())
	tiemOut := time.Second * 30
	timer := time.NewTimer(tiemOut)
	panicSleep := time.Second
	for {
		func() {
			defer func() {
				if err := recover(); err != nil {
					log.Error("监听地址节点出现异常", zap.Any("err", err))
					time.Sleep(panicSleep)
					panicSleep += time.Second
					if panicSleep > time.Second*10 {
						panicSleep = time.Second * 10
					}
				} else {
					panicSleep = time.Second
				}
			}()
			select {
			case <-timer.C:
				addrList = r.initServerAddr()
			case n, ok := <-rch:
				if !ok {
					rch = r.client.Watch(context.Background(), r.keyPrefix, clientv3.WithPrefix())
					return
				}
				for _, ev := range n.Events {
					addr := strings.TrimPrefix(string(ev.Kv.Key), r.keyPrefix)
					switch ev.Type {
					case mvccpb.PUT:
						log.Info("获取新的节点", zap.String("addr", addr))
						if !exist(addrList, addr) {
							addrList = append(addrList, resolver.Address{Addr: addr, ServerName: r.ServerName})
						}
					case mvccpb.DELETE:
						log.Warn("节点丢失", zap.String("nodeAddr", addr), zap.String("rpcService", r.ServerName))
						if s, ok := remove(addrList, addr); ok {
							addrList = s
						}
					}
				}
			}
			timer.Reset(tiemOut)
			r.cc.NewAddress(addrList)
		}()
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
	info := &configuration.ServiceRegisterInfo{}
	err := ffjson.Unmarshal(kv.Value, info)
	if err != nil {
		log.Error("注册信息反序列化失败", zap.Error(err), zap.String("body", string(kv.Value)))
		return nil
	}
	if configuration.GetModel() == r.target.Authority {
		if info.ServiceAddr == "" {
			return &resolver.Address{Addr: string(kv.Key[len(r.keyPrefix):]), ServerName: r.ServerName}
		}
		return &resolver.Address{Addr: info.ServiceAddr, ServerName: r.ServerName}
	}
	return nil
}
