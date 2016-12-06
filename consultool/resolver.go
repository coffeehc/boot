package consultool

import (
	"github.com/coffeehc/logger"
	"github.com/coffeehc/microserviceboot/base"
	"github.com/hashicorp/consul/api"
	"google.golang.org/grpc/naming"
	"math/rand"
	"net"
	"strconv"
	"time"
)

type ConsulResolver struct {
	c           *api.Client
	service     string
	tag         string
	passingOnly bool

	quitc    chan struct{}
	updatesc chan []*naming.Update
}

// NewConsulResolver initializes and returns a new ConsulResolver.
//
// It resolves addresses for gRPC connections to the given service and tag.
// If the tag is irrelevant, use an empty string.
func NewConsulResolver(client *api.Client, service, tag string) (*ConsulResolver, base.Error) {
	r := &ConsulResolver{
		c:           client,
		service:     service,
		tag:         tag,
		passingOnly: true,
		quitc:       make(chan struct{}),
		updatesc:    make(chan []*naming.Update, 1),
	}

	// Retrieve instances immediately
	instances, index, err := r.getInstances(0)
	if err != nil {
		return nil, base.NewErrorWrapper(err)
	}
	r.updatesc <- r.makeUpdates(nil, instances)

	// Start updater
	go r.updater(instances, index)

	return r, nil
}

// Resolve creates a watcher for target. The watcher interface is implemented
// by ConsulResolver as well, see Next and Close.
func (r *ConsulResolver) Resolve(target string) (naming.Watcher, error) {
	return r, nil
}

// Next blocks until an update or error happens. It may return one or more
// updates. The first call will return the full set of instances available
// as NewConsulResolver will look those up. Subsequent calls to Next() will
// block until the resolver finds any new or removed instance.
//
// An error is returned if and only if the watcher cannot recover.
func (r *ConsulResolver) Next() ([]*naming.Update, error) {
	return <-r.updatesc, nil
}

// Close closes the watcher.
func (r *ConsulResolver) Close() {
	select {
	case <-r.quitc:
	default:
		close(r.quitc)
		close(r.updatesc)
	}
}

// updater is a background process started in NewConsulResolver. It takes
// a list of previously resolved instances (in the format of host:port, e.g.
// 192.168.0.1:1234) and the last index returned from Consul.
func (r *ConsulResolver) updater(instances []string, lastIndex uint64) {
	var err error
	var oldInstances = instances
	var newInstances []string

	// TODO Cache the updates for a while, so that we don't overwhelm Consul.
	sleep := int64(time.Second * 10)
	for {
		select {
		case <-r.quitc:
			break
		default:
			func() {
				defer func() {
					if err := recover(); err != nil {
						logger.Warn("update addrs error :%s", err)
					}
				}()
				newInstances, lastIndex, err = r.getInstances(lastIndex)
				if err != nil {
					logger.Warn("lb: error retrieving instances from Consul: %v", err)
					time.Sleep(time.Duration(rand.Int63n(sleep)))
					return
				}
				updates := r.makeUpdates(oldInstances, newInstances)
				if updates == nil || len(updates) == 0 {
					return
				}
				r.updatesc <- updates
				oldInstances = newInstances
			}()
		}
	}
}

// getInstances retrieves the new set of instances registered for the
// service from Consul.
func (r *ConsulResolver) getInstances(lastIndex uint64) ([]string, uint64, error) {
	services, meta, err := r.c.Health().Service(r.service, r.tag, r.passingOnly, &api.QueryOptions{
		WaitIndex: lastIndex,
	})
	if err != nil {
		return nil, lastIndex, err
	}

	var instances []string
	for _, service := range services {
		s := service.Service.Address
		if len(s) == 0 {
			s = service.Node.Address
		}
		addr := net.JoinHostPort(s, strconv.Itoa(service.Service.Port))
		instances = append(instances, addr)
	}
	return instances, meta.LastIndex, nil
}

// makeUpdates calculates the difference between and old and a new set of
// instances and turns it into an array of naming.Updates.
func (r *ConsulResolver) makeUpdates(oldInstances, newInstances []string) []*naming.Update {
	oldAddr := make(map[string]struct{}, len(oldInstances))
	for _, instance := range oldInstances {
		oldAddr[instance] = struct{}{}
	}
	newAddr := make(map[string]struct{}, len(newInstances))
	for _, instance := range newInstances {
		newAddr[instance] = struct{}{}
	}
	var updates []*naming.Update
	for addr := range newAddr {
		if _, ok := oldAddr[addr]; !ok {
			updates = append(updates, &naming.Update{Op: naming.Add, Addr: addr})
		}
	}
	for addr := range oldAddr {
		if _, ok := newAddr[addr]; !ok {
			updates = append(updates, &naming.Update{Op: naming.Delete, Addr: addr})
		}
	}
	return updates
}
