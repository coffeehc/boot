package serviceboot

import (
	"fmt"
	"net/http"

	"encoding/json"
	"github.com/coffeehc/microserviceboot/base"
	"github.com/coffeehc/web"
	"io/ioutil"
	"github.com/golang/protobuf/proto"
	"encoding/base64"
	"github.com/coffeehc/logger"
)

func ErrorRecover(reply web.Reply) {
	if err := recover(); err != nil {
		logger.Error("处理请求,发生错误:%s",err)
		reply.As(web.Transport_Json)
		switch e := err.(type) {
		case base.Error:
			reply.With(base.NewErrorResponse(&e)).
				SetStatusCode(http.StatusBadRequest)
		case *base.Error:
			reply.With(base.NewErrorResponse(e)).
				SetStatusCode(http.StatusBadRequest)
		case base.BizErr:
			reply.With(base.NewErrorResponse(e.ToError())).
				SetStatusCode(int(e.GetHttpCode()))
		case *base.BizErr:
			reply.With(base.NewErrorResponse(e.ToError())).
				SetStatusCode(int(e.GetHttpCode()))
		case string:
			reply.With(base.NewErrorResponse(base.NewSimpleError(http.StatusBadRequest, e))).
				SetStatusCode(http.StatusBadRequest)
		case error:
			reply.With(base.NewErrorResponse(base.NewSimpleError(http.StatusInternalServerError, e.Error()))).
				SetStatusCode(http.StatusInternalServerError)
		default:
			reply.With(base.NewErrorResponse(base.NewSimpleError(http.StatusInternalServerError, fmt.Sprintf("%#v", err)))).
				SetStatusCode(http.StatusInternalServerError)
		}
		//暂时统一按照400处理
	}
}

//将 request 的 Json内容解析为 对象
func UnmarshalWhitJson(request *http.Request, data interface{}) {
	dataBytes, err := ioutil.ReadAll(request.Body)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(dataBytes, data)
	if err != nil {
		panic(err)
	}
}

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

//如果 err 不为空,直接 Panic
func PanicErr(err error) {
	if err != nil {
		panic(err)
	}
}

func ParsePathParamToBinary(pathFragments map[string]string,name string)[]byte{
	str, ok := pathFragments[name]
	if !ok {
		panic(base.BuildBizErr(fmt.Sprintf("没有指定%s值",name)))
	}
	data,err:=base64.RawURLEncoding.DecodeString(str)
	if err!=nil{
		panic(base.BuildBizErr(fmt.Sprintf("无法解析%s",name)))
	}
	return data
}
