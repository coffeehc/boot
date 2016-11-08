package grpcboot

import (
	"fmt"
	"github.com/coffeehc/logger"
)

type _logger struct {
}

func (this *_logger) Fatal(args ...interface{}) {
	logger.Printf(logger.LOGGER_LEVEL_ERROR, 4, fmt.Sprint(args))
}
func (this *_logger) Fatalf(format string, args ...interface{}) {
	logger.Printf(logger.LOGGER_LEVEL_ERROR, 4, fmt.Sprintf(format, args))
}
func (this *_logger) Fatalln(args ...interface{}) {
	logger.Printf(logger.LOGGER_LEVEL_ERROR, 4, fmt.Sprintln(args))
}
func (this *_logger) Print(args ...interface{}) {
	logger.Printf(logger.LOGGER_LEVEL_DEBUG, 4, fmt.Sprint(args))
}
func (this *_logger) Printf(format string, args ...interface{}) {
	logger.Printf(logger.LOGGER_LEVEL_DEBUG, 4, fmt.Sprintf(format, args))
}
func (this *_logger) Println(args ...interface{}) {
	logger.Printf(logger.LOGGER_LEVEL_DEBUG, 4, fmt.Sprintln(args))
}
