package log

import "go.uber.org/zap"

var baseFields []zap.Field

func SetBaseFields(fields ...zap.Field) {
	mutex.Lock()
	baseFields = fields
	mutex.Unlock()
	resetLoggers()
}

// func AddBaseFields(fields ...zap.Field) {
// 	mutex.Lock()
// 	baseFields = append(baseFields, fields...)
// 	mutex.Unlock()
// 	resetLoggers()
// }

func resetLoggers() {
	defaultLogger = rootLogger.With(baseFields...)
	internalLogger = defaultLogger.WithOptions(zap.AddCallerSkip(2))
}
