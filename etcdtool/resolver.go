package etcdtool

import (
	"context"
	"sync"
	"time"

	"github.com/coffeehc/logger"
	"github.com/coffeehc/microserviceboot/base"
	"github.com/coffeehc/microserviceboot/loadbalancer"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
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
		registerPrefix: buildServiceKeyPrefix(service, tag),
		quitc:          make(chan struct{}),
		quitUpdate:     make(chan struct{}),
		updatesc:       make(chan []*naming.Update, 1),
		updateMutex:    new(sync.Mutex),
	}

	// Retrieve instances immediately
	go r.updater()
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

func (r *_EtcdResolver) updater() {
	instancesCh := make(chan []string)
	go func() {
		for {
			instances, err := r.getInstances()
			if err != nil {
				logger.Warn("lb: error retrieving instances from etcd: %v", err)
				time.Sleep(time.Second)
				continue
			}
			logger.Debug("初始化instance is %q", instances)
			instancesCh <- instances
			return
		}
	}()
	instances := <-instancesCh
	updates := make([]*naming.Update, 0)
	for _, instance := range instances {
		if instance != "" {
			updates = append(updates, &naming.Update{Op: naming.Add, Addr: instance})
		}
	}
	r.updatesc <- updates
	//watch
	logger.Debug("watch %s", r.registerPrefix)
	watchChan := r.client.Watch(context.Background(), r.registerPrefix, clientv3.WithPrefix(), clientv3.WithCreatedNotify())
	for {
		select {
		case <-r.quitc:
			break
		case <-r.quitUpdate:
			return
		case response, ok := <-watchChan:
			if !ok {
				logger.Debug("re wartch")
				watchChan = r.client.Watch(context.Background(), r.registerPrefix, clientv3.WithPrefix(), clientv3.WithCreatedNotify())
				break
			}
			updates := make([]*naming.Update, 0)
			for _, event := range response.Events {
				switch event.Type {
				case clientv3.EventTypePut:
					addr := r.getServiceAddr(event.Kv)
					if addr != "" {
						updates = append(updates, &naming.Update{Op: naming.Add, Addr: addr})
					}
				case clientv3.EventTypeDelete:
					updates = append(updates, &naming.Update{Op: naming.Delete, Addr: string(event.Kv.Key[len(r.registerPrefix):])})
				default:
					logger.Warn("无法识别的事件,%#v", event)
				}

			}
			if len(updates) > 0 {
				r.updatesc <- updates
			}
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
		addr := r.getServiceAddr(kv)
		if addr != "" {
			address = append(address, string(kv.Key[len(r.registerPrefix):]))
		}
	}
	return address, nil
}

func (r *_EtcdResolver) getServiceAddr(kv *mvccpb.KeyValue) string {
	logger.Debug("value is %s", kv.Value)
	info := &ServiceRegisterInfo{}
	err := ffjson.Unmarshal(kv.Value, info)
	if err != nil {
		logger.Error("Unmarshal er is %s", err)
		return ""
	}
	logger.Debug("info is %#v", info)
	if info.ServiceInfo.Tag == r.tag {
		return string(kv.Key[len(r.registerPrefix):])
	}
	return ""
}

func (sr *_EtcdResolver) Delete(addr loadbalancer.Address) {
	time.Sleep(time.Second)
}
