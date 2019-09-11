package errors

func buildError(errorCode int64, message string) Error {
	return &baseError{
		Code:    errorCode,
		Message: message,
	}
}
func SystemError(message string) Error {
	return buildError(ErrorSystem, message)
}

func MessageError(message string) Error {
	return buildError(ErrorMessage, message)
}

func WrappedError(errorCode int64, err error) Error {
	return &baseError{
		Code:    errorCode,
		Message: err.Error(),
	}
}

func WrappedSystemError(err error) Error {
	return WrappedError(ErrorSystem, err)
}

func WrappedMessageError(err error) Error {
	return WrappedError(ErrorMessage, err)
}
