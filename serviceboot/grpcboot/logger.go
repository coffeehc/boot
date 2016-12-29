package grpcboot

import (
	"fmt"

	"github.com/coffeehc/logger"
)

type grpcLogger struct {
}

func (grpcLogger) Fatal(args ...interface{}) {
	logger.Printf(logger.LevelError, 5, fmt.Sprint(args))
}
func (grpcLogger) Fatalf(format string, args ...interface{}) {
	logger.Printf(logger.LevelError, 5, fmt.Sprintf(format, args))
}
func (grpcLogger) Fatalln(args ...interface{}) {
	logger.Printf(logger.LevelError, 5, fmt.Sprintln(args))
}
func (grpcLogger) Print(args ...interface{}) {
	logger.Printf(logger.LevelDebug, 5, fmt.Sprint(args))
}
func (grpcLogger) Printf(format string, args ...interface{}) {
	logger.Printf(logger.LevelDebug, 5, fmt.Sprintf(format, args))
}
func (grpcLogger) Println(args ...interface{}) {
	logger.Printf(logger.LevelDebug, 5, fmt.Sprintln(args))
}
