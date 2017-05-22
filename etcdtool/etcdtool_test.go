package etcdtool_test

import (
	"time"

	"context"

	"testing"

	"github.com/coffeehc/logger"
	"github.com/coffeehc/microserviceboot/base"
	"github.com/coffeehc/microserviceboot/etcdtool"
	"github.com/coreos/etcd/clientv3"
	. "gopkg.in/check.v1"
)

type EtcdToolSuite struct {
	etcdClient  *clientv3.Client
	serviceInfo base.ServiceInfo
}

var _ = Suite(&EtcdToolSuite{})

func Test(t *testing.T) {
	TestingT(t)
}

func (t *EtcdToolSuite) SetUpSuite(c *C) {
	logger.SetDefaultLevel("/", logger.LevelDebug)
	t.serviceInfo = base.NewSimpleServiceInfo("testService", "0.0.1", "dev", "https", "测试项目", "")
	config := &etcdtool.Config{
		Endpoints:   []string{"127.0.0.1:2379"},
		DialTimeout: 5,
		Username:    "service",
		Password:    "service#123",
	}
	client, err := etcdtool.NewClient(config)
	if err != nil {
		c.Assert(err, IsNil)
		return
	}
	t.etcdClient = client
}

func (t *EtcdToolSuite) SetUpTest(c *C) {

}

func (t *EtcdToolSuite) TearDownTest(c *C) {

}
func (t *EtcdToolSuite) TearDownSuite(c *C) {
	time.Sleep(time.Second)

}
func (t *EtcdToolSuite) TestRegister(c *C) {
	cxt := context.Background()
	cxt, _ = context.WithTimeout(cxt, time.Second*10)
	register, err := etcdtool.NewEtcdServiceRegister(t.etcdClient)
	c.Assert(err, IsNil)
	//internal_ms.testService.dev.127.0.0.1:8080
	deregister, err := register.RegService(cxt, t.serviceInfo, "127.0.0.1:8080")
	c.Assert(err, IsNil)
	c.Assert(deregister, NotNil)
	//TODO test balancer
	//balancer, err := etcdtool.NewEtcdBalancer(cxt, t.etcdClient, t.serviceInfo)
	//c.Assert(err, IsNil)
	//c.Assert(balancer, NotNil)
	//_err := balancer.Start("", loadbalancer.BalancerConfig{})
	//c.Assert(_err, IsNil)
	//addrsChan := balancer.Notify()
	//addrs := <-addrsChan
	//c.Assert(len(addrs), Equals, 1)
	//c.Assert(addrs[0].Addr, Equals, "127.0.0.1:8080")
	deregister()
}
