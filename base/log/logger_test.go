package log

import (
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestLogInit(t *testing.T) {
	SetBaseFields(zap.String("serviceName", "test"))
	logger := GetLogger()
	logger.Error("hahah", zap.Duration("d1", time.Duration(123987129386817263)))
	Debug("test1")
	logger.Sync()
}
