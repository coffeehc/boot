package common

type Loadbalance interface {
	//获取 balance信息
	AcceptNewBalanceInfo()
	//获取 Service的地址
	GetServiceAddr(serviceName string, tag string)
}
