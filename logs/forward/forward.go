package forward

import "git.xiagaogao.com/coffee/boot/errors"

type Service interface {
	Start()
}

func NewService() (Service, errors.Error) {

}
