package restclient

import ()
import (
	"net/url"
	"strings"
)

var portMap = map[string]string{
	"http":  "80",
	"https": "443",
}

// canonicalAddr returns url.Host but always with a ":port" suffix
func canonicalAddr(url *url.URL) string {
	addr := url.Host
	if !hasPort(addr) {
		return addr + ":" + portMap[url.Scheme]
	}
	return addr
}

func hasPort(s string) bool { return strings.LastIndex(s, ":") > strings.LastIndex(s, "]") }

func removeEmptyPort(host string) string {
	if hasPort(host) {
		return strings.TrimSuffix(host, ":")
	}
	return host
}
