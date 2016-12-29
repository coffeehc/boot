package grpcclient

import (
	"github.com/coffeehc/logger"
	"github.com/coffeehc/microserviceboot/loadbalancer"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type balancerAdopter struct {
	balancer     loadbalancer.Balancer
	addressCache chan []grpc.Address
}

func (ba *balancerAdopter) Start(target string, config grpc.BalancerConfig) error {
	err := ba.balancer.Start(target, loadbalancer.BalancerConfig{
		DialCreds: config.DialCreds,
	})
	if err != nil {
		return err
	}
	go func() {
		for addrs := range ba.balancer.Notify() {
			rpcAddrs := make([]grpc.Address, len(addrs))
			for i, addr := range addrs {
				rpcAddrs[i] = grpc.Address{
					Addr:     addr.Addr,
					Metadata: addr.Metadata,
				}
				logger.Debug("Notify addr %s", addr.Addr)
			}
			ba.addressCache <- rpcAddrs
		}
	}()
	return nil

}
func (ba *balancerAdopter) Up(addr grpc.Address) (down func(error)) {
	return ba.balancer.Up(loadbalancer.Address{
		Addr:     addr.Addr,
		Metadata: addr.Metadata,
	})
}
func (ba *balancerAdopter) Get(ctx context.Context, opts grpc.BalancerGetOptions) (grpc.Address, func(), error) {
	addr, put, err := ba.balancer.Get(ctx, loadbalancer.BalancerGetOptions{
		BlockingWait: opts.BlockingWait,
	})
	if err != nil {
		return grpc.Address{}, nil, err
	}
	return grpc.Address{
		Addr:     addr.Addr,
		Metadata: addr.Metadata,
	}, put, nil
}
func (ba *balancerAdopter) Notify() <-chan []grpc.Address {
	return ba.addressCache
}
func (ba *balancerAdopter) Close() error {
	close(ba.addressCache)
	return ba.balancer.Close()
}

func adopterToGRPCBalancer(balancer loadbalancer.Balancer) grpc.Balancer {
	return &balancerAdopter{
		balancer:     balancer,
		addressCache: make(chan []grpc.Address, 1),
	}
}

type reconnectionError struct {
	err error
}

func (e *reconnectionError) Error() string   { return "reconnectionError:" + e.err.Error() }
func (e *reconnectionError) Timeout() bool   { return true }
func (e *reconnectionError) Temporary() bool { return true }
