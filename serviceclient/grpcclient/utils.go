package grpcclient

import (
	"github.com/coffeehc/logger"
	"github.com/coffeehc/microserviceboot/loadbalancer"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type balancerWapper struct {
	balancer     loadbalancer.Balancer
	addressCache chan []grpc.Address
}

func (this *balancerWapper) Start(target string, config grpc.BalancerConfig) error {
	err := this.balancer.Start(target, loadbalancer.BalancerConfig{
		DialCreds: config.DialCreds,
	})
	if err != nil {
		return err
	}
	go func() {
		for {
			select {
			case addrs := <-this.balancer.Notify():
				if addrs == nil {
					logger.Debug("return")
					return
				}
				if len(addrs) == 0 {
					break
				}
				rpcAddrs := make([]grpc.Address, len(addrs))
				for i, addr := range addrs {
					rpcAddrs[i] = grpc.Address{
						Addr:     addr.Addr,
						Metadata: addr.Metadata,
					}
				}
				this.addressCache <- rpcAddrs
			}
		}
	}()
	return nil

}
func (this *balancerWapper) Up(addr grpc.Address) (down func(error)) {
	return this.balancer.Up(loadbalancer.Address{
		Addr:     addr.Addr,
		Metadata: addr.Metadata,
	})
}
func (this *balancerWapper) Get(ctx context.Context, opts grpc.BalancerGetOptions) (grpc.Address, func(), error) {
	addr, put, err := this.balancer.Get(ctx, loadbalancer.BalancerGetOptions{
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
func (this *balancerWapper) Notify() <-chan []grpc.Address {
	return this.addressCache
}
func (this *balancerWapper) Close() error {
	return this.balancer.Close()
}

func BalancerWapper(balancer loadbalancer.Balancer) grpc.Balancer {
	return &balancerWapper{
		balancer:     balancer,
		addressCache: make(chan []grpc.Address, 1),
	}
}
