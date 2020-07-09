package etcd_dc

import (
	"context"

	"git.xiagaogao.com/coffee/base/errors"
	"google.golang.org/grpc/resolver"
)

type serviceImpl struct {
}

func newService() *serviceImpl {
	impl := &serviceImpl{}
	return impl
}

func (impl *serviceImpl) GetResolverBuilder(ctx context.Context) (resolver.Builder, errors.Error) {
	return &resolverBuilder{
		ctx: ctx,
	}, nil
}
