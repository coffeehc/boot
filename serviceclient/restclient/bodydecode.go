package restclient

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/url"

	"github.com/coffeehc/microserviceboot/base"
	"github.com/golang/protobuf/proto"
)

const err_scope_rest_response = "rest response"

type ResponseBodyDecoder func(body io.ReadCloser, target interface{}) base.Error

func ResponseFormBodyDecoder(body io.ReadCloser, target interface{}) base.Error {
	defer body.Close()
	if vs, ok := target.(url.Values); ok {
		data, err := ioutil.ReadAll(body)
		if err != nil {
			return base.NewErrorWrapper(0, err_scope_rest_response, err)
		}
		values, err1 := url.ParseQuery(string(data))
		if err1 != nil {
			return base.NewErrorWrapper(0, err_scope_rest_response, err1)
		}
		for k, vss := range values {
			for _, v := range vss {
				vs.Add(k, v)
			}
		}
		return nil
	}
	return base.NewError(-1, err_scope_rest_response, "target type is not url.Value")
}

func ResponsePBBodyDecoder(body io.ReadCloser, target interface{}) base.Error {
	data, err := ioutil.ReadAll(body)
	if err != nil {
		return base.NewErrorWrapper(0, err_scope_rest_response, err)
	}
	if message, ok := target.(proto.Message); ok {
		err = proto.Unmarshal(data, message)
		if err != nil {
			return base.NewErrorWrapper(0, err_scope_rest_response, err)
		}
		return nil
	}
	return base.NewError(-1, err_scope_rest_response, "target type is not proto.Message")
}

func ResponseJsonBodyDecoder(body io.ReadCloser, target interface{}) base.Error {
	data, err := ioutil.ReadAll(body)
	if err != nil {
		return base.NewErrorWrapper(0, err_scope_rest_response, err)
	}
	err = json.Unmarshal(data, target)
	if err != nil {
		return base.NewErrorWrapper(0, err_scope_rest_response, err)
	}
	return nil
}
