package loadbalancer

import (
	"github.com/coffeehc/microserviceboot/base"
	"google.golang.org/grpc/naming"
)

func newAddrArrayBalancer(addrs []string) (Balancer, base.Error) {
	r, err := newAddrArrayResolver(addrs)
	if err != nil {
		return nil, err
	}
	return RoundRobin(r), nil
}

type addrArrayResolver struct {
	updatesc chan []*naming.Update
}

func newAddrArrayResolver(addrs []string) (*addrArrayResolver, base.Error) {
	if addrs == nil || len(addrs) == 0 {
		return nil, base.NewError(-1, errScopeBalance, "addrs is nil")
	}
	resolver := &addrArrayResolver{
		updatesc: make(chan []*naming.Update, 1),
	}
	go func() {
		updates := make([]*naming.Update, len(addrs))
		for i, addr := range addrs {
			updates[i] = &naming.Update{
				Op:   naming.Add,
				Addr: addr,
			}
		}
		resolver.updatesc <- updates
	}()
	return resolver, nil
}

func (sr *addrArrayResolver) Resolve(target string) (naming.Watcher, error) {
	return sr, nil
}

func (sr *addrArrayResolver) Next() ([]*naming.Update, error) {
	return <-sr.updatesc, nil
}

func (sr *addrArrayResolver) Close() {
	close(sr.updatesc)
}
