package restboot

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/coffeehc/httpx"
	"github.com/coffeehc/logger"
	"github.com/coffeehc/microserviceboot/base"
	"github.com/golang/protobuf/proto"
	"github.com/pquerna/ffjson/ffjson"
)

const errScopeRest = "restRequest"

//ErrorRecover 对 err 进行转换
func ErrorRecover(reply httpx.Reply) {
	if err := recover(); err != nil {
		logger.Error("处理请求,发生错误:%s", err)
		var errorResponse base.Error
		statusCode := http.StatusInternalServerError
		switch e := err.(type) {
		case base.Error:
			statusCode = http.StatusBadRequest
			errorResponse = e
		case string:
			errorResponse = base.NewError(base.ErrCodeBaseRPCUnknown, errScopeRest, e)
		case error:
			errorResponse = base.NewErrorWrapper(errScopeRest, base.ErrCodeBaseRPCUnknown, e)
		default:
			errorResponse = base.NewError(base.ErrCodeBaseRPCUnknown, errScopeRest, fmt.Sprintf("%#v", err))
		}
		//暂时统一按照400处理
		reply.SetStatusCode(statusCode).With(errorResponse).As(httpx.DefaultRenderJSON)
	}
}

//UnmarshalWhitJSON 将 request 的 Json内容解析为 对象
func UnmarshalWhitJSON(request *http.Request, data interface{}) {
	dataBytes, err := ioutil.ReadAll(request.Body)
	if err != nil {
		panic(err)
	}
	err = ffjson.Unmarshal(dataBytes, data)
	if err != nil {
		panic(err)
	}
}

//UnmarshalWhitProtobuf 将 request 的 Protobuf内容解析为 对象
func UnmarshalWhitProtobuf(request *http.Request, data proto.Message) {
	dataBytes, err := ioutil.ReadAll(request.Body)
	if err != nil {
		panic(err)
	}
	err = proto.Unmarshal(dataBytes, data)
	if err != nil {
		panic(err)
	}
}

//PanicErr 如果 err 不为空,直接 Panic
func PanicErr(err error) {
	if err != nil {
		panic(err)
	}
}

//ParsePathParamToBinary 解析Path 参数解析为二进制数据
func ParsePathParamToBinary(pathFragments map[string]string, name string) []byte {
	str, ok := pathFragments[name]
	if !ok {
		panic(base.NewError(base.ErrCodeBaseRPCInvalidArgument, errScopeRest, fmt.Sprintf("没有指定%s值", name)))
	}
	data, err := base64.RawURLEncoding.DecodeString(str)
	if err != nil {
		panic(base.NewError(base.ErrCodeBaseRPCInvalidArgument, errScopeRest, fmt.Sprintf("无法解析%s", name)))
	}
	return data
}
