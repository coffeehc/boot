package etcd_rc

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"git.xiagaogao.com/coffee/base/errors"
	"git.xiagaogao.com/coffee/base/log"
	"git.xiagaogao.com/coffee/boot/component/etcd"
	"git.xiagaogao.com/coffee/boot/configuration"
	"git.xiagaogao.com/coffee/boot/plugin/manage"
	"git.xiagaogao.com/coffee/boot/plugin/rpc"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/concurrency"
	"go.uber.org/zap"
)

type serviceImpl struct {
	client *clientv3.Client
}

func newServiceImpl() *serviceImpl {
	client := etcd.GetEtcdClient()
	return &serviceImpl{
		client: client,
	}
}

func (impl *serviceImpl) GetClient() *clientv3.Client {
	return impl.client
}

func buildServiceKeyPrefix() string {
	return fmt.Sprintf("/ms/registers/%s/%s/", configuration.GetServiceName(), configuration.GetRunModel())
}

func (impl *serviceImpl) Register(ctx context.Context, serviceInfo configuration.ServiceInfo) errors.Error {
	scope := zap.String("scope", "etcd.register")
	if ctx.Err() != nil {
		return errors.MessageError("服务注册已经关闭")
	}
	serviceKey := fmt.Sprintf("%s%s/%s", buildServiceKeyPrefix(), serviceInfo.Version, rpc.GetService().GetRPCServerAddr())
	registerInfo := &configuration.ServiceRegisterInfo{
		Info:           serviceInfo,
		ServiceAddr:    rpc.GetService().GetRPCServerAddr(),
		ManageEndpoint: manage.GetManageEndpoint(),
		Metadata:       serviceInfo.Metadata,
	}
	value, err := json.Marshal(registerInfo)
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
				session, err := concurrency.NewSession(impl.client, concurrency.WithTTL(5))
				if err != nil {
					log.Error("创建ETCD Session异常", zap.Error(err), scope)
					goto SLEEP
				}
				_, err = impl.client.Put(ctx, serviceKey, string(value), clientv3.WithLease(session.Lease()))
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
