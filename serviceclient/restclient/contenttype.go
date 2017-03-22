package restclient

import "github.com/pquerna/ffjson/ffjson"

var JsonContentType = &_JsonContentType{}

type ContentType interface {
	GetContentType() string
	Encode(v interface{}) ([]byte, error)
	Decoder(data []byte, v interface{}) error
}

type _JsonContentType struct {
}

func (jct _JsonContentType) GetContentType() string {
	return "application/json; charset=utf-8"
}
func (jct _JsonContentType) Encode(v interface{}) ([]byte, error) {
	return ffjson.Marshal(v)
}
func (jct _JsonContentType) Decoder(data []byte, v interface{}) error {
	return ffjson.Unmarshal(data, v)
}
