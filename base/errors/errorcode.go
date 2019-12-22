package errors

const (
	_baseError int64 = 0x10000000

	// 系统级别的错误,包括IO异常,空指针,等
	ErrorSystem = _baseError | 0x1000000
	// 业务相关的异常
	ErrorMessage = _baseError | 0x2000000

	ErrorMessageNotFount = ErrorMessage | 0x1

	ErrorSystemInternal = ErrorSystem | 0x1
	// ErrCodeScopeBaseRPC RPC级别的 ErrCode
	ErrorSystemDB = ErrorSystem | 0x2

	ErrorSystemRedis = ErrorSystem | 0x3
	// RPC错误,包含编解码
	ErrorSystemRPC = ErrorSystem | 0x4

	ErrorSystemNet = ErrorSystem | 0x5
)

func EqualError(srcCode, targetCode int64) bool {
	return srcCode&targetCode == targetCode
}

func IsBaseErrorCode(code int64) bool {
	return EqualError(code, _baseError)
}

func IsBaseError(err error) bool {
	if e, ok := err.(Error); ok {
		return EqualError(e.GetCode(), _baseError)
	}
	return false
}

func IsSystemError(err error) bool {
	if e, ok := err.(Error); ok {
		return EqualError(e.GetCode(), ErrorSystem)
	}
	return false
}

func IsMessageError(err error) bool {
	if e, ok := err.(Error); ok {
		return EqualError(e.GetCode(), ErrorMessage)
	}
	return false
}

func IsNetError(err error) bool {
	if e, ok := err.(Error); ok {
		return EqualError(e.GetCode(), ErrorSystemNet)
	}
	return false
}

func IsDBError(err error) bool {
	if e, ok := err.(Error); ok {
		return EqualError(e.GetCode(), ErrorSystemDB)
	}
	return false
}

func IsRedisErro(err error) bool {
	if e, ok := err.(Error); ok {
		return EqualError(e.GetCode(), ErrorSystemRedis)
	}
	return false
}

func IsRPCError(err error) bool {
	if e, ok := err.(Error); ok {
		return EqualError(e.GetCode(), ErrorSystemRPC)
	}
	return false
}

func IsInternalError(err error) bool {
	if e, ok := err.(Error); ok {
		return EqualError(e.GetCode(), ErrorSystemInternal)
	}
	return false
}

func IsNotFountError(err error) bool {
	if e, ok := err.(Error); ok {
		return EqualError(e.GetCode(), ErrorMessageNotFount)
	}
	return false
}
