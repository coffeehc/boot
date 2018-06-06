package loadbalancer

import (
	"crypto/tls"
	"net"
	"time"

	"git.xiagaogao.com/coffee/boot/errors"
	"google.golang.org/grpc/naming"
)

var defaultTlsConfig = &tls.Config{
	InsecureSkipVerify: true,
}

func NewAddrArrayBalancer(ctx, addrs []string, ssl bool) (Balancer, errors.Error) {
	r, err := newAddrArrayResolver(addrs, ssl)
	if err != nil {
		return nil, err
	}
	return RoundRobin(ctx, r), nil
}

type addrArrayResolver struct {
	updatesc    chan []*naming.Update
	addrMonitor map[string]struct{}
	ssl         bool
}

func newAddrArrayResolver(addrs []string, ssl bool) (*addrArrayResolver, errors.Error) {
	if addrs == nil || len(addrs) == 0 {
		return nil, errors.NewError(-1, errScopeBalance, "addrs is nil")
	}
	resolver := &addrArrayResolver{
		updatesc:    make(chan []*naming.Update, 1),
		addrMonitor: make(map[string]struct{}, len(addrs)),
		ssl:         ssl,
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
	go resolver.monitorAddr()
	return resolver, nil
}

func (sr *addrArrayResolver) monitorAddr() {
	timeout := time.Second * 3
	timer := time.NewTimer(timeout)
	for {
		select {
		case <-timer.C:
			for addr, _ := range sr.addrMonitor {
				var conn net.Conn
				var err error
				if sr.ssl {
					conn, err = tls.Dial("tcp", addr, defaultTlsConfig)
				} else {
					conn, err = net.Dial("tcp", addr)
				}
				if err == nil && conn != nil {
					conn.Close()
					delete(sr.addrMonitor, addr)
					sr.updatesc <- []*naming.Update{&naming.Update{Op: naming.Add, Addr: addr}}
				}
			}
			timer.Reset(timeout)
		}
	}
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

func (sr *addrArrayResolver) Delete(addr Address) {
	sr.updatesc <- []*naming.Update{&naming.Update{Op: naming.Delete, Addr: addr.Addr}}
	//启动地址监控
	sr.addrMonitor[addr.Addr] = struct{}{}

}
