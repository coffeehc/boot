package bootutils

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"os"

	"git.xiagaogao.com/coffee/boot/errors"
	"git.xiagaogao.com/coffee/boot/logs"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

const envIPInterfaceName = "NET_INTERFACE"

func GetLocalIP(errorService errors.Service) (string, errors.Error) {
	ifs, _ := net.Interfaces()
	for _, iface := range ifs {
		addr, _ := iface.Addrs()
		fmt.Printf("%s-->%#q", iface.Name, addr)
	}
	if interfaceName, ok := os.LookupEnv(envIPInterfaceName); ok {
		netInterface, err := net.InterfaceByName(interfaceName)
		if err != nil {
			return "", errorService.SystemError(fmt.Sprintf("获取指定网络接口[s%]失败", interfaceName))
		}
		addrs, err := netInterface.Addrs()
		if err != nil || len(addrs) == 0 {
			return "", errorService.SystemError(fmt.Sprintf("获取指定网络接口[s%]地址失败", interfaceName))
		}
		return getActiveIP(addrs, errorService)
	}
	addrs, err := net.InterfaceAddrs()
	if err != nil || len(addrs) == 0 {
		return "", errorService.SystemError("获取本地Ip地址失败")
	}
	return getActiveIP(addrs, errorService)
}

func getActiveIP(addrs []net.Addr, errorService errors.Service) (string, errors.Error) {
	for _, addr := range addrs {
		fmt.Printf("地址为:%s\n", addr)
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}
	return "", errorService.SystemError("没有可用的有效 Ip")
}

func WarpServerAddr(serviceAddr string, errorService errors.Service) (string, errors.Error) {
	if serviceAddr == "" {
		return "", errorService.SystemError("服务地址不能为空")
	}
	addr, err := net.ResolveTCPAddr("tcp", serviceAddr)
	if err != nil {
		return "", errorService.SystemError(fmt.Sprintf("服务地址不是一个标准的tcp地址:%s", err))
	}
	serverAddr := serviceAddr
	if addr.IP.Equal(net.IPv4zero) {
		localIp, err := GetLocalIP(errorService)
		if err != nil {
			return "", errorService.WappedSystemError(err)
		}
		serverAddr = fmt.Sprintf("%s:%d", localIp, addr.Port)
	}
	return serverAddr, nil
}

//LoadConfig load the config from config path
func LoadConfig(ctx context.Context, configPath string, config interface{}, errorService errors.Service, logger *zap.Logger) errors.Error {
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return errorService.WappedSystemError(err, logs.F_ExtendData(configPath))
	}
	err = yaml.Unmarshal(data, config)
	if err != nil {
		return errorService.WappedSystemError(err, logs.F_ExtendData(configPath))
	}
	return nil
}
