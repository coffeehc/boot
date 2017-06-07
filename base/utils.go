package base

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"

	"github.com/coffeehc/logger"
	"gopkg.in/yaml.v2"
)

const errScopeLoadConfig = "loadConfig"

//LoadConfig load the config from config path
func LoadConfig(configPath string, config interface{}) Error {
	logger.Debug("load config file %s", configPath)
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return NewError(ErrCodeBaseSystemConfig, errScopeLoadConfig, err.Error())
	}
	err = yaml.Unmarshal(data, config)
	if err != nil {
		return NewError(ErrCodeBaseSystemConfig, errScopeLoadConfig, err.Error())
	}
	return nil
}

const envIPInterfaceName = "NET_INTERFACE"

func GetLocalIP() (string, Error) {
	if interfaceName, ok := os.LookupEnv(envIPInterfaceName); ok {
		netInterface, err := net.InterfaceByName(interfaceName)
		if err != nil {
			return "", NewError(ErrCodeBaseSystemInit, "serviceboot", fmt.Sprintf("获取指定网络接口[s%]失败", interfaceName))
		}
		addrs, err := netInterface.Addrs()
		if err != nil || len(addrs) == 0 {
			return "", NewError(ErrCodeBaseSystemInit, "serviceboot", fmt.Sprintf("获取指定网络接口[s%]地址失败", interfaceName))
		}
		return getActiveIP(addrs)
	}
	addrs, err := net.InterfaceAddrs()
	if err != nil || len(addrs) == 0 {
		return "", NewError(ErrCodeBaseSystemInit, "serviceboot", "获取本地Ip地址失败")
	}
	return getActiveIP(addrs)
}

func getActiveIP(addrs []net.Addr) (string, Error) {
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}
	return "", NewError(ErrCodeBaseSystemInit, "serviceboot", "没有可用的有效 Ip")
}
