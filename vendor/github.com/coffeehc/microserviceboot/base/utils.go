package base

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"

	"context"

	"github.com/coffeehc/logger"
	xcontext "golang.org/x/net/context"
	"gopkg.in/yaml.v2"
)

const envServiceMode = "SERVICE_MODE"

func GetServiceMode() string {
	if mode, ok := os.LookupEnv("envServiceMode"); ok {
		return mode
	}
	return "dev"
}

const errScopeLoadConfig = "loadConfig"

//LoadConfig load the config from config path
func LoadConfig(configPath string, config interface{}) Error {
	logger.Debug("load config file %s", configPath)
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return NewError(Error_System, errScopeLoadConfig, err.Error())
	}
	err = yaml.Unmarshal(data, config)
	if err != nil {
		return NewError(Error_System, errScopeLoadConfig, err.Error())
	}
	return nil
}

const envIPInterfaceName = "NET_INTERFACE"

func GetLocalIP() (string, Error) {
	if interfaceName, ok := os.LookupEnv(envIPInterfaceName); ok {
		netInterface, err := net.InterfaceByName(interfaceName)
		if err != nil {
			return "", NewError(Error_System, "serviceboot", fmt.Sprintf("获取指定网络接口[s%]失败", interfaceName))
		}
		addrs, err := netInterface.Addrs()
		if err != nil || len(addrs) == 0 {
			return "", NewError(Error_System, "serviceboot", fmt.Sprintf("获取指定网络接口[s%]地址失败", interfaceName))
		}
		return getActiveIP(addrs)
	}
	addrs, err := net.InterfaceAddrs()
	if err != nil || len(addrs) == 0 {
		return "", NewError(Error_System, "serviceboot", "获取本地Ip地址失败")
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
	return "", NewError(Error_System, "serviceboot", "没有可用的有效 Ip")
}

func ContextToXContext(cxt context.Context) xcontext.Context {
	return cxt.(xcontext.Context)
}

func XContextToContext(cxt xcontext.Context) context.Context {
	return cxt.(context.Context)
}

func WarpServerAddr(serviceAddr string) (string, Error) {
	if serviceAddr == "" {
		return "", NewError(Error_System, "base", "服务地址不能为空")
	}
	addr, err := net.ResolveTCPAddr("tcp", serviceAddr)
	if err != nil {
		return "", NewError(Error_System, "base", fmt.Sprintf("服务地址不是一个标准的tcp地址:%s", err))
	}
	serverAddr := serviceAddr
	if addr.IP.Equal(net.IPv4zero) {
		localIp, err := GetLocalIP()
		if err != nil {
			return "", NewErrorWrapper(Error_System, "base", err)
		}
		serverAddr = fmt.Sprintf("%s:%d", localIp, addr.Port)
	}
	return serverAddr, nil
}
