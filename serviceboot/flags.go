package serviceboot

import "github.com/spf13/pflag"

const (
	Flag_configPath          = "config"
	Flag_singleService       = "single_service"
	Flag_serviceEndpointFlag = "env_service_endpoint"
)

var configPath = pflag.String(Flag_configPath, "./config.yml", "配置文件路径")
var singleService = pflag.Bool(Flag_singleService, false, "是否是单体服务")
var serviceEndpointFlag = pflag.String(Flag_serviceEndpointFlag, "0.0.0.0:8888", "服务地址")
