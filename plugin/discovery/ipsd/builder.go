package ipsd

import (
	"context"
	"google.golang.org/grpc/resolver"
)

const ServiceProtocolScheme = "ip"

type IpResolverBuilder struct {
	ctx            context.Context
	defaultSrvAddr []string
	scheme         string
	resolver       *ipResolver
}

func (impl *IpResolverBuilder) UpdateAddress(addresses []string) {
	impl.defaultSrvAddr = addresses
	if impl.resolver != nil {
		impl.resolver.defaultSrvAddr = impl.defaultSrvAddr
		impl.resolver.initServerAddr()
	}

}

func (impl *IpResolverBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	if impl.resolver != nil {
		return impl.resolver, nil
	}
	ctx, cancel := context.WithCancel(impl.ctx)
	r := &ipResolver{
		cc:             cc,
		ctx:            ctx,
		cancel:         cancel,
		defaultSrvAddr: impl.defaultSrvAddr,
		target:         target,
	}
	impl.resolver = r
	r.initServerAddr()
	return r, nil
}

func (impl *IpResolverBuilder) Scheme() string {
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

func (impl *ipResolver) ResolveNow(_ resolver.ResolveNowOptions) {
	//log.Debug("ResolveNow-----------------")
	impl.initServerAddr()
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
