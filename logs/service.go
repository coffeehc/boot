package logs

import (
	"os"
	"time"

	"git.xiagaogao.com/coffee/boot"
	"github.com/spf13/pflag"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logLevel = pflag.String("logger_level", "info", "日志级别(debug,warn,info,error)")

var levelMap = map[string]zapcore.Level{
	"debug": zapcore.DebugLevel,
	"warn":  zapcore.WarnLevel,
	"info":  zapcore.InfoLevel,
	"error": zapcore.ErrorLevel,
}

type Service interface {
	GetLogger() *zap.Logger
	SetLevel(level zapcore.Level)
	NewLogger(skip int) *zap.Logger
}

func NewService(serviceInfo boot.ServiceInfo) (Service, error) {
	level := zap.NewAtomicLevelAt(zap.InfoLevel)
	levelStr, ok := os.LookupEnv("ENV_LOGGER_LEVEL")
	if ok {
		if l, ok1 := levelMap[levelStr]; ok1 {
			level.SetLevel(l)
		}
	} else {
		if l, ok1 := levelMap[*logLevel]; ok1 {
			level.SetLevel(l)
		}
	}
	writerSync, err := newMQWriterSync()
	if err != nil {
		return nil, err
	}
	logger := newLogger(level, writerSync, 0)
	logger = logger.WithOptions(zap.Fields(zap.String(K_ServiceName, serviceInfo.ServiceName)))
	return &serviceImpl{
		logger:     logger,
		level:      level,
		writerSync: writerSync,
	}, nil
}

type serviceImpl struct {
	logger     *zap.Logger
	level      zap.AtomicLevel
	writerSync ExtWriterSync
	loggers    map[int]*zap.Logger
}

func (impl *serviceImpl) GetLogger() *zap.Logger {
	return impl.logger
}

func (impl *serviceImpl) SetLevel(level zapcore.Level) {
	impl.level.SetLevel(level)
}

func (impl *serviceImpl) NewLogger(skip int) *zap.Logger {
	logger, ok := impl.loggers[skip]
	if ok {
		return logger
	}
	logger = newLogger(impl.level, impl.writerSync, skip)
	impl.loggers[skip] = logger
	return logger
}

type ExtWriterSync interface {
	zapcore.WriteSyncer
	ExtWrite(bs []byte, ent zapcore.Entry) (int, error)
}

func newMQWriterSync() (ExtWriterSync, error) {
	ws := &mqWriterSync{writerSync: zapcore.AddSync(os.Stdout)}
	//mqaddr, ok := os.LookupEnv("ENV_MQ_ADDR")
	//if !ok {
	//	return ws, errors.NewError(errors.Error_Message, "bus", "没有设置mq的地址")
	//}
	//config := &mqservice.VhostConfig{
	//	Vhost:    "bus",
	//	User:     "bus",
	//	Password: "bus#123",
	//	MQAddr:   mqaddr,
	//}
	//producer, err := mqservice.NewProducer("logs", config, 5, true, func(ctx context.Context, channel *amqp.Channel) errors.Error {
	//	err := channel.ExchangeDeclare("logs", "topic", true, false, false, false, nil)
	//	if err != nil {
	//		return errors.NewErrorWrapper(errors.Error_System, "bus", err)
	//	}
	//	_, err = channel.QueueDeclare("alllogs", true, false, false, false, nil)
	//	if err != nil {
	//		return errors.NewErrorWrapper(errors.Error_System, "bus", err)
	//	}
	//	err = channel.QueueBind("alllogs", "log.#", "logs", false, nil)
	//	if err != nil {
	//		return errors.NewErrorWrapper(errors.Error_System, "bus", err)
	//	}
	//	return nil
	//})
	//if err != nil {
	//	return nil, err
	//}
	//ws.producer = producer
	return ws, nil
}

type mqWriterSync struct {
	//producer   mqservice.Producer
	writerSync zapcore.WriteSyncer
}

func (impl *mqWriterSync) Write(bs []byte) (int, error) {
	return 0, nil
}

func (impl *mqWriterSync) ExtWrite(bs []byte, ent zapcore.Entry) (int, error) {
	//if impl.producer == nil {
	return impl.writerSync.Write(bs)
	//}
	//TODO这里需要考虑堵塞的问题
	//impl.producer.AsyncPublish("logs", "logs.all", bs)
	//return len(bs), nil
}

func (impl *mqWriterSync) Sync() error {
	return nil
}

func newLogger(level zap.AtomicLevel, writerSync ExtWriterSync, skip int) *zap.Logger {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:       K_Time,
		LevelKey:      K_level,
		NameKey:       K_Name,
		CallerKey:     K_Call,
		MessageKey:    K_Message,
		StacktraceKey: K_Stacktrace,
		LineEnding:    zapcore.DefaultLineEnding,
		EncodeLevel:   zapcore.CapitalLevelEncoder,
		EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(t.Format("2006-01-02T15:04:05.000"))
		},
		EncodeDuration: zapcore.NanosDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
	encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	encoder := zapcore.NewConsoleEncoder(encoderConfig)
	if boot.IsProductModel() {
		encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}
	loggerCore := &loggerCore{
		LevelEnabler: level,
		enc:          encoder,
		out:          writerSync,
	}
	opts := []zap.Option{zap.AddCaller(), zap.AddStacktrace(zapcore.DPanicLevel), zap.AddCallerSkip(skip)}
	//opts = append(opts, zap.WrapCore(func(core zapcore.Core) zapcore.Core {
	//	return zapcore.NewSampler(core, time.Second, 3, 10)
	//}))
	return zap.New(loggerCore, opts...)
}
