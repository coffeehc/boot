package manage

import (
	"github.com/spf13/pflag"
)

var manage_endpoint = pflag.String("env_manage_endpoint", "0.0.0.0:7777", "管理服务地址")
