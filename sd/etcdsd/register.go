package etcdsd

import (
	"context"
	"fmt"
	"time"

	"git.xiagaogao.com/coffee/boot"
	"git.xiagaogao.com/coffee/boot/errors"
	"git.xiagaogao.com/coffee/boot/logs"
	"git.xiagaogao.com/coffee/boot/sd"
	"github.com/coreos/etcd/clientv3"
	"github.com/pquerna/ffjson/ffjson"
	"go.uber.org/zap"
)

func RegisterService(ctx context.Context, client *clientv3.Client, info boot.ServiceInfo, serviceAddr string, errorService errors.Service, logger *zap.Logger) errors.Error {
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
	serviceKey := fmt.Sprintf("%s%s", sd.BuildServiceKeyPrefix(info), serviceAddr)
	value, err := ffjson.Marshal(&sd.ServiceRegisterInfo{ServiceInfo: info, ServerAddr: serviceAddr})
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
		err := RegisterService(ctx, client, info, serviceAddr, errorService, logger)
		if err != nil {
			logger.Error("注册服务发生了错误", logs.F_Error(err))
		}
	}()
	logger.Debug("向Etcd注册服务成功", logs.F_ExtendData(serviceKey))
	return nil
}
