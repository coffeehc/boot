package mqservice

import (
	"context"
	"fmt"

	"time"

	"git.xiagaogao.com/coffee/boot/errors"
	"github.com/coffeehc/logger"
	"github.com/streadway/amqp"
)

type Producer interface {
	Close()
	GetChannel() (*amqp.Channel, errors.Error)
	SyncPublish(exchange, key string, msg []byte) errors.Error
	AsyncPublish(exchange, key string, msg []byte)
	SyncPublishing(exchange, key string, publishing amqp.Publishing) errors.Error
	AsyncPublishing(exchange, key string, publishing amqp.Publishing)
}

func NewProducer(name string, vHostConfig *VhostConfig, concurrency int, confirm bool, init func(context.Context, *amqp.Channel) errors.Error) (Producer, errors.Error) {
	mqService, err := NewMQService(vHostConfig)
	if err != nil {
		return nil, err
	}
	if concurrency == 0 {
		concurrency = 1
	}
	sessions := make(chan *producerSession, concurrency)
	for i := 0; i < concurrency; i++ {
		session, err := newProducerSession(fmt.Sprintf("%s-%d", name, i), mqService, confirm)
		if err != nil {
			return nil, err
		}
		sessions <- session
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
	publishQueue := make(chan producerInfo, concurrency)
	impl := &producerImpl{
		name:         name,
		mqService:    mqService,
		concurrency:  concurrency,
		confirm:      confirm,
		closeFunc:    cancel,
		publishQueue: publishQueue,
		sessions:     sessions,
	}
	go func() {
		for {
			select {
			case info := <-publishQueue:
				err := impl.SyncPublishing(info.exchange, info.key, info.publishing)
				if err != nil {
					go func() {
						time.Sleep(time.Millisecond * 300)
						publishQueue <- info
					}()

				}
			}
		}
	}()
	return impl, nil
}

type producerInfo struct {
	exchange   string
	key        string
	publishing amqp.Publishing
}

type producerImpl struct {
	name         string
	mqService    MQService
	concurrency  int
	confirm      bool
	publishQueue chan producerInfo
	closeFunc    func()
	sessions     chan *producerSession
}

func (impl *producerImpl) Close() {
	impl.closeFunc()
	impl.mqService.Close()
}

func (impl *producerImpl) SyncPublish(exchange, key string, msg []byte) errors.Error {
	session := <-impl.sessions
	defer func() {
		impl.sessions <- session
	}()
	return session.publish(exchange, key, amqp.Publishing{
		Body: msg,
	})
}

func (impl *producerImpl) AsyncPublish(exchange, key string, msg []byte) {
	impl.publishQueue <- producerInfo{exchange, key, amqp.Publishing{
		Body: msg,
	}}
}

func (impl *producerImpl) SyncPublishing(exchange, key string, publishing amqp.Publishing) errors.Error {
	session := <-impl.sessions
	defer func() {
		impl.sessions <- session
	}()
	return session.publish(exchange, key, publishing)
}

func (impl *producerImpl) AsyncPublishing(exchange, key string, publishing amqp.Publishing) {
	impl.publishQueue <- producerInfo{exchange, key, publishing}
}

func (impl *producerImpl) GetChannel() (*amqp.Channel, errors.Error) {
	session := <-impl.sessions
	return session.getChannel()
}

type producerSession struct {
	id          string
	confirm     bool
	confirmChan chan amqp.Confirmation
	mqService   MQService
}

func newProducerSession(name string, mqService MQService, confirm bool) (*producerSession, errors.Error) {
	session := &producerSession{
		id:        name,
		confirm:   confirm,
		mqService: mqService,
	}
	_, err := session.getChannel()
	if err != nil {
		return nil, err
	}
	return session, nil
}

func (session *producerSession) getChannel() (*amqp.Channel, errors.Error) {
	channel, err := session.mqService.GetChannel(session.id, true, func(ctx context.Context, channel *amqp.Channel) errors.Error {
		if !session.confirm {
			return nil
		}
		err := channel.Confirm(false)
		if err != nil {
			return errors.NewErrorWrapper(errors.Error_System, "mq", err)
		}
		session.confirmChan = channel.NotifyPublish(make(chan amqp.Confirmation, 1))
		return nil
	})
	if err != nil {
		return nil, err
	}
	return channel, nil
}

func (session *producerSession) publish(exchange, key string, msg amqp.Publishing) errors.Error {
	channel, err := session.getChannel()
	if err != nil {
		logger.Error("获取channel失败,1秒后重新获取")
		time.Sleep(time.Second)
		return session.publish(exchange, key, msg)
	}
	_err := channel.Publish(exchange, key, false, false, msg)
	if _err != nil {
		return errors.NewErrorWrapper(errors.Error_System, "mq", _err)
	}
	if session.confirm {
		confirmed, ok := <-session.confirmChan
		if !ok || !confirmed.Ack {
			return errors.NewError(errors.Error_Message, "mq", "通道关闭或消息未投递成功,请重试!")
		}
	}
	return nil
}
