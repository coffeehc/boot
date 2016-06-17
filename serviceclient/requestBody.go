package serviceclient

import (
	"io"

	"github.com/go-resty/resty"
	"github.com/golang/protobuf/proto"
	"gopkg.in/square/go-jose.v1/json"
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

type RequestJsonBody struct {
	data interface{}
}

func (this RequestJsonBody) SetBody(request *resty.Request) {
	d, _ := json.Marshal(this.data)
	request.SetBody(d).SetHeader("Content-Type", "application/json")
}

func NewRequestJsonBody(data interface{}) RequestBody {
	return &RequestJsonBody{data}
}

type RequestPBBody struct {
	data proto.Message
}

func (this RequestPBBody) SetBody(request *resty.Request) {
	d, _ := proto.Marshal(this.data)
	request.SetBody(d).SetHeader("Content-Type", "application/x-protobuf")
}

func NewRequestPBBody(data proto.Message) RequestBody {
	return &RequestPBBody{data}
}
