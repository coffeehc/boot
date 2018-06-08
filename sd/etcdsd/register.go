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
)

func RegisterService(ctx context.Context, client *clientv3.Client, info boot.ServiceInfo, serviceAddr string) errors.Error {
	logger := logs.GetLogger(ctx)
	rootErrorService := errors.GetRootErrorService(ctx)
	errorService := rootErrorService.NewService("sd")
	if ctx.Err() != nil {
		return errorService.BuildMessageError("服务注册已经关闭")
	}
	ttl := int64(15)
	lease := clientv3.NewLease(client)
	resp, err := lease.Grant(ctx, ttl)
	if err != nil {
		return errorService.BuildWappedSystemError(err)
	}
	serviceKey := fmt.Sprintf("%s%s", sd.BuildServiceKeyPrefix(info), serviceAddr)
	value, err := ffjson.Marshal(&sd.ServiceRegisterInfo{ServiceInfo: info.(*boot.SimpleServiceInfo), ServerAddr: serviceAddr})
	if err != nil {
		return errorService.BuildWappedSystemError(err)
	}
	_, err = client.Put(ctx, serviceKey, string(value), clientv3.WithLease(resp.ID))
	if err != nil {
		return errorService.BuildWappedSystemError(err)
	}
	ch, err := lease.KeepAlive(ctx, resp.ID)
	if err != nil {
		return errorService.BuildWappedSystemError(err)
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
		err := RegisterService(ctx, client, info, serviceAddr)
		if err != nil {
			logger.Error("注册服务发生了错误", logs.F_Error(err))
		}
	}()
	return nil
}
