package common

type Service interface {
	Run() error
	Stop() error
	GetServiceInfo() ServiceInfo
	GetEndPoints() []*EndPoint
}
