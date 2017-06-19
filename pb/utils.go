package pb

import (
	"reflect"

	"github.com/coffeehc/logger"
	"github.com/golang/protobuf/proto"
	google_protobuf "github.com/golang/protobuf/ptypes/any"
)

var (
	registry = make(map[string]reflect.Type)
)

// RegisterType 例子registry["HttpCheck"] = reflect.TypeOf(HttpCheck{})
func RegisterType(class string, t reflect.Type) {
	registry[class] = t
}

func UnmarshalAny(any *google_protobuf.Any) (interface{}, error) {
	class := any.TypeUrl
	bytes := any.Value

	instance := reflect.New(registry[class]).Interface()
	err := proto.Unmarshal(bytes, instance.(proto.Message))
	if err != nil {
		return nil, err
	}
	logger.Debug("instance: %v", instance)

	return instance, nil
}

func MarshalAny(class string, message proto.Message) (*google_protobuf.Any, error) {
	data, err := proto.Marshal(message)
	if err != nil {
		return nil, err
	}
	return &google_protobuf.Any{
		TypeUrl: class,
		Value:   data,
	}, nil
}
