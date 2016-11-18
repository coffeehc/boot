package grpcbase

import (
	"fmt"
	"github.com/coffeehc/logger"
)

type GrpcLogger struct {
}

func (this GrpcLogger) Fatal(args ...interface{}) {
	logger.Printf(logger.LOGGER_LEVEL_ERROR, 5, fmt.Sprint(args))
}
func (this GrpcLogger) Fatalf(format string, args ...interface{}) {
	logger.Printf(logger.LOGGER_LEVEL_ERROR, 5, fmt.Sprintf(format, args))
}
func (this GrpcLogger) Fatalln(args ...interface{}) {
	logger.Printf(logger.LOGGER_LEVEL_ERROR, 5, fmt.Sprintln(args))
}
func (this GrpcLogger) Print(args ...interface{}) {
	logger.Printf(logger.LOGGER_LEVEL_DEBUG, 5, fmt.Sprint(args))
}
func (this GrpcLogger) Printf(format string, args ...interface{}) {
	logger.Printf(logger.LOGGER_LEVEL_DEBUG, 5, fmt.Sprintf(format, args))
}
func (this GrpcLogger) Println(args ...interface{}) {
	logger.Printf(logger.LOGGER_LEVEL_DEBUG, 5, fmt.Sprintln(args))
}
