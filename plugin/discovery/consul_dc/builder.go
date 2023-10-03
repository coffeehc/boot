package consul_dc

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/coffeehc/base/errors"
	"github.com/coffeehc/base/log"
	"github.com/coffeehc/boot/configuration"
	"github.com/hashicorp/consul/api"
	"go.uber.org/zap"
	"google.golang.org/grpc/resolver"
)

const ServiceProtocolScheme = "console"

type resolverBuilder struct {
	ctx context.Context
}

func (impl *resolverBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	ctx, cancel := context.WithCancel(impl.ctx)
	resolver := &consulResolver{
		target: target,
		cc:     cc,
		ctx:    ctx,
		cancel: cancel,
		// client:     consul.GetService().GetConsulClient(),
		ServerName: target.Endpoint(),
	}
	go resolver.watch()
	return resolver, nil
}

func (impl *resolverBuilder) Scheme() string {
	return ServiceProtocolScheme
}

type consulResolver struct {
	mutex      sync.Mutex
	cc         resolver.ClientConn
	client     *api.Client
	ctx        context.Context
	cancel     context.CancelFunc
	target     resolver.Target
	ServerName string
	lastIndex  uint64
}

func (impl *consulResolver) ResolveNow(resolver.ResolveNowOptions) {
}

// Close closes the resolver.
func (impl *consulResolver) Close() {
	impl.cancel()
}

func (impl *consulResolver) watch() {
	for impl.ctx.Err() == nil {
		impl.resolver()
	}
}

func (impl *consulResolver) resolver() error {
	impl.mutex.Lock()
	defer impl.mutex.Unlock()
	services, mateinfo, err := impl.client.Health().ServiceMultipleTags(impl.ServerName, []string{configuration.GetRunModel()}, true, &api.QueryOptions{
		WaitIndex: impl.lastIndex,
	})
	if err != nil {
		log.Error("接收service地址失败", zap.Error(err), zap.String("rpcServiceName", impl.ServerName))
		time.Sleep(time.Second * 3)
		return errors.MessageError(err.Error())
	}
	impl.lastIndex = mateinfo.LastIndex
	var newAddrs []resolver.Address
	for _, service := range services {
		addr := fmt.Sprintf("%v:%v", service.Service.Address, service.Service.Port)
		// log.Debug("adding service addrs", zap.String("ServerName", impl.ServerName), zap.String("serviceAddr", addr))
		newAddrs = append(newAddrs, resolver.Address{Addr: addr})
	}
	// log.Debug("获取了所有的Service地址", zap.String("ServerName", impl.ServerName), zap.Int("count", len(newAddrs)))
	impl.cc.UpdateState(resolver.State{Addresses: newAddrs})
	return nil
}
