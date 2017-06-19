package base

const (
	_baseError = 0x10000000

	// 系统级别的错误,包括IO异常,空指针,等
	Error_System = _baseError | 0x10000
	//业务相关的异常
	Error_Message = _baseError | 0x20000

	Error_Message_NotFount = Error_Message | 0x1

	Error_System_Internal = Error_System | 0x1
	//ErrCodeScopeBaseRPC RPC级别的 ErrCode
	Error_System_DB = Error_System | 0x2

	Error_System_Redis = Error_System | 0x3
	//RPC错误,包含编解码
	Error_System_RPC = Error_System | 0x4
)

func isError(srcCode, targetCode int32) bool {
	return srcCode&targetCode == targetCode
}

func IsBaseError(code int32) bool {
	return isError(code, _baseError)
}

func IsSystemError(code int32) bool {
	return isError(code, Error_System)
}

func IsMessageError(code int32) bool {
	return isError(code, Error_Message)
}

func IsDBError(code int32) bool {
	return isError(code, Error_System_DB)
}

func IsRedisErro(code int32) bool {
	return isError(code, Error_System_Redis)
}

func IsRPCError(code int32) bool {
	return isError(code, Error_System_RPC)
}

func IsInternalError(code int32) bool {
	return isError(code, Error_System_Internal)
}

func IsNotFountError(code int32) bool {
	return isError(code, Error_Message_NotFount)
}
