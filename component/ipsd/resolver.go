package ipsd

import (
	"context"

	"git.xiagaogao.com/coffee/boot/base/errors"
	"google.golang.org/grpc/resolver"
)

const IpScheme = "ip"

func RegisterResolver(ctx context.Context, defaultSrvAddr ...string) errors.Error {
	rb, err := newResolver(ctx, defaultSrvAddr...)
	if err != nil {
		return err
	}
	resolver.Register(rb)
	return nil
}

func newResolver(ctx context.Context, defaultSrvAddr ...string) (resolver.Builder, errors.Error) {
	rb := &ipResolverBuilder{
		ctx:            ctx,
		defaultSrvAddr: defaultSrvAddr,
	}
	return rb, nil
}

type ipResolverBuilder struct {
	ctx            context.Context
	defaultSrvAddr []string
	scheme         string
}

func (impl *ipResolverBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOption) (resolver.Resolver, error) {
	ctx, cancel := context.WithCancel(impl.ctx)
	r := &etcdResolver{
		cc:             cc,
		ctx:            ctx,
		cancel:         cancel,
		defaultSrvAddr: impl.defaultSrvAddr,
		target:         target,
	}
	r.initServerAddr()
	return r, nil
}

func (impl *ipResolverBuilder) Scheme() string {
	return IpScheme
}

type etcdResolver struct {
	cc             resolver.ClientConn
	ctx            context.Context
	cancel         context.CancelFunc
	defaultSrvAddr []string
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
	addrList := []resolver.Address{}
	for _, addr := range r.defaultSrvAddr {
		addrList = append(addrList, resolver.Address{Addr: addr})
	}
	r.cc.NewAddress(addrList)
	return addrList
}
