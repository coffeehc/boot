package client

import (
	"github.com/go-resty/resty"
	"io"
)

type RequestBody interface {
	SetBody(request *resty.Request)
}

type RequestFileBody struct {
	param    string
	fileName string
	file     io.Reader
}

func NewRequestFileBody(param, fileName string, file io.Reader) RequestBody {
	return &RequestFileBody{param: param, fileName: fileName, file: file}
}

func (this RequestFileBody) SetBody(request *resty.Request) {
	request.SetFileReader(this.param, this.fileName, this.file)
}

type RequestFormBody struct {
	data map[string]string
}

func NewRequestFormBody(data map[string]string) RequestBody {
	return &RequestFormBody{data: data}
}

func (this RequestFormBody) SetBody(request *resty.Request) {
	request.SetFormData(this.data)
}

type RequestDataBody struct {
	data interface{}
}

func NewRequestDataBody(data map[string]string) RequestBody {
	return &RequestDataBody{data: data}
}

func (this RequestDataBody) SetBody(request *resty.Request) {
	request.SetBody(this.data)
}
