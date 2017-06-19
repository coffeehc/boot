package base

const (
	errCode_Base = 0x10000000

	// 系统级别的错误,包括IO异常,空指针,等
	ErrCode_System = errCode_Base | 0x10000
	//业务相关的异常
	ErrCode_Message = errCode_Base | 0x20000

	ErrCode_NotFount = ErrCode_Message | 0x1

	ErrCode_Internal = ErrCode_System | 0x1
	//ErrCodeScopeBaseRPC RPC级别的 ErrCode
	ErrCode_DB = ErrCode_System | 0x2

	ErrCode_Redis = ErrCode_System | 0x3
	//RPC错误,包含编解码
	ErrCode_RPC = ErrCode_System | 0x4
)

func isError(srcCode, targetCode int64) bool {
	return srcCode&targetCode == targetCode
}

func IsSystemError(code int64) bool {
	return isError(code, ErrCode_System)
}

func IsMessageError(code int64) bool {
	return isError(code, ErrCode_Message)
}

func IsDBError(code int64) bool {
	return isError(code, ErrCode_DB)
}

func IsRedisErro(code int64) bool {
	return isError(code, ErrCode_Redis)
}

func IsRPCError(code int64) bool {
	return isError(code, ErrCode_RPC)
}

func IsInternalError(code int64) bool {
	return isError(code, ErrCode_Internal)
}

func IsNotFountError(code int64) bool {
	return isError(code, ErrCode_NotFount)
}
