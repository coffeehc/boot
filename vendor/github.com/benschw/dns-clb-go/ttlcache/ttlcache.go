package ttlcache

import (
	"github.com/benschw/dns-clb-go/dns"
	"net"
	"time"
	"sync"
)

func NewTtlCache(lib dns.Lookup, ttl int) *TtlCache {
	c := new(TtlCache)
	c.lib = lib
	c.ttl = ttl
	c.lastUpdate = 0
	c.rwmutex = new(sync.RWMutex)
	return c
}

type TtlCache struct {
	lib        dns.Lookup
	ttl        int
	lastUpdate int32
	srvs       []net.SRV
	as         map[string]string
	rwmutex *sync.RWMutex
}

func (l *TtlCache) LookupSRV(name string) ([]net.SRV, error) {
	err := l.checkCache()
	if err != nil {
		return nil, err
	}

	if len(l.srvs) == 0 {
		l.srvs, err = l.lib.LookupSRV(name)
		if err != nil {
			return nil, err
		}
	}
	return l.srvs, nil
}

func (l *TtlCache) LookupA(name string) (string, error) {
	err := l.checkCache()
	if err != nil {
		return "", err
	}
	l.rwmutex.RLock()
	_, ok := l.as[name]
	l.rwmutex.RUnlock()
	if !ok {
		l.rwmutex.Lock()
		l.as[name], err = l.lib.LookupA(name)
		l.rwmutex.Unlock()
		if err != nil {
			return "", err
		}
	}
	l.rwmutex.RLock()
	defer l.rwmutex.RUnlock()
	return l.as[name], nil
}

func (l *TtlCache) checkCache() error {
	now := int32(time.Now().Unix())
	if l.lastUpdate+int32(l.ttl) < now {
		l.lastUpdate = now
		l.srvs = []net.SRV{}
		l.as = map[string]string{}
	}
	return nil
}
