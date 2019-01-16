package boot

import "github.com/spf13/pflag"

var runModel = pflag.String("run_model", "", "运行模式,必填（dev，test，product或其他）")
