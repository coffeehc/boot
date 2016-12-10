package restboot

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/coffeehc/logger"
	"github.com/coffeehc/microserviceboot/base"
	"github.com/coffeehc/web"
	"github.com/golang/protobuf/proto"
	"io/ioutil"
	"net/http"
)

const err_scope_rest  = "rest"

func ErrorRecover(reply web.Reply) {
	if err := recover(); err != nil {
		logger.Error("处理请求,发生错误:%s", err)
		var errorResponse *base.ErrorResponse
		switch e := err.(type) {
		case *base.ErrorResponse:
			errorResponse = e
		case base.ErrorResponse:
			errorResponse = &e
		case base.BaseError:
			errorResponse = base.NewErrorResponse(http.StatusBadRequest, e.GetErrorCode(), e.Error(), "")
		case *base.BaseError:
			errorResponse = base.NewErrorResponse(http.StatusBadRequest, e.GetErrorCode(), e.Error(), "")
		case base.Error:
			errorResponse = base.NewErrorResponse(http.StatusBadRequest, e.GetErrorCode(), e.Error(), "")
		case string:
			errorResponse = base.NewErrorResponse(http.StatusInternalServerError, base.ERROR_CODE_BASE_SYSTEM_ERROR, e, "")
		case error:
			errorResponse = base.NewErrorResponse(http.StatusInternalServerError, base.ERROR_CODE_BASE_SYSTEM_ERROR, e.Error(), "")
		default:
			errorResponse = base.NewErrorResponse(http.StatusInternalServerError, base.ERROR_CODE_BASE_SYSTEM_ERROR, fmt.Sprintf("%#v", err), "")
		}
		//暂时统一按照400处理
		reply.SetStatusCode(errorResponse.GetHttpCode()).With(errorResponse).As(web.Default_Render_Json)
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

func ParsePathParamToBinary(pathFragments map[string]string, name string) []byte {
	str, ok := pathFragments[name]
	if !ok {
		panic(base.NewError(base.ERROR_CODE_BASE_INVALID_PARAM,err_scope_rest, fmt.Sprintf("没有指定%s值", name)))
	}
	data, err := base64.RawURLEncoding.DecodeString(str)
	if err != nil {
		panic(base.NewError(base.ERROR_CODE_BASE_DECODE_ERROR,err_scope_rest, fmt.Sprintf("无法解析%s", name)))
	}
	return data
}
