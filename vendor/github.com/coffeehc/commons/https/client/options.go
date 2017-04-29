package client

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"sync"
	"time"
)

type HTTPClientOptions struct {
	Timeout                        time.Duration //对应从连接(Dial)到读完response body的整个时间(包含所有的redirect的时间)
	DialerTimeout                  time.Duration //限制建立TCP连接的时间
	DialerKeepAlive                time.Duration
	TransportTLSHandshakeTimeout   time.Duration //限制 TLS握手的时间
	TransportResponseHeaderTimeout time.Duration //限制读取response header的时间
	//TransportExpectContinueTimeout time.Duration //(可能引起不必要的风险,直接忽略这个值)限制client在发送包含 Expect: 100-continue的header到收到继续发送body的response之间的时间等待。注意在1.6中设置这个值会禁用HTTP/2(DefaultTransport自1.6.2起是个特例)
	TransportIdleConnTimeout     time.Duration //控制连接池中一个连接可以idle多长时间
	TransportMaxIdleConns        int
	TransportMaxIdleConnsPerHost int

	GlobalHeaderSetting HeaderSetting

	mutex *sync.Mutex
}

func (co *HTTPClientOptions) checkMuext() {
	if co.mutex == nil {
		co.mutex = new(sync.Mutex)
	}
}

func (co *HTTPClientOptions) AddHeaderSetting(hs HeaderSetting) {
	co.checkMuext()
	co.mutex.Lock()
	defer co.mutex.Unlock()
	if co.GlobalHeaderSetting == nil {
		co.GlobalHeaderSetting = hs
		return
	}
	co.GlobalHeaderSetting = co.GlobalHeaderSetting.AddSetting(hs)
}

func (co *HTTPClientOptions) GetTimeout() time.Duration {
	if co.Timeout == 0 {
		co.Timeout = 30 * time.Second
	}
	return co.Timeout
}
func (co *HTTPClientOptions) GetDialerTimeout() time.Duration {
	if co.DialerTimeout == 0 {
		co.DialerTimeout = 3 * time.Second
	}
	return co.DialerTimeout
}
func (co *HTTPClientOptions) GetDialerKeepAlive() time.Duration {
	//if co.DialerKeepAlive == 0 {
	//	co.DialerKeepAlive = 60 * time.Second
	//}
	return co.DialerKeepAlive
}
func (co *HTTPClientOptions) GetTransportTLSHandshakeTimeout() time.Duration {
	if co.TransportTLSHandshakeTimeout == 0 {
		co.TransportTLSHandshakeTimeout = 3 * time.Second
	}
	return co.TransportTLSHandshakeTimeout
}
func (co *HTTPClientOptions) GetTransportResponseHeaderTimeout() time.Duration {
	if co.TransportResponseHeaderTimeout == 0 {
		co.TransportResponseHeaderTimeout = 3 * time.Second
	}
	return co.TransportResponseHeaderTimeout
}

func (co *HTTPClientOptions) GetTransportIdleConnTimeout() time.Duration {
	if co.TransportIdleConnTimeout == 0 {
		co.TransportIdleConnTimeout = 90 * time.Second
	}
	return co.TransportIdleConnTimeout
}

func (co *HTTPClientOptions) GetTransportMaxIdleConns() int {
	if co.TransportMaxIdleConns == 0 {
		co.TransportMaxIdleConns = 1000
	}
	return co.TransportMaxIdleConns
}

func (co *HTTPClientOptions) GetTransportMaxIdleConnsPerHost() int {
	if co.TransportMaxIdleConnsPerHost == 0 {
		co.TransportMaxIdleConnsPerHost = 1000
	}
	return co.TransportMaxIdleConnsPerHost
}

func (co *HTTPClientOptions) setHeader(req *http.Request) {
	if co.GlobalHeaderSetting != nil {
		co.GlobalHeaderSetting.Setting(req.Header)
	}
}

func (co *HTTPClientOptions) NewDialer() *net.Dialer {
	return &net.Dialer{
		Timeout:   co.GetDialerTimeout(),
		KeepAlive: co.GetDialerKeepAlive(),
	}
}

func (co *HTTPClientOptions) NewTransport(dialContext func(ctx context.Context, network, address string) (net.Conn, error)) *http.Transport {
	disableKeepAlives := false
	if co.GetDialerKeepAlive() == 0 {
		disableKeepAlives = true
	}
	if dialContext == nil {
		dialContext = co.NewDialer().DialContext
	}
	return &http.Transport{
		DialContext:         dialContext,
		MaxIdleConnsPerHost: co.GetTransportMaxIdleConnsPerHost(),
		MaxIdleConns:        co.GetTransportMaxIdleConns(),
		IdleConnTimeout:     co.GetTransportIdleConnTimeout(),
		TLSHandshakeTimeout: co.GetTransportTLSHandshakeTimeout(),
		//ExpectContinueTimeout: 1 * time.Second,
		DisableKeepAlives: disableKeepAlives,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify:          true,
			DynamicRecordSizingDisabled: true,
		},
	}
}
