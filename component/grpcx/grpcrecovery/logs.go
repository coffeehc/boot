package grpcrecovery

import (
	"github.com/coffeehc/base/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc/grpclog"
)

var DisableGrpcLog = false

var level = zap.NewAtomicLevelAt(zap.ErrorLevel)

func SetLogLevel(l zapcore.Level) {
	level.SetLevel(l)
}

type zapLogger struct {
	logger *zap.SugaredLogger
}

// 创建封装了zap的对象，该对象是对LoggerV2接口的实现
func NewZapLogger() grpclog.LoggerV2 {
	return &zapLogger{
		logger: log.GetLogger().WithOptions(zap.AddCallerSkip(2)).Sugar(),
	}
}
func (zl *zapLogger) Info(args ...interface{}) {
	if !DisableGrpcLog && level.Enabled(zap.InfoLevel) {
		zl.logger.Info(args...)
	}
}

func (zl *zapLogger) Infoln(args ...interface{}) {
	if !DisableGrpcLog && level.Enabled(zap.InfoLevel) {
		zl.logger.Info(args...)
	}
}
func (zl *zapLogger) Infof(format string, args ...interface{}) {
	if !DisableGrpcLog && level.Enabled(zap.InfoLevel) {
		zl.logger.Infof(format, args...)
	}
}

func (zl *zapLogger) Warning(args ...interface{}) {
	if !DisableGrpcLog && level.Enabled(zap.WarnLevel) {
		zl.logger.Warn(args...)
	}
}

func (zl *zapLogger) Warningln(args ...interface{}) {
	if !DisableGrpcLog && level.Enabled(zap.WarnLevel) {
		zl.logger.Warn(args...)
	}
}

func (zl *zapLogger) Warningf(format string, args ...interface{}) {
	if !DisableGrpcLog && level.Enabled(zap.WarnLevel) {
		zl.logger.Warnf(format, args...)
	}
}

func (zl *zapLogger) Error(args ...interface{}) {
	if !DisableGrpcLog && level.Enabled(zap.ErrorLevel) {
		zl.logger.Error(args...)
	}
}

func (zl *zapLogger) Errorln(args ...interface{}) {
	if !DisableGrpcLog && level.Enabled(zap.ErrorLevel) {
		zl.logger.Error(args...)
	}
}

func (zl *zapLogger) Errorf(format string, args ...interface{}) {
	if !DisableGrpcLog && level.Enabled(zap.ErrorLevel) {
		zl.logger.Errorf(format, args...)
	}
}

func (zl *zapLogger) Fatal(args ...interface{}) {
	if !DisableGrpcLog && level.Enabled(zap.ErrorLevel) {
		zl.logger.Fatal(args...)
	}
}

func (zl *zapLogger) Fatalln(args ...interface{}) {
	if !DisableGrpcLog && level.Enabled(zap.FatalLevel) {
		zl.logger.Fatal(args...)
	}
}

// Fatalf logs to fatal level
func (zl *zapLogger) Fatalf(format string, args ...interface{}) {
	if !DisableGrpcLog && level.Enabled(zap.FatalLevel) {
		zl.logger.Fatalf(format, args...)
	}
}

// V reports whether verbosity level l is at least the requested verbose level.
func (zl *zapLogger) V(v int) bool {
	return true
}
