package logs

import "go.uber.org/zap"

const K_Time = "_t"
const K_level = "_l"
const K_Name = "_n"
const K_Call = "_c"
const K_Stacktrace = "_s"
const K_Message = "_m"
const K_Cause = "_ca"
const K_ServiceName = "_sn"
const K_ServiceScope = "_sc"
const K_ErrorCode = "_ec"
const K_ExtendData = "_ed"

func F_ExtendData(extData interface{}) zap.Field {
	return zap.Any(K_ExtendData, extData)
}

func F_Error(err error) zap.Field {
	return zap.String(K_Cause, err.Error())
}
