package client

import (
	"net/http"
	"net/http/cookiejar"
	"sync"
)

func NewCookieJarManager() CookieJarManager {
	return &_CookieJarManager{
		jars:    make(map[string]http.CookieJar),
		rwMutex: new(sync.RWMutex),
	}
}

type _CookieJarManager struct {
	jars    map[string]http.CookieJar
	rwMutex *sync.RWMutex
}

func (cjm *_CookieJarManager) GetCookieJar(key string) http.CookieJar {
	cjm.rwMutex.RLock()
	jar, ok := cjm.jars[key]
	cjm.rwMutex.RUnlock()
	if ok {
		return jar
	}
	cjm.rwMutex.Lock()
	jar, ok = cjm.jars[key]
	if !ok {
		jar, _ = cookiejar.New(nil)
		cjm.jars[key] = jar
	}
	cjm.rwMutex.Unlock()
	return jar
}

func (cjm *_CookieJarManager) RemoveCookieJar(key string) {
	cjm.rwMutex.Lock()
	defer cjm.rwMutex.Unlock()
	delete(cjm.jars, key)
}
