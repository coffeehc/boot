package baseserviceclient

import (
	"context"

	"git.xiagaogao.com/baseservices/baseservice/baseservicecommon"

	"github.com/coffeehc/microserviceboot/base"
	"github.com/coffeehc/microserviceboot/loadbalancer"
	"github.com/coffeehc/microserviceboot/serviceclient/grpcclient"
)

type BaseServiceClient interface {
	GetSequenceService() baseservicecommon.SequenceServiceClient
	GetSecurityServiceClient() baseservicecommon.SecurityServiceClient
	CreateSequence() (int64, error)
	GetSequenceInfo(sequence int64) (*baseservicecommon.SequenceInfo, error)
	GenerateRandKey(scope string, bits int32, storageType baseservicecommon.StorageType, expireMs int64) (*baseservicecommon.KeyInfo, error)
	GenerateRSAKey(scope string, bits int32, storageType baseservicecommon.StorageType, expireMs int64) (*baseservicecommon.KeyInfo, error)
	GetRSAKey(scope string, keyId int64, storageType baseservicecommon.StorageType) (*baseservicecommon.KeyInfo, error)
	GetRandKey(scope string, keyId int64, storageType baseservicecommon.StorageType) (*baseservicecommon.KeyInfo, error)
	GenerateIdCard() (*baseservicecommon.IdCard, error)
}

func NewBaseServiceClient(cxt context.Context, serviceInfo base.ServiceInfo, grpcClient grpcclient.GRPCClient, balancer loadbalancer.Balancer, block bool) (BaseServiceClient, base.Error) {
	baseServiceClient := &_BaseServiceClient{}
	clientConn, err := grpcClient.NewClientConn(cxt, serviceInfo, balancer, 0, block)
	if err != nil {
		return nil, err
	}
	baseServiceClient.sequenceServiceClient = baseservicecommon.NewSequenceServiceClient(clientConn)
	baseServiceClient.securityServiceClient = baseservicecommon.NewSecurityServiceClient(clientConn)
	baseServiceClient.baseServiceClient = baseservicecommon.NewBaseServiceClient(clientConn)
	return baseServiceClient, nil
}

type _BaseServiceClient struct {
	sequenceServiceClient baseservicecommon.SequenceServiceClient
	securityServiceClient baseservicecommon.SecurityServiceClient
	baseServiceClient     baseservicecommon.BaseServiceClient
}

func (this *_BaseServiceClient) GetSequenceService() baseservicecommon.SequenceServiceClient {
	return this.sequenceServiceClient
}

func (this *_BaseServiceClient) GetSecurityServiceClient() baseservicecommon.SecurityServiceClient {
	return this.securityServiceClient
}

var _sequenceGenerate = &baseservicecommon.SequenceGenerate{}

func (this *_BaseServiceClient) CreateSequence() (int64, error) {
	sequenceId, err := this.sequenceServiceClient.GenerateSequence(context.Background(), _sequenceGenerate)
	if err != nil {
		return 0, err
	}
	return sequenceId.SequenceId, nil
}

func (this *_BaseServiceClient) GetSequenceInfo(sequence int64) (*baseservicecommon.SequenceInfo, error) {
	return this.sequenceServiceClient.GetSequenceInfo(context.Background(), &baseservicecommon.SequenceId{
		SequenceId: sequence,
	})
}

func (this *_BaseServiceClient) GenerateRandKey(scope string, bits int32, storageType baseservicecommon.StorageType, expireMs int64) (*baseservicecommon.KeyInfo, error) {
	return this.securityServiceClient.GenerateRandKey(context.Background(), &baseservicecommon.KeyGenerate{
		Scope:    scope,
		Bits:     bits,
		Type:     storageType,
		ExpireMs: expireMs,
	})
}

func (this *_BaseServiceClient) GenerateRSAKey(scope string, bits int32, storageType baseservicecommon.StorageType, expireMs int64) (*baseservicecommon.KeyInfo, error) {
	return this.securityServiceClient.GenerateRSAKey(context.Background(), &baseservicecommon.KeyGenerate{
		Scope:    scope,
		Bits:     bits,
		Type:     storageType,
		ExpireMs: expireMs,
	})
}

func (this *_BaseServiceClient) GetRSAKey(scope string, keyId int64, storageType baseservicecommon.StorageType) (*baseservicecommon.KeyInfo, error) {
	return this.securityServiceClient.GetRSAKey(context.Background(), &baseservicecommon.KeyQuery{
		Scope: scope,
		KeyId: keyId,
		Type:  storageType,
	})
}

func (this *_BaseServiceClient) GetRandKey(scope string, keyId int64, storageType baseservicecommon.StorageType) (*baseservicecommon.KeyInfo, error) {
	return this.securityServiceClient.GetRandKey(context.Background(), &baseservicecommon.KeyQuery{
		Scope: scope,
		KeyId: keyId,
		Type:  storageType,
	})
}

func (this *_BaseServiceClient) GenerateIdCard() (*baseservicecommon.IdCard, error) {
	return this.baseServiceClient.GenerateIDCard(context.Background(), &baseservicecommon.IDCardGenerate{})
}
