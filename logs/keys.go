package logs

import (
	"go.uber.org/zap"
)

const (
	K_Time         = "t"
	K_level        = "l"
	K_Name         = "n"
	K_Call         = "c"
	K_Stacktrace   = "s"
	K_Message      = "m"
	K_Cause        = "ca"
	K_ServiceName  = "sn"
	K_ServiceScope = "sc"
	K_ErrorCode    = "ec"
	K_ExtendData   = "ed"
	K_AccessUrl    = "au"
	K_Account      = "ac"
)

func F_Account(accountId int64) zap.Field {
	return zap.Int64(K_Account, accountId)
}

func F_ExtendData(extData interface{}) zap.Field {
	return zap.Any(K_ExtendData, extData)
}

func F_Error(err error) zap.Field {
	return zap.String(K_Cause, err.Error())
}

func F_ErrorStr(err string) zap.Field {
	return zap.String(K_Cause, err)
}
