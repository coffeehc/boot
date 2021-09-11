package internal

import "github.com/coffeehc/base/errors"

type serviceImpl struct {
	registerCenter RegisterCenter
}

func (impl *serviceImpl) GetRegisterCenter() RegisterCenter {
	return impl.registerCenter
}

func (impl *serviceImpl) SetRegisterCenter(center RegisterCenter) errors.Error {
	if impl.registerCenter != nil {
		//	其实这个理还可以改造一下，弄多注册中心
		return errors.SystemError("已经存在注册中心，不能重复注册")
	}
	impl.registerCenter = center
	return nil
}
