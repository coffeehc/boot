package base

const (
	//ErrCodeScopeBase 基础的 ErrCode
	ErrCodeScopeBase = 0x00000000
)

const (
	//ErrCodeScopeBaseSystem 系统级别的 ErrCode
	ErrCodeScopeBaseSystem = ErrCodeScopeBase | 0x00000100
	//ErrCodeScopeBaseRPC RPC级别的 ErrCode
	ErrCodeScopeBaseRPC = ErrCodeScopeBase | 0x00000200

	ErrCodeScopeBaseMessage = ErrCodeScopeBase | 0x00000300
)

var (
	ErrCodeBaseMessage int64 = ErrCodeScopeBaseMessage | 0x0
)

var (
	//ErrCodeBaseSystemUnknown 未知错误
	ErrCodeBaseSystemUnknown int64 = ErrCodeScopeBaseSystem | 0x0
	//ErrCodeBaseSystemInit 初始化失败
	ErrCodeBaseSystemInit int64 = ErrCodeScopeBaseSystem | 0x1
	//ErrCodeBaseSystemConfig 配置错误
	ErrCodeBaseSystemConfig int64 = ErrCodeScopeBaseSystem | 0x2
	//ErrCodeBaseSystemInvalidParam 无效参数
	ErrCodeBaseSystemInvalidParam int64 = ErrCodeScopeBaseSystem | 0x4
	//ErrCodeBaseSystemMarshal 序列化错误
	ErrCodeBaseSystemMarshal int64 = ErrCodeScopeBaseSystem | 0x5
	//ErrCodeBaseSystemUnmarshal 反序列化错误
	ErrCodeBaseSystemUnmarshal int64 = ErrCodeScopeBaseSystem | 0x6
	//ErrCodeBaseSystemEncode 编码错误
	ErrCodeBaseSystemEncode int64 = ErrCodeScopeBaseSystem | 0x7
	//ErrCodeBaseSystemDecode 解码错误
	ErrCodeBaseSystemDecode int64 = ErrCodeScopeBaseSystem | 0x8
	//ErrCodeBaseSystemServiceRegister 注册服务失败
	ErrCodeBaseSystemServiceRegister int64 = ErrCodeScopeBaseSystem | 0x9
	//ErrCodeBaseSystemTypeConversion 类型转换错误
	ErrCodeBaseSystemTypeConversion int64 = ErrCodeScopeBaseSystem | 0x1
	//ErrCodeBaseSystemNil 空指针
	ErrCodeBaseSystemNil int64 = ErrCodeScopeBaseSystem | 0xb
	//ErrCodeBaseSystemDB DB错误
	ErrCodeBaseSystemDB int64 = ErrCodeScopeBaseSystem | 0xc
)
var (
	//ErrCodeBaseRPCUnknown 请求错误,原因未知
	ErrCodeBaseRPCUnknown int64 = ErrCodeScopeBaseRPC | 0x0
	//ErrCodeBaseRPCCancelled 请求被取消
	ErrCodeBaseRPCCancelled int64 = ErrCodeScopeBaseRPC | 0x3
	//ErrCodeBaseRPCInvalidArgument 无效参数
	ErrCodeBaseRPCInvalidArgument int64 = ErrCodeScopeBaseRPC | 0x4
	//ErrCodeBaseRPCTimeout 超时
	ErrCodeBaseRPCTimeout int64 = ErrCodeScopeBaseRPC | 0x5
	//ErrCodeBaseRPCNotFount 没有找到资源
	ErrCodeBaseRPCNotFount int64 = ErrCodeScopeBaseRPC | 0x6
	//ErrCodeBaseRPCAlreadyExests 已经存在的实例再创建的时候错误
	ErrCodeBaseRPCAlreadyExests int64 = ErrCodeScopeBaseRPC | 0x7
	//ErrCodeBaseRPCPermissionDenied 权限不足
	ErrCodeBaseRPCPermissionDenied int64 = ErrCodeScopeBaseRPC | 0x8
	//ErrCodeBaseRPCResourceExhausted 资源耗尽
	ErrCodeBaseRPCResourceExhausted int64 = ErrCodeScopeBaseRPC | 0x9
	//ErrCodeBaseRPCFailedPrecondition 前置条件失败
	ErrCodeBaseRPCFailedPrecondition int64 = ErrCodeScopeBaseRPC | 0xa
	//ErrCodeBaseRPCAborted 中途失败
	ErrCodeBaseRPCAborted int64 = ErrCodeScopeBaseRPC | 0xb
	//ErrCodeBaseRPCOutOfRange 超出范围
	ErrCodeBaseRPCOutOfRange int64 = ErrCodeScopeBaseRPC | 0xc
	//ErrCodeBaseRPCUnImplemented 没有实现
	ErrCodeBaseRPCUnImplemented int64 = ErrCodeScopeBaseRPC | 0xd
	//ErrCodeBaseRPCInternal 内部错误
	ErrCodeBaseRPCInternal int64 = ErrCodeScopeBaseRPC | 0xe
	//ErrCodeBaseRPCUnAvailable 服务不可用
	ErrCodeBaseRPCUnAvailable int64 = ErrCodeScopeBaseRPC | 0xf
	//ErrCodeBaseRPCDataLoss 数据丢失
	ErrCodeBaseRPCDataLoss int64 = ErrCodeScopeBaseRPC | 0x10
	//ErrCodeBaseRPCUnAuthenticated 没有认证
	ErrCodeBaseRPCUnAuthenticated int64 = ErrCodeScopeBaseRPC | 0x11
)
