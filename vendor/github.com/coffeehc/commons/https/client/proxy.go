package client

import (
	"bufio"
	"context"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"net"
	"net/http"
	"net/url"
	"strings"
	"unicode/utf8"

	"github.com/coffeehc/commons/convers"
	"golang.org/x/net/idna"
	"golang.org/x/net/proxy"
	"golang.org/x/text/unicode/norm"
	"golang.org/x/text/width"
)

var ErrorScope_ProxyConnect = "proxyconnect"

var NotSuportScheme = errors.New("not suport proxy scheme")

var defauleTLSConfig = &tls.Config{
	InsecureSkipVerify:          true,
	DynamicRecordSizingDisabled: true,
}

func NewProxyDialer(proxyTarget string, forward *net.Dialer) (*ProxyDialer, error) {
	_url, err := url.Parse(proxyTarget)
	if err != nil {
		return nil, err
	}
	scheme := _url.Scheme
	if scheme != "http" && scheme != "https" && scheme != "socks5" {
		return nil, NotSuportScheme
	}
	if forward == nil {
		forward = &net.Dialer{}
	}
	return &ProxyDialer{
		forward:    forward,
		proxyAddr:  canonicalAddr(_url),
		proxyType:  scheme,
		authHeader: getProxyAuthHeader(_url),
		userInfo:   _url.User,
	}, nil

}

type ProxyDialer struct {
	forward    *net.Dialer
	proxyAddr  string
	proxyType  string
	authHeader string
	userInfo   *url.Userinfo
}

func (d *ProxyDialer) Dial(network, address string) (net.Conn, error) {
	return d.DialContext(context.Background(), network, address)
}

func (d *ProxyDialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	switch d.proxyType {
	case "http":
		return d.connectHTTPProxy(ctx, network, address)
	case "https":
		return d.connectHTTPSProxy(ctx, network, address)
	case "socks5":
		return d.connectSOCKS5Proxy(ctx, network, address)
	}
	return nil, NotSuportScheme
}

func (d *ProxyDialer) connectSOCKS5Proxy(ctx context.Context, network, address string) (net.Conn, error) {
	var auth *proxy.Auth
	if d.userInfo != nil {
		psw, _ := d.userInfo.Password()
		auth = &proxy.Auth{
			User:     d.userInfo.Username(),
			Password: psw,
		}
	}
	dialer, err := proxy.SOCKS5("tcp", d.proxyAddr, auth, d.forward)
	if err != nil {
		return nil, &net.OpError{Op: ErrorScope_ProxyConnect, Net: "tcp", Err: err}
	}
	return dialer.Dial(network, address)
}

func (d *ProxyDialer) connectHTTPProxy(ctx context.Context, network, address string) (net.Conn, error) {
	conn, err := d.forward.DialContext(ctx, "tcp", d.proxyAddr)
	if err != nil {
		return nil, &net.OpError{Op: ErrorScope_ProxyConnect, Net: "tcp", Err: err}
	}
	return conn, nil
}

func (d *ProxyDialer) connectHTTPSProxy(ctx context.Context, network, address string) (net.Conn, error) {
	conn, err := d.forward.DialContext(ctx, "tcp", d.proxyAddr)
	if err != nil {
		return nil, &net.OpError{Op: ErrorScope_ProxyConnect, Net: "tcp", Err: err}
	}
	//握手实在CONECT之后的,看Transport的源码就知道了
	connectReq := &http.Request{
		Method: "CONNECT",
		URL:    &url.URL{Opaque: address},
		Host:   address,
		Header: make(http.Header),
	}
	if d.authHeader != "" {
		connectReq.Header.Set("Proxy-Authorization", d.authHeader)
	}
	connectReq.Write(conn)
	br := bufio.NewReader(conn)
	resp, err := http.ReadResponse(br, connectReq)
	if err != nil {
		conn.Close()
		return nil, err
	}
	if resp.StatusCode != 200 {
		f := strings.SplitN(resp.Status, " ", 2)
		conn.Close()
		return nil, errors.New(f[1])
	}
	return conn, nil
}

func getProxyAuthHeader(url *url.URL) string {
	userInfo := url.User
	if userInfo == nil {
		return ""
	}
	password, _ := userInfo.Password()
	return base64.StdEncoding.EncodeToString(convers.StringToBytes("Basic " + userInfo.Username() + ":" + password))
}

func canonicalAddr(url *url.URL) string {
	addr := url.Hostname()
	if v, err := idnaASCII(addr); err == nil {
		addr = v
	}
	port := url.Port()
	if port == "" {
		port = portMap[url.Scheme]
	}
	return net.JoinHostPort(addr, port)
}

func idnaASCII(v string) (string, error) {
	if isASCII(v) {
		return v, nil
	}
	// The idna package doesn't do everything from
	// https://tools.ietf.org/html/rfc5895 so we do it here.
	// TODO(bradfitz): should the idna package do this instead?
	v = strings.ToLower(v)
	v = width.Fold.String(v)
	v = norm.NFC.String(v)
	return idna.ToASCII(v)
}

func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] >= utf8.RuneSelf {
			return false
		}
	}
	return true
}

var portMap = map[string]string{
	"http":  "80",
	"https": "443",
}
