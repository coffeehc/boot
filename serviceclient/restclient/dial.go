package restclient

import (
	"context"
	"net"
	"time"

	"github.com/coffeehc/microserviceboot/loadbalancer"
)

type _BalanceDialer struct {
	Timeout   time.Duration
	Deadline  time.Time
	KeepAlive time.Duration
	Cancel    <-chan struct{}
	balancer  loadbalancer.Balancer
}

func (d *_BalanceDialer) deadline(ctx context.Context, now time.Time) (earliest time.Time) {
	if d.Timeout != 0 { // including negative, for historical reasons
		earliest = now.Add(d.Timeout)
	}
	if d, ok := ctx.Deadline(); ok {
		earliest = minNonzeroTime(earliest, d)
	}
	return minNonzeroTime(earliest, d.Deadline)
}

func (d *_BalanceDialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	if ctx == nil {
		panic("nil context")
	}
	deadline := d.deadline(ctx, time.Now())
	if !deadline.IsZero() {
		if d, ok := ctx.Deadline(); !ok || deadline.Before(d) {
			subCtx, cancel := context.WithDeadline(ctx, deadline)
			defer cancel()
			ctx = subCtx
		}
	}
	if oldCancel := d.Cancel; oldCancel != nil {
		subCtx, cancel := context.WithCancel(ctx)
		defer cancel()
		go func() {
			select {
			case <-oldCancel:
				cancel()
			case <-subCtx.Done():
			}
		}()
		ctx = subCtx
	}
	addr, _, err := d.balancer.Get(ctx, loadbalancer.BalancerGetOptions{
		BlockingWait: true,
	})

	if err != nil {
		return nil, &net.OpError{Op: "dial", Net: network, Source: nil, Addr: nil, Err: err}
	}

	c, err := net.Dial(network, addr.Addr)
	if err != nil {
		return nil, err
	}

	if tc, ok := c.(*net.TCPConn); ok && d.KeepAlive > 0 {
		tc.SetKeepAlive(true)
		tc.SetKeepAlivePeriod(d.KeepAlive)
	}
	return c, nil
}

func minNonzeroTime(a, b time.Time) time.Time {
	if a.IsZero() {
		return b
	}
	if b.IsZero() || a.Before(b) {
		return a
	}
	return b
}
