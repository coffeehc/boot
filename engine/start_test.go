package engine_test

import (
	"context"
	"testing"
	"time"

	"git.xiagaogao.com/coffee/boot/base/log"
	"git.xiagaogao.com/coffee/boot/component/etcdsd"
	"git.xiagaogao.com/coffee/boot/configuration"
	"git.xiagaogao.com/coffee/boot/engine"
	"git.xiagaogao.com/coffee/boot/plugin/pluginutils"
	"git.xiagaogao.com/coffee/boot/plugin/rpc"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func TestStartEngine(t *testing.T) {
	viper.Set("grpc", rpc.RpcConfig{
		RPCServerAddr: "0.0.0.0:0",
	})
	viper.Set("etcd", etcdsd.Config{
		Endpoints: []string{"192.168.3.2:2379"},
	})
	configuration.SetModel("dev")
	ctx, _ := context.WithTimeout(context.TODO(), time.Second*30)
	engine.StartEngine(ctx, configuration.ServiceInfo{
		ServiceName: "test",
	}, pluginutils.EnableMicorPlugin, func(ctx context.Context, cmd *cobra.Command, args []string) error {
		log.Debug("哈哈哈")

		return nil
	})
}
