package etcd

import (
	"context"
	"testing"
	"time"

	"git.xiagaogao.com/coffee/base/log"
	"git.xiagaogao.com/coffee/boot/configuration"
	"git.xiagaogao.com/coffee/boot/plugin"
	"git.xiagaogao.com/coffee/boot/testutils"
	"go.uber.org/zap"
	"gopkg.in/check.v1"
)

var _ = check.Suite(&suite{})

func Test(t *testing.T) { check.TestingT(t) }

type suite struct {
	dir       string // 测试用的临时目录
	f         string // 测试用的临时文件
	ctx       context.Context
	cancelFun context.CancelFunc
}

func (s *suite) TearDownSuite(c *check.C) {
	defer log.GetLogger().Sync()
}

func (s *suite) SetUpSuite(c *check.C) {
	testutils.InitTestConfig()
	s.ctx, s.cancelFun = context.WithTimeout(context.TODO(), time.Second*30)
	configuration.DisableRemoteConfig()
	configuration.InitConfiguration(context.TODO(), configuration.ServiceInfo{
		ServiceName: "test",
	})
	EnablePlugin(context.TODO())
	plugin.StartPlugins(s.ctx)
}

func (impl *suite) TestGetDir(c *check.C) {
	service := GetService()
	list, err := service.GetVersion()
	if err != nil {
		log.Error("错误", err.GetFieldsWithCause()...)
		c.FailNow()
		return
	}
	for _, memeber := range list {
		log.Debug("member信息", zap.Any("member", memeber))
	}
}
