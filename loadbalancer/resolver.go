package loadbalancer

import (
	"github.com/coffeehc/microserviceboot/base"
	"google.golang.org/grpc/naming"
)

func NewSimpleBalancer(addrs []string) (Balancer, base.Error) {
	r, err := NewSimpleResolver(addrs)
	if err != nil {
		return nil, err
	}
	return RoundRobin(r), nil
}

type SimpleResolver struct {
	updatesc chan []*naming.Update
}

func NewSimpleResolver(addrs []string) (*SimpleResolver, base.Error) {
	if addrs == nil || len(addrs) == 0 {
		return nil, base.NewError(-1, err_scope_balance, "addrs is nil")
	}
	resolver := &SimpleResolver{
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

func (this *SimpleResolver) Resolve(target string) (naming.Watcher, error) {
	return this, nil
}

func (this *SimpleResolver) Next() ([]*naming.Update, error) {
	return <-this.updatesc, nil
}

func (this *SimpleResolver) Close() {
	close(this.updatesc)
}
