package errors

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

func equalError(srcCode, targetCode int32) bool {
	return srcCode&targetCode == targetCode
}

func IsBaseErrorCode(code int32) bool {
	return equalError(code, _baseError)
}

func IsBaseError(err error) bool {
	if e, ok := err.(Error); ok {
		return equalError(e.GetCode(), _baseError)
	}
	return false
}

func IsSystemError(err error) bool {
	if e, ok := err.(Error); ok {
		return equalError(e.GetCode(), Error_System)
	}
	return false
}

func IsMessageError(err error) bool {
	if e, ok := err.(Error); ok {
		return equalError(e.GetCode(), Error_Message)
	}
	return false
}

func IsDBError(err error) bool {
	if e, ok := err.(Error); ok {
		return equalError(e.GetCode(), Error_System_DB)
	}
	return false
}

func IsRedisErro(err error) bool {
	if e, ok := err.(Error); ok {
		return equalError(e.GetCode(), Error_System_Redis)
	}
	return false
}

func IsRPCError(err error) bool {
	if e, ok := err.(Error); ok {
		return equalError(e.GetCode(), Error_System_RPC)
	}
	return false
}

func IsInternalError(err error) bool {
	if e, ok := err.(Error); ok {
		return equalError(e.GetCode(), Error_System_Internal)
	}
	return false
}

func IsNotFountError(err error) bool {
	if e, ok := err.(Error); ok {
		return equalError(e.GetCode(), Error_Message_NotFount)
	}
	return false
}
