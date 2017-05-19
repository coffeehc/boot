package etcdtool

import (
	"context"
	"math/rand"
	"sync"
	"time"

	"github.com/coffeehc/logger"
	"github.com/coffeehc/microserviceboot/base"
	"github.com/coffeehc/microserviceboot/loadbalancer"
	"github.com/coreos/etcd/clientv3"
	"github.com/pquerna/ffjson/ffjson"
	"google.golang.org/grpc/naming"
)

type _EtcdResolver struct {
	client         *clientv3.Client
	service        string
	tag            string
	registerPrefix string

	quitc       chan struct{}
	quitUpdate  chan struct{}
	updatesc    chan []*naming.Update
	updateMutex *sync.Mutex
}

func newEtcdResolver(client *clientv3.Client, service, tag string) (naming.Resolver, base.Error) {
	r := &_EtcdResolver{
		client:         client,
		service:        service,
		tag:            tag,
		registerPrefix: buildServiceKeyPrefix(service),
		quitc:          make(chan struct{}),
		quitUpdate:     make(chan struct{}),
		updatesc:       make(chan []*naming.Update, 1),
		updateMutex:    new(sync.Mutex),
	}

	// Retrieve instances immediately
	instancesCh := make(chan []string)
	go func() {
		sleep := int64(time.Second * 3)
		for {
			instances, err := r.getInstances()
			if err != nil {
				logger.Warn("lb: error retrieving instances from etcd: %v", err)
				time.Sleep(time.Duration(rand.Int63n(sleep)))
				continue
			}
			logger.Debug("初始化instance is %q", instances)
			instancesCh <- instances
			return
		}
	}()
	r.updatesc <- r.makeUpdates(nil, <-instancesCh)
	//go r.updater(instances)
	return r, nil
}

func (r *_EtcdResolver) Resolve(target string) (naming.Watcher, error) {
	return r, nil
}

func (r *_EtcdResolver) Next() ([]*naming.Update, error) {
	return <-r.updatesc, nil
}

func (r *_EtcdResolver) Close() {
	select {
	case <-r.quitc:
	default:
		close(r.quitc)
		close(r.updatesc)
	}
}

func (r *_EtcdResolver) updater(instances []string) {
	var err error
	var oldInstances = instances
	var newInstances []string

	// TODO Cache the updates for a while, so that we don't overwhelm Consul.
	sleep := int64(time.Second * 3)
	for {
		select {
		case <-r.quitc:
			break
		case <-r.quitUpdate:
			return
		default:
			func() {
				defer func() {
					if err := recover(); err != nil {
						logger.Warn("update addrs error :%s", err)
					}
				}()
				newInstances, err = r.getInstances()
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

func (r *_EtcdResolver) getInstances() ([]string, error) {
	response, err := r.client.KV.Get(context.Background(), r.registerPrefix, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	address := []string{}
	for _, kv := range response.Kvs {
		logger.Debug("value is %s", kv.Value)
		info := &ServiceRegisterInfo{}
		err := ffjson.Unmarshal(kv.Value, info)
		if err != nil {
			logger.Error("Unmarshal er is %s", err)
			continue
		}
		logger.Debug("info is %#v", info)
		if info.ServiceInfo.Tag == r.tag {
			address = append(address, string(kv.Key[len(r.registerPrefix):]))
		}
	}
	return address, nil
}

func (r *_EtcdResolver) makeUpdates(oldInstances, newInstances []string) []*naming.Update {
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

func (sr *_EtcdResolver) Delete(addr loadbalancer.Address) {
	sr.updateMutex.Lock()
	defer sr.updateMutex.Unlock()
	logger.Warn("delete addr [%s]", addr.Addr)
	sr.updatesc <- []*naming.Update{&naming.Update{Op: naming.Delete, Addr: addr.Addr}}
	sr.quitUpdate <- struct{}{}
	time.Sleep(time.Second * 10)
	instances := make([]string, 0)
	go sr.updater(instances)
}
