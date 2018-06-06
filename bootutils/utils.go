package bootutils

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"os"

	"git.xiagaogao.com/coffee/boot/errors"
	"gopkg.in/yaml.v2"
)

const envIPInterfaceName = "NET_INTERFACE"

func GetLocalIP() (string, errors.Error) {
	if interfaceName, ok := os.LookupEnv(envIPInterfaceName); ok {
		netInterface, err := net.InterfaceByName(interfaceName)
		if err != nil {
			return "", errors.NewError(errors.Error_System, "serviceboot", fmt.Sprintf("获取指定网络接口[s%]失败", interfaceName))
		}
		addrs, err := netInterface.Addrs()
		if err != nil || len(addrs) == 0 {
			return "", errors.NewError(errors.Error_System, "serviceboot", fmt.Sprintf("获取指定网络接口[s%]地址失败", interfaceName))
		}
		return getActiveIP(addrs)
	}
	addrs, err := net.InterfaceAddrs()
	if err != nil || len(addrs) == 0 {
		return "", errors.NewError(errors.Error_System, "serviceboot", "获取本地Ip地址失败")
	}
	return getActiveIP(addrs)
}

func getActiveIP(addrs []net.Addr) (string, errors.Error) {
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}
	return "", errors.NewError(errors.Error_System, "serviceboot", "没有可用的有效 Ip")
}

func WarpServerAddr(serviceAddr string) (string, errors.Error) {
	if serviceAddr == "" {
		return "", errors.NewError(errors.Error_System, "base", "服务地址不能为空")
	}
	addr, err := net.ResolveTCPAddr("tcp", serviceAddr)
	if err != nil {
		return "", errors.NewError(errors.Error_System, "base", fmt.Sprintf("服务地址不是一个标准的tcp地址:%s", err))
	}
	serverAddr := serviceAddr
	if addr.IP.Equal(net.IPv4zero) {
		localIp, err := GetLocalIP()
		if err != nil {
			return "", errors.NewErrorWrapper(errors.Error_System, "base", err)
		}
		serverAddr = fmt.Sprintf("%s:%d", localIp, addr.Port)
	}
	return serverAddr, nil
}

const errScopeLoadConfig = "loadConfig"

//LoadConfig load the config from config path
func LoadConfig(ctx context.Context, configPath string, config interface{}) errors.Error {
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return errors.NewError(errors.Error_System, errScopeLoadConfig, err.Error())
	}
	err = yaml.Unmarshal(data, config)
	if err != nil {
		return errors.NewError(errors.Error_System, errScopeLoadConfig, err.Error())
	}
	return nil
}
