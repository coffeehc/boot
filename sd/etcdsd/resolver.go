package etcdsd

import (
	"context"
	"fmt"
	"strings"
	"time"

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

const MicorScheme = "micor"

func RegisterResolver(ctx context.Context, client *clientv3.Client, serviceInfo *boot.ServiceInfo, errorService errors.Service, logger *zap.Logger, defaultSrvAddr ...string) errors.Error {
	rb, err := newResolver(ctx, client, serviceInfo, errorService, logger, defaultSrvAddr...)
	if err != nil {
		return err
	}
	resolver.Register(rb)
	return nil
}

func newResolver(ctx context.Context, client *clientv3.Client, serviceInfo *boot.ServiceInfo, errorService errors.Service, logger *zap.Logger, defaultSrvAddr ...string) (resolver.Builder, errors.Error) {
	rb := &etcdResolverBuilder{
		client:         client,
		ctx:            ctx,
		defaultSrvAddr: defaultSrvAddr,
		scheme:         sd.BuildServiceKeyPrefix(serviceInfo),
		serviceInfo:    serviceInfo,
		errorService:   errorService,
		logger:         logger,
	}
	return rb, nil
}

type etcdResolverBuilder struct {
	client         *clientv3.Client
	ctx            context.Context
	defaultSrvAddr []string
	scheme         string
	serviceInfo    *boot.ServiceInfo
	errorService   errors.Service
	logger         *zap.Logger
}

func (impl *etcdResolverBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOption) (resolver.Resolver, error) {
	ctx, cancel := context.WithCancel(impl.ctx)
	r := &etcdResolver{
		cc:             cc,
		client:         impl.client,
		ctx:            ctx,
		cancel:         cancel,
		defaultSrvAddr: impl.defaultSrvAddr,
		logger:         impl.logger,
		keyPrefix:      fmt.Sprintf("/ms/registers/%s/%s/", target.Endpoint, target.Authority),
		target:         target,
		ServerName:     impl.serviceInfo.ServiceName,
	}
	addrList := r.initServerAddr()
	r.cc.NewAddress(addrList)
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
	ServerName     string
}

func (impl *etcdResolver) ResolveNow(ro resolver.ResolveNowOption) {
}

// Close closes the resolver.
func (impl *etcdResolver) Close() {
	impl.cancel()
}

func (r *etcdResolver) initServerAddr() []resolver.Address {
	addrList := []resolver.Address{}
	for _, addr := range r.defaultSrvAddr {
		addrList = append(addrList, resolver.Address{Addr: addr, ServerName: r.ServerName})
	}
	// r.logger.Debug("Get service endpoints", zap.String("prefix", r.keyPrefix))
	getResp, err := r.client.Get(context.Background(), r.keyPrefix, clientv3.WithPrefix())
	if err != nil {
		r.logger.Error("etcd获取服务节点信息失败:%s", zap.Any(logs.K_Cause, err))
	} else {
		if getResp.Count == 0 {
			r.logger.Warn(fmt.Sprintf("服务[%s]没有足够的节点使用", r.target.Endpoint))
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
					r.logger.Error("监听地址节点出现异常", zap.Any("err", err))
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
			case n, closed := <-rch:
				if closed {
					rch = r.client.Watch(context.Background(), r.keyPrefix, clientv3.WithPrefix())
					return
				}
				for _, ev := range n.Events {
					addr := strings.TrimPrefix(string(ev.Kv.Key), r.keyPrefix)
					switch ev.Type {
					case mvccpb.PUT:
						if !exist(addrList, addr) {
							addrList = append(addrList, resolver.Address{Addr: addr, ServerName: r.ServerName})
						}
					case mvccpb.DELETE:
						r.logger.Error("节点丢失", logs.F_ExtendData(addr), zap.String(logs.K_rpcService, r.ServerName))
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
	info := &sd.ServiceRegisterInfo{}
	err := ffjson.Unmarshal(kv.Value, info)
	if err != nil {
		r.logger.Error("Unmarshal err", zap.Any(logs.K_Cause, err), zap.String("kv.Value", string(kv.Value)))
		return nil
	}
	if boot.RunModel() == r.target.Authority {
		if info.ServerAddr == "" {
			return &resolver.Address{Addr: string(kv.Key[len(r.keyPrefix):]), ServerName: r.ServerName}
		}
		return &resolver.Address{Addr: info.ServerAddr, ServerName: r.ServerName}
	}
	return nil
}
