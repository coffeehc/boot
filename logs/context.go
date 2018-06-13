package logs

//const (
//	_LOGGER        = "_logger"
//	_LOGGERSERVICE = "_logger_service"
//)
//
//func GetLogger(ctx context.Context) *zap.Logger {
//	return ctx.Value(_LOGGER).(*zap.Logger)
//}
//
//func SetLogger(ctx context.Context, log *zap.Logger) context.Context {
//	return context.WithValue(ctx, _LOGGER, log)
//}
//
//func GetLoggerService(ctx context.Context) Service {
//	return ctx.Value(_LOGGERSERVICE).(Service)
//}
//
//func SetLoggerService(ctx context.Context, service Service) context.Context {
//	return context.WithValue(ctx, _LOGGERSERVICE, service)
//}
