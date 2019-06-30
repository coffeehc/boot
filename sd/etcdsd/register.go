package etcdsd

import (
	"context"
	"fmt"
	"time"

	"git.xiagaogao.com/coffee/boot"
	"git.xiagaogao.com/coffee/boot/errors"
	"git.xiagaogao.com/coffee/boot/sd"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/json-iterator/go"
	"go.uber.org/zap"
)

func RegisterService(ctx context.Context, client *clientv3.Client, info *boot.ServiceInfo, serviceAddr string, manageEndpoint string, data map[string]string, errorService errors.Service, logger *zap.Logger) errors.Error {
	errorService = errorService.NewService("sd")
	if ctx.Err() != nil {
		return errorService.MessageError("服务注册已经关闭")
	}
	serviceKey := fmt.Sprintf("%s%s", sd.BuildServiceKeyPrefix(info), serviceAddr)
	registerInfo := &sd.ServiceRegisterInfo{
		ServiceInfo:    info,
		ServerAddr:     serviceAddr,
		ManageEndpoint: manageEndpoint,
		Metadata:       data,
	}
	value, err := jsoniter.Marshal(registerInfo)
	if err != nil {
		return errorService.WrappedSystemError(err)
	}
	go func() {
		for {
			func() {
				defer func() {
					if err := recover(); err != nil {
						_err := errors.ConverUnkonwError(err, errorService)
						logger.Error("注册模块异常", _err.GetFieldsWithCause()...)
					}
				}()
				session, err := concurrency.NewSession(client, concurrency.WithTTL(5))
				if err != nil {
					logger.Error("创建ETCD Session异常", zap.Error(err))
					goto SLEEP
				}
				_, err = client.Put(ctx, serviceKey, string(value), clientv3.WithLease(session.Lease()))
				if err != nil {
					logger.Error("设置注册信息KV失败", zap.Error(err))
					goto SLEEP
				}
				logger.Debug("注册服务成功", zap.String("serviceKey", serviceKey))
				<-session.Done()
			SLEEP:
				time.Sleep(time.Second)
			}()
		}
	}()
	return nil
}
