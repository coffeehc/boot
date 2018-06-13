package ipsd

import (
	"context"

	"git.xiagaogao.com/coffee/boot/errors"
	"go.uber.org/zap"
	"google.golang.org/grpc/resolver"
)

const IpScheme = "ip"

func RegisterResolver(ctx context.Context, errorService errors.Service, logger *zap.Logger, defaultSrvAddr ...string) errors.Error {
	rb, err := newResolver(ctx, errorService, logger, defaultSrvAddr...)
	if err != nil {
		return err
	}
	resolver.Register(rb)
	return nil
}

func newResolver(ctx context.Context, errorService errors.Service, logger *zap.Logger, defaultSrvAddr ...string) (resolver.Builder, errors.Error) {
	rb := &ipResolverBuilder{
		ctx:            ctx,
		defaultSrvAddr: defaultSrvAddr,
		errorService:   errorService,
		logger:         logger,
	}
	return rb, nil
}

type ipResolverBuilder struct {
	ctx            context.Context
	defaultSrvAddr []string
	errorService   errors.Service
	scheme         string
	logger         *zap.Logger
}

func (impl *ipResolverBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOption) (resolver.Resolver, error) {
	ctx, cancel := context.WithCancel(impl.ctx)
	r := &etcdResolver{
		cc:             cc,
		ctx:            ctx,
		cancel:         cancel,
		defaultSrvAddr: impl.defaultSrvAddr,
		logger:         impl.logger,
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
	addrList := []resolver.Address{}
	for _, addr := range r.defaultSrvAddr {
		addrList = append(addrList, resolver.Address{Addr: addr})
	}
	r.cc.NewAddress(addrList)
	return addrList
}
