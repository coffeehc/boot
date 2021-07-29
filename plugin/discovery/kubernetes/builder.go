package kubernetes

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"git.xiagaogao.com/coffee/base/errors"
	"git.xiagaogao.com/coffee/base/log"
	"go.uber.org/zap"
	"google.golang.org/grpc/resolver"
)

const ServiceProtocolScheme = "kubernetes"

type resolverBuilder struct {
	ctx            context.Context
	defaultSrvAddr []string
	scheme         string
}

func (impl *resolverBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	ctx, cancel := context.WithCancel(impl.ctx)
	resolver := &kubernetesResolver{
		target: target,
		cc:     cc,
		ctx:    ctx,
		cancel: cancel,
		// client:     consul.GetService().GetConsulClient(),
		Endpoint: target.Endpoint,
	}
	go resolver.watch()
	return resolver, nil
}

func (impl *resolverBuilder) Scheme() string {
	return ServiceProtocolScheme
}

type kubernetesResolver struct {
	mutex     sync.Mutex
	cc        resolver.ClientConn
	ctx       context.Context
	cancel    context.CancelFunc
	target    resolver.Target
	Endpoint  string
	lastIndex uint64
	addrs     map[string]struct{}
}

func (impl *kubernetesResolver) ResolveNow(ro resolver.ResolveNowOptions) {
}

// Close closes the resolver.
func (impl *kubernetesResolver) Close() {
	impl.cancel()
}

func (impl *kubernetesResolver) watch() {
	for impl.ctx.Err() == nil {
		impl.resolver()
		time.Sleep(time.Second * 30)
	}
}

func (impl *kubernetesResolver) resolver() errors.Error {
	impl.mutex.Lock()
	defer impl.mutex.Unlock()
	addrList := make([]resolver.Address, 0)
	host, port, err := net.SplitHostPort(impl.Endpoint)
	if err != nil {
		log.Error("地址解析错误", zap.Error(err))
		return errors.ConverError(err)
	}
	addrs, err := net.LookupHost(host)
	if err != nil {
		log.Error("地址DNS解析错误", zap.Error(err))
		return errors.ConverError(err)
	}
	for _, addr := range addrs {
		// log.Debug("获取地址", zap.String("host", host), zap.String("addr", addr))
		addrList = append(addrList, resolver.Address{
			Addr: fmt.Sprintf("%s:%s", addr, port),
		},
		)
	}
	impl.cc.UpdateState(resolver.State{Addresses: addrList})
	return nil
}
