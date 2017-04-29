package client

import "net/http"

type HeaderSetting interface {
	Setting(http.Header)
	AddSetting(HeaderSetting) HeaderSetting
}

type defaultHeaderSetting struct {
	key   string
	value string
	prev  HeaderSetting
}

func (h *defaultHeaderSetting) Setting(header http.Header) {
	if h.key != "" {
		header.Set(h.key, h.value)
	}
	if h.prev != nil {
		h.prev.Setting(header)
	}
}
func (h *defaultHeaderSetting) AddSetting(hs HeaderSetting) HeaderSetting {
	h.prev = hs
	return hs
}

func NewHeaderUserAgent(v string) HeaderSetting {
	return &defaultHeaderSetting{
		key:   "User-Agent",
		value: v,
	}
}

func NewHeaderReferer(v string) HeaderSetting {
	return &defaultHeaderSetting{
		key:   "Referer",
		value: v,
	}
}
