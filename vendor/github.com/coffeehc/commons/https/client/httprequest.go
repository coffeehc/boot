package client

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
)

func NewHTTPRequest(method, urlStr string) (HTTPRequest, error) {
	req, err := http.NewRequest(method, urlStr, nil)
	if err != nil {
		return nil, err
	}
	return &_HTTPRequest{
		req: req,
	}, nil
}

type _HTTPRequest struct {
	req       *http.Request
	cookieJar http.CookieJar
	transport *http.Transport
}

func (_req *_HTTPRequest) SetTransport(transport *http.Transport) {
	_req.transport = transport
}

func (_req *_HTTPRequest) SetMethod(method string) {
	_req.req.Method = method
}
func (_req *_HTTPRequest) SetHeader(k, v string) {
	_req.req.Header.Set(k, v)
}
func (_req *_HTTPRequest) SetCookieJar(cookieJar http.CookieJar) {
	_req.cookieJar = cookieJar
}
func (_req *_HTTPRequest) SetBody(body []byte) {
	_req.req.Body = ioutil.NopCloser(bytes.NewReader(body))
}
func (_req *_HTTPRequest) SetBodyStream(reader io.ReadCloser) {
	_req.req.Body = reader
}

func (_req *_HTTPRequest) SetURI(requestURL string) error {
	_url, err := url.ParseRequestURI(requestURL)
	if err != nil {
		return err
	}
	_req.req.URL = _url
	return nil
}
func (_req *_HTTPRequest) SetBasicAuth(username, password string) {
	_req.req.SetBasicAuth(username, password)
}
func (_req *_HTTPRequest) SetContentType(contentType string) {
	_req.req.Header.Set("Content-Type", contentType)
}
func (_req *_HTTPRequest) SetCookie(cookie *http.Cookie) {
	_req.req.AddCookie(cookie)
}
func (_req *_HTTPRequest) SetReferer(referer string) {
	_req.req.Header.Set("referer", referer)
}
func (_req *_HTTPRequest) SetUserAgent(userAgent string) {
	_req.req.Header.Set("user-agent", userAgent)
}

func (_req *_HTTPRequest) SetProto(proto string) {
	_req.req.Proto = proto
}
func (_req *_HTTPRequest) GetRealRequest() *http.Request {
	return _req.req
}
