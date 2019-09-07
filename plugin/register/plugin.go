package register

import (
	"context"
	"fmt"
	"sync"
	"time"

	"git.xiagaogao.com/coffee/boot/base/errors"
	"git.xiagaogao.com/coffee/boot/base/log"
	"git.xiagaogao.com/coffee/boot/component/etcdsd"
	"git.xiagaogao.com/coffee/boot/configuration"
	"git.xiagaogao.com/coffee/boot/plugin"
	"git.xiagaogao.com/coffee/boot/plugin/manage"
	"git.xiagaogao.com/coffee/boot/plugin/rpc"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/json-iterator/go"
	"go.uber.org/zap"
)

type pluginImpl struct {
}

func (impl *pluginImpl) Start(ctx context.Context) errors.Error {
	registerService(ctx, etcdsd.GetEtcdClient(), nil)
	return nil
}
func (impl *pluginImpl) Stop(ctx context.Context) errors.Error {
	return nil
}

var registered = false
var mutex = new(sync.Mutex)

func EnablePlugin(ctx context.Context) {
	mutex.Lock()
	defer mutex.Unlock()
	if registered {
		return
	}
	etcdsd.InitEtcdClient()
	rpc.EnablePlugin(ctx)
	manage.EnablePlugin(ctx)
	plugin.RegisterPlugin("serviceRegister", &pluginImpl{})
}

func registerService(ctx context.Context, client *clientv3.Client, metadata map[string]string) errors.Error {
	scope := zap.String("scope", "etcd.register")
	if ctx.Err() != nil {
		return errors.MessageError("服务注册已经关闭")
	}
	serviceInfo := configuration.GetServiceInfo()
	serviceKey := fmt.Sprintf("%s%s/%s", etcdsd.BuildServiceKeyPrefix(), serviceInfo.Version, rpc.GetRPCServerAddr())
	registerInfo := &configuration.ServiceRegisterInfo{
		Info:           serviceInfo,
		ServiceAddr:    rpc.GetRPCServerAddr(),
		ManageEndpoint: manage.GetManageEndpoint(),
		Metadata:       metadata,
	}
	value, err := jsoniter.Marshal(registerInfo)
	if err != nil {
		return errors.WrappedSystemError(err)
	}
	go func() {
		for {
			if ctx.Err() != nil {
				return
			}
			func() {
				defer func() {
					if err := recover(); err != nil {
						_err := errors.ConverUnknowError(err)
						log.Error("服务注册异常", _err.GetFieldsWithCause(scope)...)
					}
				}()
				session, err := concurrency.NewSession(client, concurrency.WithTTL(5))
				if err != nil {
					log.Error("创建ETCD Session异常", zap.Error(err), scope)
					goto SLEEP
				}
				_, err = client.Put(ctx, serviceKey, string(value), clientv3.WithLease(session.Lease()))
				if err != nil {
					log.Error("设置注册信息KV失败", zap.Error(err))
					goto SLEEP
				}
				log.Info("注册服务成功", zap.String("serviceKey", serviceKey))
				<-session.Done()
			SLEEP:
				if ctx.Err() != nil {
					return
				}
				time.Sleep(time.Second)
			}()
		}
	}()
	return nil
}
