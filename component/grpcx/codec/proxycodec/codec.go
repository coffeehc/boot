package proxycodec

import (
	"bytes"
	"fmt"

	"google.golang.org/grpc/encoding"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/protoadapt"
)

// Name 是已有透明代理 codec 的注册名，线上协议保持不变。
const Name = "proxyCodec"

func init() {
	encoding.RegisterCodec(Codec())
}

// Codec 对 ProxyFrame 直接透传 payload，其他消息使用 protobuf 编解码。
// 保留 encoding.Codec 接口，gRPC v1.84 仍支持此接口并自动适配内部 buffer API。
func Codec() encoding.Codec {
	return CodecWithParent(&protoCodec{})
}

// CodecWithParent 对 ProxyFrame 直接透传 payload，其他消息交给非 nil fallback。
// fallback 必须支持并发调用且不得持有 Unmarshal 的临时输入 buffer。
func CodecWithParent(fallback encoding.Codec) encoding.Codec {
	return &proxyCodec{fallback}
}

// proxyCodec 为透明代理保留原始 protobuf 消息的字节表示。
type proxyCodec struct {
	// parentCodec 编解码普通消息，由调用方在构造时提供。
	parentCodec encoding.Codec
}

// 序列化函数，
// 尝试将消息转换为*frame类型，并返回frame的payload实现序列化
// 若失败，则采用变量parentCodec中的Marshal进行序列化
func (c *proxyCodec) Marshal(v interface{}) ([]byte, error) {
	out, ok := v.(*ProxyFrame)
	if !ok {
		return c.parentCodec.Marshal(v)
	}
	return out.GetPayload(), nil

}

// 反序列化函数，
// 尝试通过将消息转为*frame类型，提取出payload到[]byte，实现反序列化
// 若失败，则采用变量parentCodec中的Unmarshal进行反序列化
func (c *proxyCodec) Unmarshal(data []byte, v interface{}) error {
	dst, ok := v.(*ProxyFrame)
	if !ok {
		return c.parentCodec.Unmarshal(data, v)
	}
	if dst == nil {
		return fmt.Errorf("proxycodec: 不能解码到 nil ProxyFrame")
	}
	// gRPC 解码后会释放或复用接收 buffer，消息必须拥有自己的 payload。
	dst.Payload = bytes.Clone(data)
	return nil
}

func (c *proxyCodec) Name() string {
	return Name
}

// protoCodec 支持新旧 protobuf 消息接口，对不支持的类型返回错误。
type protoCodec struct{}

func (protoCodec) Marshal(v interface{}) ([]byte, error) {
	var message proto.Message
	switch value := v.(type) {
	case proto.Message:
		message = value
	case protoadapt.MessageV1:
		message = protoadapt.MessageV2Of(value)
	default:
		return nil, fmt.Errorf("proxycodec: 不支持的 protobuf 类型 %T", v)
	}
	return proto.Marshal(message)
}

func (protoCodec) Unmarshal(data []byte, v interface{}) error {
	var message proto.Message
	switch value := v.(type) {
	case proto.Message:
		message = value
	case protoadapt.MessageV1:
		message = protoadapt.MessageV2Of(value)
	default:
		return fmt.Errorf("proxycodec: 不支持的 protobuf 类型 %T", v)
	}
	return proto.Unmarshal(data, message)
}

func (protoCodec) Name() string {
	return "proto"
}
