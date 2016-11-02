package main

type ServiceInfo struct {
}

func (this *ServiceInfo) GetApiDefine() string {
	return ""
}

//获取 Service 名称
func (this *ServiceInfo) GetServiceName() string {
	return ""
}

//获取服务版本号
func (this *ServiceInfo) GetVersion() string {
	return ""
}

//获取服务描述
func (this *ServiceInfo) GetDescriptor() string {
	return ""
}

//获取 Service tags
func (this *ServiceInfo) GetServiceTags() []string {
	return nil
}
