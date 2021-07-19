package kubernetes

import (
	"context"
	"sync"

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
	resolver.initServerAddr()
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
}

func (impl *kubernetesResolver) ResolveNow(ro resolver.ResolveNowOptions) {
}

// Close closes the resolver.
func (impl *kubernetesResolver) Close() {
	impl.cancel()
}

func (impl *kubernetesResolver) initServerAddr() []resolver.Address {
	// addr,err = net.ResolveTCPAddr("tcp",impl.Endpoint)
	// if err!=nil{
	//   return nil
	// }
	addrList := []resolver.Address{
		{Addr: impl.Endpoint},
	}
	log.Debug("实际客户端地址", zap.Any("addList", addrList))
	return addrList
}
