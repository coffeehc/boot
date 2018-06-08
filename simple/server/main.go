package main

import (
	"context"

	"git.xiagaogao.com/coffee/boot/serviceboot"
)

func main() {
	serviceboot.ServiceLaunch(context.Background(), &ServiceImpl{})
}
