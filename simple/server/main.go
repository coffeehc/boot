package main

import (
	"context"

	"git.xiagaogao.com/coffee/boot/serviceboot"
)

func main() {
	serviceboot.ServiceLaunch("simple_service", context.Background(), &ServiceImpl{})
}
