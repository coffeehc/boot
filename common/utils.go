package common

import (
	"net"

	"github.com/coffeehc/logger"
)

var (
	localIp = net.IPv4(127, 0, 0, 1)
)

func GetLocalIp() net.IP {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		logger.Error("无法获取网络接口信息,%s", err)
		return localIp
	}

	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP
			}
		}
	}
	return localIp
}
