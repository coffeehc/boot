package commons

import (
	"github.com/coffeehc/boot/component/grpc/alts/altsproto"
	"google.golang.org/grpc/credentials"
)

var _ credentials.AuthInfo = (*altsAuthInfo)(nil)

// altsAuthInfo exposes security information from the ALTS handshake to the
// application. altsAuthInfo is immutable and implements credentials.AuthInfo.
type altsAuthInfo struct {
	p *altsproto.AltsContext
	credentials.CommonAuthInfo
}

// New returns a new altsAuthInfo object given handshaker results.
func NewAuthInfo(result *altsproto.HandshakerResult) credentials.AuthInfo {
	return newAuthInfo(result)
}

func newAuthInfo(result *altsproto.HandshakerResult) *altsAuthInfo {
	return &altsAuthInfo{
		p: &altsproto.AltsContext{
			ApplicationProtocol: result.GetApplicationProtocol(),
			RecordProtocol:      result.GetRecordProtocol(),
			// TODO: assign security level from result.
			SecurityLevel:       altsproto.SecurityLevel_INTEGRITY_AND_PRIVACY,
			PeerServiceAccount:  result.GetPeerIdentity().GetServiceAccount(),
			LocalServiceAccount: result.GetLocalIdentity().GetServiceAccount(),
			PeerRpcVersions:     result.GetPeerRpcVersions(),
			PeerAttributes:      result.GetPeerIdentity().GetAttributes(),
		},
		CommonAuthInfo: credentials.CommonAuthInfo{SecurityLevel: credentials.PrivacyAndIntegrity},
	}
}

// AuthType identifies the context as providing ALTS authentication information.
func (s *altsAuthInfo) AuthType() string {
	return "alts"
}

// ApplicationProtocol returns the context's application protocol.
func (s *altsAuthInfo) ApplicationProtocol() string {
	return s.p.GetApplicationProtocol()
}

// RecordProtocol returns the context's record protocol.
func (s *altsAuthInfo) RecordProtocol() string {
	return s.p.GetRecordProtocol()
}

// SecurityLevel returns the context's security level.
func (s *altsAuthInfo) SecurityLevel() altsproto.SecurityLevel {
	return s.p.GetSecurityLevel()
}

// PeerServiceAccount returns the context's peer service account.
func (s *altsAuthInfo) PeerServiceAccount() string {
	return s.p.GetPeerServiceAccount()
}

// LocalServiceAccount returns the context's local service account.
func (s *altsAuthInfo) LocalServiceAccount() string {
	return s.p.GetLocalServiceAccount()
}

// PeerRPCVersions returns the context's peer RPC versions.
func (s *altsAuthInfo) PeerRPCVersions() *altsproto.RpcProtocolVersions {
	return s.p.GetPeerRpcVersions()
}

// PeerAttributes returns the context's peer attributes.
func (s *altsAuthInfo) PeerAttributes() map[string]string {
	return s.p.GetPeerAttributes()
}
