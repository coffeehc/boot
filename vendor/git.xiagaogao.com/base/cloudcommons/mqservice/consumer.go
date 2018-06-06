package mqservice

import (
	"context"

	"time"

	"fmt"
	"sync/atomic"

	"git.xiagaogao.com/coffee/boot/errors"
	"github.com/coffeehc/logger"
	"github.com/streadway/amqp"
)

type ConsumerHandler func(ctx context.Context, msg amqp.Delivery, errs chan<- errors.Error) (ack bool, requeue bool)

type Consumer interface {
	Close()
	Consume(queue string, consumerSize int, prefetchCount int, handleTimeout time.Duration, rootCtx context.Context, handler ConsumerHandler) <-chan errors.Error
}

//创建一个Consumer服务,统一都不默认Ack.默认重新投递
func NewConsumer(name string, vHostConfig *VhostConfig, init func(context.Context, *amqp.Channel) errors.Error) (Consumer, errors.Error) {
	mqService, err := NewMQService(vHostConfig)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	if init != nil {
		channel, err := mqService.GetChannel(name+"-init", false, nil)
		if err != nil {
			return nil, err
		}
		defer mqService.CloseChannel(name + "-init")
		err = init(ctx, channel)
		if err != nil {
			return nil, err
		}
	}
	return &consumerImpl{
		name:      name,
		mqService: mqService,
		closeFunc: cancel,
		close:     make(chan struct{}, 1),
	}, nil
}

type consumerImpl struct {
	name          string
	mqService     MQService
	closeFunc     func()
	close         chan struct{}
	consumerIndex int64
}

//设置Qos,设置Comsume数量
func (impl *consumerImpl) Consume(queue string, consumerSize int, prefetchCount int, handleTimeout time.Duration, rootCtx context.Context, handler ConsumerHandler) <-chan errors.Error {
	errs := make(chan errors.Error, consumerSize+1000)
	if handler == nil {
		errs <- errors.NewError(errors.Error_Message, "mq", "没有指定消息处理函数")
		return errs
	}
	for i := 0; i < consumerSize; i++ {
		go func() {
			closeNotify := make(chan *amqp.Error, 1)
			timer := time.NewTimer(time.Second)
			closeNotify <- &amqp.Error{
				Server:  false,
				Recover: false,
			}
			name := fmt.Sprintf("%s-%d", impl.name, atomic.AddInt64(&impl.consumerIndex, 1))
			for {
				select {
				case <-closeNotify:
					//name := impl.name
					channel, err := impl.mqService.GetChannel(name, true, func(ctx context.Context, channel *amqp.Channel) errors.Error {
						return nil
					})
					if err != nil {
						logger.Error("创建Channel失败,%s", err)
						errs <- err
						return
					}
					channel.Qos(prefetchCount, 0, false)
					closeNotify = channel.NotifyClose(make(chan *amqp.Error, 1))
					go func() {
						deliveries, err := channel.Consume(queue, name, false, false, false, false, nil)
						if err != nil {
							errs <- errors.NewErrorWrapper(errors.Error_System, "mq", err)
						}
						pool := make(chan struct{}, prefetchCount)
						for msg := range deliveries {
							pool <- struct{}{}
							go func(rootCtx context.Context, msg amqp.Delivery) {
								defer func() {
									<-pool
								}()
								var ctx context.Context
								if handleTimeout == 0 {
									ctx = rootCtx
								}
								ctx, _ = context.WithTimeout(rootCtx, handleTimeout)
								timeOut := false
								var ack int64 = 0
								go func() {
									select {
									case <-ctx.Done():
										if atomic.CompareAndSwapInt64(&ack, 0, 1) {
											timeOut = true
											logger.Warn("操作超时")
											msg.Nack(false, true)
										}
									}
								}()
								ok, requeue := handler(ctx, msg, errs)
								if atomic.CompareAndSwapInt64(&ack, 0, 1) {
									if ok && !requeue {
										msg.Ack(false)
									} else {
										msg.Nack(false, requeue)
									}

								}
							}(rootCtx, msg)
						}
					}()
				case <-impl.close:
					return
				case <-timer.C:
					timer.Reset(time.Second)
				}
			}
		}()
	}
	return errs
}

func (impl *consumerImpl) Close() {
	close(impl.close)
	impl.closeFunc()
	impl.mqService.Close()
}
