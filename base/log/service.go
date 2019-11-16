package log

import (
	"os"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

func init() {
	InitLogger(true)
}

type Config struct {
	Level         string
	FileConfig    *FileLogConfig
	EnableConsole bool
	EnableColor   bool // 仅仅对Console有效
	EnableSampler bool
}

type FileLogConfig struct {
	FileName   string
	Disable    bool
	Maxsize    int
	MaxBackups int
	MaxAge     int
	Compress   bool
}

// 远程日志存储
var rootLogger *zap.Logger
var defaultLogger *zap.Logger
var internalLogger *zap.Logger
var level = zap.NewAtomicLevel()
var mutex = new(sync.Mutex)

func GetLogger() *zap.Logger {
	return defaultLogger
}

func InitLogger(force bool) {
	mutex.Lock()
	defer mutex.Unlock()
	if !force && rootLogger != nil {
		return
	}
	if rootLogger == nil {
		if !viper.IsSet("logger") {
			viper.SetDefault("logger", &Config{
				Level:         "debug",
				EnableConsole: true,
				EnableColor:   true,
			})
		}
		//  初始化本地化的日志
		encodeConfig := newEncodeConfig()
		encodeConfig.EncodeLevel = zapcore.LowercaseColorLevelEncoder
		core := zapcore.NewCore(zapcore.NewConsoleEncoder(encodeConfig), zapcore.AddSync(os.Stdout), zapcore.DebugLevel)
		rootLogger = zap.New(core, zap.AddStacktrace(zapcore.DPanicLevel), zap.AddCaller())
		zap.ReplaceGlobals(rootLogger)
	}
	conf := &Config{}
	err := viper.UnmarshalKey("logger", conf)
	if err != nil {
		if defaultLogger == nil {
			rootLogger.Fatal("解析日志配置失败", zap.Error(err))
		}
		rootLogger.Error("解析日志配置失败", zap.Error(err))
		return
	}
	logCores := make([]zapcore.Core, 0)
	var logLevel = zap.InfoLevel
	switch strings.ToLower(conf.Level) {
	case "debug":
		logLevel = zap.DebugLevel
	case "info":
		logLevel = zap.InfoLevel
	case "error":
		logLevel = zap.ErrorLevel
	case "dPanic":
		logLevel = zap.DPanicLevel
	case "panic":
		logLevel = zap.PanicLevel
	case "fatal":
		logLevel = zap.FatalLevel
	default:
		logLevel = zap.InfoLevel
	}
	level.SetLevel(logLevel)
	fileLogConfig := conf.FileConfig
	if fileLogConfig != nil && !fileLogConfig.Disable {
		if fileLogConfig.FileName == "" {
			rootLogger.Fatal("没有指定日志目录")
		}
		if fileLogConfig.MaxAge == 0 {
			fileLogConfig.MaxAge = 3
		}
		if fileLogConfig.Maxsize == 0 {
			fileLogConfig.Maxsize = 10
		}
		logFileWrite := &lumberjack.Logger{
			Filename:   fileLogConfig.FileName,
			MaxSize:    fileLogConfig.Maxsize,    // megabytes
			MaxBackups: fileLogConfig.MaxBackups, // 最多保留3个备份
			MaxAge:     fileLogConfig.MaxAge,     // days
			Compress:   fileLogConfig.Compress,   // 是否压缩 disabled by default
		}
		core := zapcore.NewCore(zapcore.NewJSONEncoder(newEncodeConfig()), zapcore.AddSync(logFileWrite), level)
		if conf.EnableSampler {
			core = zapcore.NewSampler(core, time.Second*5, 100, 10)
		}
		logCores = append(logCores, core)
	}
	if conf.EnableConsole {
		encodeConfig := newEncodeConfig()
		if conf.EnableColor {
			encodeConfig.EncodeLevel = zapcore.LowercaseColorLevelEncoder
		}
		core := zapcore.NewCore(zapcore.NewConsoleEncoder(encodeConfig), zapcore.AddSync(os.Stdout), level)
		if conf.EnableSampler {
			core = zapcore.NewSampler(core, time.Second*5, 100, 5)
		}
		logCores = append(logCores, core)
	}

	rootLogger = zap.New(zapcore.NewTee(logCores...), zap.AddStacktrace(zapcore.DPanicLevel), zap.AddCaller())
	resetLoggers()
	zap.ReplaceGlobals(rootLogger)
}

func WatchLevel() {
	viper.WatchConfig()
	viper.WatchRemoteConfig()
	viper.OnConfigChange(func(in fsnotify.Event) {
		newLevel := strings.ToLower(viper.GetString("logger.level"))
		if strings.ToLower(level.String()) == newLevel {
			return
		}
		Info("log level变更", zap.String("newLevel", newLevel), zap.String("oldLevel", level.String()))
		var logLevel = zap.InfoLevel
		switch newLevel {
		case "debug":
			logLevel = zap.DebugLevel
		case "info":
			logLevel = zap.InfoLevel
		case "error":
			logLevel = zap.ErrorLevel
		case "dPanic":
			logLevel = zap.DPanicLevel
		case "panic":
			logLevel = zap.PanicLevel
		case "fatal":
			logLevel = zap.FatalLevel
		default:
			logLevel = zap.InfoLevel
		}
		level.SetLevel(logLevel)
	})
}

func newEncodeConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder, // 小写编码器
		EncodeTime:     zapcore.ISO8601TimeEncoder,    // ISO8601 UTC 时间格式
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder, // 径编码器
	}
}

func Debug(msg string, fields ...zap.Field) {
	sendLog(zap.DebugLevel, msg, fields...)
}

func Info(msg string, fields ...zap.Field) {
	sendLog(zap.InfoLevel, msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	sendLog(zap.WarnLevel, msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	sendLog(zap.ErrorLevel, msg, fields...)
}

func Panic(msg string, fields ...zap.Field) {
	sendLog(zap.PanicLevel, msg, fields...)
}

func DPanic(msg string, fields ...zap.Field) {
	sendLog(zap.DPanicLevel, msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	sendLog(zap.FatalLevel, msg, fields...)
}

func sendLog(level zapcore.Level, msg string, fields ...zap.Field) {
	if ce := internalLogger.Check(level, msg); ce != nil {
		ce.Write(fields...)
	}
}
