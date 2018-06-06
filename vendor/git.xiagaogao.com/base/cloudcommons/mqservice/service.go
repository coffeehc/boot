package mqservice

import (
	"time"

	"sync"

	"context"

	"git.xiagaogao.com/coffee/boot/errors"
	"github.com/coffeehc/logger"
	"github.com/streadway/amqp"
)

//amqp://guest:guest@localhost:5672/

type MQService interface {
	GetConnection() (*amqp.Connection, errors.Error)
	GetChannel(name string, reCreate bool, init func(context.Context, *amqp.Channel) errors.Error) (*amqp.Channel, errors.Error)
	CloseChannel(name string)
	Close()
}

func NewMQService(vHostConfig *VhostConfig) (MQService, errors.Error) {
	impl := &mqServiceImpl{
		vHostConfig:  vHostConfig,
		connMutex:    new(sync.Mutex),
		channelMutex: new(sync.Mutex),
		channels:     new(sync.Map), //make(map[string]*amqp.Channel),
	}
	_, err := impl.GetConnection()
	if err != nil {
		return nil, err
	}
	return impl, nil
}

type mqServiceImpl struct {
	vHostConfig  *VhostConfig
	conn         *amqp.Connection
	channels     *sync.Map //map[string]*amqp.Channel
	connMutex    *sync.Mutex
	channelMutex *sync.Mutex
	close        bool
}

func (impl *mqServiceImpl) GetConnection() (*amqp.Connection, errors.Error) {
	if impl.close {
		return nil, errors.NewError(errors.Error_Message, "mq", "service close")
	}
	if impl.conn == nil {
		impl.connMutex.Lock()
		defer impl.connMutex.Unlock()
		if impl.conn != nil {
			return impl.conn, nil
		}
		conn, err := amqp.DialConfig(impl.vHostConfig.toUrl(), amqp.Config{Heartbeat: time.Second, Locale: "UTF-8"})
		if err != nil {
			logger.Error("连接MQ失败,%s", err)
			return nil, errors.NewErrorWrapper(errors.Error_System, "mq", err)
		}
		impl.conn = conn
		notify := make(chan *amqp.Error, 1)
		conn.NotifyClose(notify)
		go func(notify chan *amqp.Error) {
			timer := time.NewTimer(time.Minute * 5)
			check := true
			for check {
				select {
				case <-notify:
					logger.Debug("链接已经关闭,重新创建")
					impl.conn = nil
					go impl.GetConnection()
					check = false
					break
				case <-timer.C:
					timer.Reset(time.Minute * 5)
				}
			}
		}(notify)
	}
	return impl.conn, nil
}

func (impl *mqServiceImpl) GetChannel(name string, reCreate bool, init func(context.Context, *amqp.Channel) errors.Error) (*amqp.Channel, errors.Error) {
	if impl.close {
		return nil, errors.NewError(errors.Error_Message, "mq", "service close")
	}
	v, ok := impl.channels.Load(name)
	if !ok {
		logger.Debug("获取Channel[%s],加锁", name)
		impl.channelMutex.Lock()
		defer impl.channelMutex.Unlock()
		conn, err := impl.GetConnection()
		if err != nil {
			time.Sleep(time.Second)
			return impl.GetChannel(name, reCreate, init)
		}
		v, ok = impl.channels.Load(name)
		if ok {
			return v.(*amqp.Channel), nil
		}
		channel, _err := conn.Channel()
		if _err != nil {
			return nil, errors.NewErrorWrapper(errors.Error_System, "mq", _err)
		}
		impl.channels.Store(name, channel)
		cxt, cancel := context.WithCancel(context.Background())
		if reCreate {
			notify := make(chan *amqp.Error, 1)
			channel.NotifyClose(notify)
			go func(name string, notify chan *amqp.Error) {
				timer := time.NewTimer(time.Minute * 3)
				check := true
				for check {
					select {
					case <-notify:
						cancel()
						logger.Debug("Channel[%s]已经关闭,重新创建", name)
						impl.channels.Delete(name)
						go impl.GetChannel(name, reCreate, init)
						check = false
						break
					case <-timer.C:
						timer.Reset(time.Minute * 3)
					}
				}
			}(name, notify)
		}
		if init != nil {
			err = init(cxt, channel)
			if err != nil {
				return nil, err
			}
		}
		return channel, nil
	}
	return v.(*amqp.Channel), nil
}

func (impl *mqServiceImpl) CloseChannel(name string) {
	impl.channelMutex.Lock()
	defer impl.channelMutex.Unlock()
	channel, ok := impl.channels.Load(name)
	if !ok {
		return
	}
	impl.channels.Delete(name)
	channel.(*amqp.Channel).Close()
}

func (impl *mqServiceImpl) Close() {
	impl.close = true
	impl.channels.Range(func(key interface{}, value interface{}) bool {
		value.(*amqp.Channel).Close()
		return true
	})
	impl.conn.Close()
}
