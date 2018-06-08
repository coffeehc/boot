// Code generated by protoc-gen-go. DO NOT EDIT.
// source: baseService.proto

/*
Package baseservicecommon is a generated protocol buffer package.

It is generated from these files:
	baseService.proto
	securityService.proto
	sequenceService.proto

It has these top-level messages:
	IDCardGenerate
	IdCard
	KeyGenerate
	KeyQuery
	KeyInfo
	CaptchaRequest
	Captcha
	SequenceGenerate
	SequenceId
	SequenceInfo
*/
package baseservicecommon

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import _ "google.golang.org/genproto/googleapis/api/annotations"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

// 身份证导出请求
type IDCardGenerate struct {
}

func (m *IDCardGenerate) Reset()                    { *m = IDCardGenerate{} }
func (m *IDCardGenerate) String() string            { return proto.CompactTextString(m) }
func (*IDCardGenerate) ProtoMessage()               {}
func (*IDCardGenerate) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

type IdCard struct {
	Id   string `protobuf:"bytes,1,opt,name=id" json:"id,omitempty"`
	Name string `protobuf:"bytes,2,opt,name=name" json:"name,omitempty"`
}

func (m *IdCard) Reset()                    { *m = IdCard{} }
func (m *IdCard) String() string            { return proto.CompactTextString(m) }
func (*IdCard) ProtoMessage()               {}
func (*IdCard) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *IdCard) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *IdCard) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func init() {
	proto.RegisterType((*IDCardGenerate)(nil), "baseservice.base.IDCardGenerate")
	proto.RegisterType((*IdCard)(nil), "baseservice.base.IdCard")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for BaseService service

type BaseServiceClient interface {
	// 获取一个新的 身份证信息
	GenerateIDCard(ctx context.Context, in *IDCardGenerate, opts ...grpc.CallOption) (*IdCard, error)
}

type baseServiceClient struct {
	cc *grpc.ClientConn
}

func NewBaseServiceClient(cc *grpc.ClientConn) BaseServiceClient {
	return &baseServiceClient{cc}
}

func (c *baseServiceClient) GenerateIDCard(ctx context.Context, in *IDCardGenerate, opts ...grpc.CallOption) (*IdCard, error) {
	out := new(IdCard)
	err := grpc.Invoke(ctx, "/baseservice.base.BaseService/GenerateIDCard", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for BaseService service

type BaseServiceServer interface {
	// 获取一个新的 身份证信息
	GenerateIDCard(context.Context, *IDCardGenerate) (*IdCard, error)
}

func RegisterBaseServiceServer(s *grpc.Server, srv BaseServiceServer) {
	s.RegisterService(&_BaseService_serviceDesc, srv)
}

func _BaseService_GenerateIDCard_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(IDCardGenerate)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BaseServiceServer).GenerateIDCard(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/baseservice.base.BaseService/GenerateIDCard",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BaseServiceServer).GenerateIDCard(ctx, req.(*IDCardGenerate))
	}
	return interceptor(ctx, in, info, handler)
}

var _BaseService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "baseservice.base.BaseService",
	HandlerType: (*BaseServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GenerateIDCard",
			Handler:    _BaseService_GenerateIDCard_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "baseService.proto",
}

func init() { proto.RegisterFile("baseService.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 238 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x12, 0x4c, 0x4a, 0x2c, 0x4e,
	0x0d, 0x4e, 0x2d, 0x2a, 0xcb, 0x4c, 0x4e, 0xd5, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0x12, 0x00,
	0x09, 0x15, 0x43, 0x85, 0x40, 0x6c, 0x29, 0x99, 0xf4, 0xfc, 0xfc, 0xf4, 0x9c, 0x54, 0xfd, 0xc4,
	0x82, 0x4c, 0xfd, 0xc4, 0xbc, 0xbc, 0xfc, 0x92, 0xc4, 0x92, 0xcc, 0xfc, 0xbc, 0x62, 0x88, 0x7a,
	0x25, 0x01, 0x2e, 0x3e, 0x4f, 0x17, 0xe7, 0xc4, 0xa2, 0x14, 0xf7, 0xd4, 0xbc, 0xd4, 0xa2, 0xc4,
	0x92, 0x54, 0x25, 0x1d, 0x2e, 0x36, 0xcf, 0x14, 0x90, 0x88, 0x10, 0x1f, 0x17, 0x53, 0x66, 0x8a,
	0x04, 0xa3, 0x02, 0xa3, 0x06, 0x67, 0x10, 0x53, 0x66, 0x8a, 0x90, 0x10, 0x17, 0x4b, 0x5e, 0x62,
	0x6e, 0xaa, 0x04, 0x13, 0x58, 0x04, 0xcc, 0x36, 0x6a, 0x63, 0xe4, 0xe2, 0x76, 0x42, 0xb8, 0x42,
	0xa8, 0x9c, 0x8b, 0x0f, 0x66, 0x12, 0xc4, 0x5c, 0x21, 0x05, 0x3d, 0x74, 0x27, 0xe9, 0xa1, 0xda,
	0x28, 0x25, 0x81, 0x45, 0x05, 0xd8, 0x05, 0x4a, 0x9a, 0x4d, 0x97, 0x9f, 0x4c, 0x66, 0x52, 0x56,
	0x92, 0x43, 0x52, 0x00, 0xf6, 0x48, 0x99, 0xa1, 0x3e, 0x48, 0x48, 0x3f, 0x33, 0x25, 0x39, 0xb1,
	0x28, 0xc5, 0x8a, 0x51, 0xcb, 0x29, 0x20, 0xca, 0x2f, 0x3d, 0xb3, 0x44, 0xaf, 0x22, 0x33, 0x31,
	0x3d, 0x31, 0x3f, 0x3d, 0x31, 0x5f, 0x2f, 0x39, 0x3f, 0x57, 0x1f, 0x49, 0x5b, 0x31, 0x32, 0x07,
	0x99, 0x9d, 0x9c, 0x9f, 0x9b, 0x9b, 0x9f, 0x67, 0x8d, 0x21, 0x92, 0xc4, 0x06, 0x0e, 0x21, 0x63,
	0x40, 0x00, 0x00, 0x00, 0xff, 0xff, 0x44, 0x54, 0x97, 0x32, 0x66, 0x01, 0x00, 0x00,
}