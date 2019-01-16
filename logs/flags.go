package logs

import "github.com/spf13/pflag"

var logLevel = pflag.String("logger_level", "", "日志级别(debug,warn,info,error)")
