package rpc

import (
	"fmt"
	"net"
	"os"

	"github.com/coffeehc/base/errors"
	"github.com/coffeehc/base/log"
	"go.uber.org/zap"
)

const envIPInterfaceName = "NET_INTERFACE"

func GetLocalIP() (net.IP, error) {
	if interfaceName, ok := os.LookupEnv(envIPInterfaceName); ok && interfaceName != "" {
		netInterface, err := net.InterfaceByName(interfaceName)
		if err != nil {
			log.Error("获取指定网络接口失败", zap.String("interfaceName", interfaceName))
			return net.IPv4zero, errors.SystemError("获取指定网络接口失败")
		}
		addrs, err := netInterface.Addrs()
		if err != nil || len(addrs) == 0 {
			log.Error("获取指定网络接口地址失败", zap.String("interfaceName", interfaceName))
			return net.IPv4zero, errors.SystemError("获取指定网络接口地址失败")
		}
		return getActiveIP(addrs)
	}
	addrs, err := net.InterfaceAddrs()
	if err != nil || len(addrs) == 0 {
		log.Error("获取本地Ip失败")
		return net.IPv4zero, errors.SystemError("获取本地Ip地址失败")
	}
	return getActiveIP(addrs)
}

func getActiveIP(addrs []net.Addr) (net.IP, error) {
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP, nil
			}
		}
	}
	return net.IPv4zero, errors.SystemError("没有可用的有效 Ip")
}

func WarpServiceAddr(serviceAddr string) (string, error) {
	if serviceAddr == "" {
		return "", errors.SystemError("服务地址不能为空")
	}
	addr, err := net.ResolveTCPAddr("tcp4", serviceAddr)
	if err != nil {
		return "", errors.SystemError(fmt.Sprintf("服务地址不是一个标准的tcp地址:%s", err))
	}
	if addr.IP.Equal(net.IPv4zero) {
		localIp, err := GetLocalIP()
		if err != nil {
			return "", err
		}
		addr.IP = localIp
	}

	return addr.String(), nil
}
