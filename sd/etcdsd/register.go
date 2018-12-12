package etcdsd

import (
	"context"
	"fmt"
	"os"
	"time"

	"git.xiagaogao.com/coffee/boot"
	"git.xiagaogao.com/coffee/boot/errors"
	"git.xiagaogao.com/coffee/boot/logs"
	"git.xiagaogao.com/coffee/boot/sd"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/pquerna/ffjson/ffjson"
	"go.uber.org/zap"
)

func RegisterService(ctx context.Context, client *clientv3.Client, info *boot.ServiceInfo, serviceAddr string, manageEndpoint string, data string, errorService errors.Service, logger *zap.Logger) errors.Error {
	errorService = errorService.NewService("sd")
	if ctx.Err() != nil {
		return errorService.MessageError("服务注册已经关闭")
	}
	serviceKey := fmt.Sprintf("%s%s", sd.BuildServiceKeyPrefix(info), serviceAddr)
	registerInfo := &sd.ServiceRegisterInfo{
		ServiceInfo:    info,
		ServerAddr:     serviceAddr,
		ManageEndpoint: manageEndpoint,
		Data:           data,
	}
	value, err := ffjson.Marshal(registerInfo)
	if err != nil {
		return errorService.WrappedSystemError(err)
	}
	go func() {
		for {
			func() {
				defer func() {
					recover()
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
				<-session.Done()
			SLEEP:
				time.Sleep(time.Second)
			}()
		}
	}()
	return nil
}

func RegisterService_back(ctx context.Context, client *clientv3.Client, info *boot.ServiceInfo, serviceAddr string, errorService errors.Service, logger *zap.Logger) errors.Error {
	regServiceEndpoint, ok := os.LookupEnv("ENV_REG_SERVICE_ENDPOINT") //此处用于自定义一个注册服务，用于在前端有代理的情况下使用
	if !ok {
		regServiceEndpoint = serviceAddr
	}
	errorService = errorService.NewService("sd")
	if ctx.Err() != nil {
		return errorService.MessageError("服务注册已经关闭")
	}
	ttl := int64(5)
	lease := clientv3.NewLease(client)
	resp, err := lease.Grant(ctx, ttl)
	if err != nil {
		return errorService.WrappedSystemError(err)
	}
	serviceKey := fmt.Sprintf("%s%s", sd.BuildServiceKeyPrefix(info), regServiceEndpoint)
	value, err := ffjson.Marshal(&sd.ServiceRegisterInfo{ServiceInfo: info, ServerAddr: regServiceEndpoint})
	if err != nil {
		return errorService.WrappedSystemError(err)
	}
	_, err = client.Put(ctx, serviceKey, string(value), clientv3.WithLease(resp.ID))
	if err != nil {
		return errorService.WrappedSystemError(err)
	}
	ch, err := lease.KeepAlive(ctx, resp.ID)
	if err != nil {
		return errorService.WrappedSystemError(err)
	}
	go func() {
		for {
			resp := <-ch
			if resp == nil {
				lease.Close()
				break
			}
		}
		time.Sleep(time.Second)
		err := RegisterService_back(ctx, client, info, regServiceEndpoint, errorService, logger)
		if err != nil {
			logger.Error("注册服务发生了错误", logs.F_Error(err))
		}
	}()
	logger.Debug("向Etcd注册服务成功", logs.F_ExtendData(serviceKey))
	return nil
}
