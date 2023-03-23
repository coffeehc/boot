package ipsd

import (
	"context"
	"google.golang.org/grpc/resolver"
)

const ServiceProtocolScheme = "ip"

type resolverBuilder struct {
	ctx            context.Context
	defaultSrvAddr []string
	scheme         string
	resolver       *ipResolver
}

func (impl *resolverBuilder) UpdateAddress(addresses []string) {
	impl.defaultSrvAddr = addresses
	impl.resolver.initServerAddr()
}

func (impl *resolverBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	ctx, cancel := context.WithCancel(impl.ctx)
	r := &ipResolver{
		cc:             cc,
		ctx:            ctx,
		cancel:         cancel,
		defaultSrvAddr: impl.defaultSrvAddr,
		target:         target,
	}
	r.initServerAddr()
	return r, nil
}

func (impl *resolverBuilder) Scheme() string {
	return ServiceProtocolScheme
}

type ipResolver struct {
	cc             resolver.ClientConn
	ctx            context.Context
	cancel         context.CancelFunc
	defaultSrvAddr []string
	keyPrefix      string
	target         resolver.Target
}

func (impl *ipResolver) ResolveNow(options resolver.ResolveNowOptions) {
	//log.Debug("ResolveNow-----------------")
}

// Close closes the resolver.
func (impl *ipResolver) Close() {
	impl.cancel()
}

func (r *ipResolver) initServerAddr() []resolver.Address {
	addrList := []resolver.Address{}
	for _, addr := range r.defaultSrvAddr {
		addrList = append(addrList, resolver.Address{Addr: addr})
	}
	r.cc.UpdateState(resolver.State{
		Addresses: addrList,
	})
	return addrList
}
