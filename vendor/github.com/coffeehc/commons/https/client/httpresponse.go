package client

import (
	"io"
	"net/http"
)

func newHTTPResponse(resp *http.Response) HTTPResponse {
	return &_HTTPResponse{
		resp: resp,
	}
}

type _HTTPResponse struct {
	resp *http.Response
}

func (r *_HTTPResponse) GetBody() io.ReadCloser {
	return r.resp.Body
}
func (r *_HTTPResponse) GetRealResponse() *http.Response {
	return r.resp
}

func (r *_HTTPResponse) GetHeader() http.Header {
	return r.resp.Header
}

func (r *_HTTPResponse) GetContentType() string {
	return r.resp.Header.Get("Content-Type")
}

func (r *_HTTPResponse) GetStatusCode() int {
	return r.resp.StatusCode
}
