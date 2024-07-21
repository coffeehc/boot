package alts

import (
	"context"
	"errors"
	"fmt"
	"github.com/coffeehc/base/log"
	"github.com/coffeehc/boot/component/grpcx/alts/altsproto"
	"github.com/coffeehc/boot/component/grpcx/alts/internal/commons"
	"github.com/coffeehc/boot/component/grpcx/alts/internal/service"
	"go.uber.org/zap"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/grpclog"
	"net"
	"time"
)

const (
	// hypervisorHandshakerServiceAddress represents the default ALTS gRPC
	// handshaker service address in the hypervisor.
	//hypervisorHandshakerServiceAddress = "dns:///metadata.google.internal.:8080"
	// defaultTimeout specifies the server handshake timeout.
	defaultTimeout = 30.0 * time.Second
	// The following constants specify the minimum and maximum acceptable
	// protocol versions.
	protocolVersionMaxMajor = 2
	protocolVersionMaxMinor = 1
	protocolVersionMinMajor = 2
	protocolVersionMinMinor = 1
)

var (
	maxRPCVersion = &altsproto.RpcProtocolVersions_Version{
		Major: protocolVersionMaxMajor,
		Minor: protocolVersionMaxMinor,
	}
	minRPCVersion = &altsproto.RpcProtocolVersions_Version{
		Major: protocolVersionMinMajor,
		Minor: protocolVersionMinMinor,
	}
	// ErrUntrustedPlatform is returned from ClientHandshake and
	// ServerHandshake is running on a platform where the trustworthiness of
	// the handshaker service is not guaranteed.
	ErrUntrustedPlatform = errors.New("ALTS: untrusted platform. ALTS is only supported on GCP")
	logger               = grpclog.Component("alts")
)

// AuthInfo exposes security information from the ALTS handshake to the
// application. This interface is to be implemented by ALTS. Users should not
// need a brand new implementation of this interface. For situations like
// testing, any new implementation should embed this interface. This allows
// ALTS to add new methods to this interface.
type AuthInfo interface {
	// ApplicationProtocol returns application protocol negotiated for the
	// ALTS connection.
	ApplicationProtocol() string
	// RecordProtocol returns the record protocol negotiated for the ALTS
	// connection.
	RecordProtocol() string
	// SecurityLevel returns the security level of the created ALTS secure
	// channel.
	SecurityLevel() altsproto.SecurityLevel
	// PeerServiceAccount returns the peer service account.
	PeerServiceAccount() string
	// LocalServiceAccount returns the local service account.
	LocalServiceAccount() string
	// PeerRPCVersions returns the RPC version supported by the peer.
	PeerRPCVersions() *altsproto.RpcProtocolVersions
}

// altsTC is the credentials required for authenticating a connection using ALTS.
// It implements credentials.TransportCredentials interface.
type altsTC struct {
	info        *credentials.ProtocolInfo
	side        int
	accounts    []string
	hsAddress   string
	serviceName string
}

func NewALTS(side int, accounts []string, hsAddress string, serviceName string) credentials.TransportCredentials {
	if hsAddress == "" {
		log.Error("没有指定ALTS中心地址")
		return nil
	}
	return &altsTC{
		info: &credentials.ProtocolInfo{
			SecurityProtocol: "alts",
			SecurityVersion:  "1.0",
		},
		side:        side,
		accounts:    accounts,
		hsAddress:   hsAddress,
		serviceName: serviceName,
	}
}

// ClientHandshake implements the client side handshake protocol.
func (g *altsTC) ClientHandshake(ctx context.Context, addr string, rawConn net.Conn) (_ net.Conn, _ credentials.AuthInfo, err error) {
	// Connecting to ALTS handshaker service.
	hsConn, err := service.Dial(g.hsAddress)
	if err != nil {
		log.Debug("连接认证中心失败", zap.Error(err))
		return nil, nil, err
	}
	// Do not close hsConn since it is shared with other handshakes.

	// Possible context leak:
	// The cancel function for the child context we create will only be
	// called a non-nil error is returned.
	var cancel context.CancelFunc
	ctx, cancel = context.WithCancel(ctx)
	defer func() {
		if err != nil {
			cancel()
		}
	}()
	opts := commons.DefaultClientHandshakerOptions(g.serviceName)
	opts.TargetName = addr
	opts.TargetServiceAccounts = g.accounts
	opts.ClientIdentity = &altsproto.Identity{
		ServiceName: g.serviceName,
	}
	opts.RPCVersions = &altsproto.RpcProtocolVersions{
		MaxRpcVersion: maxRPCVersion,
		MinRpcVersion: minRPCVersion,
	}
	chs, err := commons.NewClientHandshaker(ctx, hsConn, rawConn, opts)
	if err != nil {
		return nil, nil, err
	}
	defer func() {
		if err != nil {
			chs.Close()
		}
	}()
	secConn, authInfo, err := chs.ClientHandshake(ctx)
	if err != nil {
		log.Debug("认证握手是失败", zap.Error(err))
		return nil, nil, err
	}
	altsAuthInfo, ok := authInfo.(AuthInfo)
	if !ok {
		return nil, nil, errors.New("client-side auth info is not of type alts.AuthInfo")
	}
	match, _ := checkRPCVersions(opts.RPCVersions, altsAuthInfo.PeerRPCVersions())
	if !match {
		log.Debug("RPC版本不匹配", zap.Error(err))
		return nil, nil, fmt.Errorf("server-side RPC versions are not compatible with this client, local versions: %v, peer versions: %v", opts.RPCVersions, altsAuthInfo.PeerRPCVersions())
	}
	return secConn, authInfo, nil
}

// ServerHandshake implements the server side ALTS handshaker.
func (g *altsTC) ServerHandshake(rawConn net.Conn) (_ net.Conn, _ credentials.AuthInfo, err error) {
	// Connecting to ALTS handshaker service.
	hsConn, err := service.Dial(g.hsAddress)
	if err != nil {
		return nil, nil, err
	}
	// Do not close hsConn since it's shared with other handshakes.
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	opts := commons.DefaultServerHandshakerOptions(g.serviceName)
	opts.TargetServiceAccounts = g.accounts
	opts.RPCVersions = &altsproto.RpcProtocolVersions{
		MaxRpcVersion: maxRPCVersion,
		MinRpcVersion: minRPCVersion,
	}
	shs, err := commons.NewServerHandshaker(ctx, hsConn, rawConn, opts)
	if err != nil {
		return nil, nil, err
	}
	defer func() {
		if err != nil {
			shs.Close()
		}
	}()
	secConn, authInfo, err := shs.ServerHandshake(ctx)
	if err != nil {
		return nil, nil, err
	}
	altsAuthInfo, ok := authInfo.(AuthInfo)
	if !ok {
		return nil, nil, errors.New("server-side auth info is not of type alts.AuthInfo")
	}
	match, _ := checkRPCVersions(opts.RPCVersions, altsAuthInfo.PeerRPCVersions())
	if !match {
		return nil, nil, fmt.Errorf("client-side RPC versions is not compatible with this server, local versions: %v, peer versions: %v", opts.RPCVersions, altsAuthInfo.PeerRPCVersions())
	}
	return secConn, authInfo, nil
}

func (g *altsTC) Info() credentials.ProtocolInfo {
	return *g.info
}

func (g *altsTC) Clone() credentials.TransportCredentials {
	info := *g.info
	var accounts []string
	if g.accounts != nil {
		accounts = make([]string, len(g.accounts))
		copy(accounts, g.accounts)
	}
	return &altsTC{
		info:      &info,
		side:      g.side,
		hsAddress: g.hsAddress,
		accounts:  accounts,
	}
}

func (g *altsTC) OverrideServerName(serverNameOverride string) error {
	g.info.ServerName = serverNameOverride
	return nil
}

// compareRPCVersion returns 0 if v1 == v2, 1 if v1 > v2 and -1 if v1 < v2.
func compareRPCVersions(v1, v2 *altsproto.RpcProtocolVersions_Version) int {
	switch {
	case v1.GetMajor() > v2.GetMajor(),
		v1.GetMajor() == v2.GetMajor() && v1.GetMinor() > v2.GetMinor():
		return 1
	case v1.GetMajor() < v2.GetMajor(),
		v1.GetMajor() == v2.GetMajor() && v1.GetMinor() < v2.GetMinor():
		return -1
	}
	return 0
}

// checkRPCVersions performs a version check between local and peer rpc protocol
// versions. This function returns true if the check passes which means both
// parties agreed on a common rpc protocol to use, and false otherwise. The
// function also returns the highest common RPC protocol version both parties
// agreed on.
func checkRPCVersions(local, peer *altsproto.RpcProtocolVersions) (bool, *altsproto.RpcProtocolVersions_Version) {
	if local == nil || peer == nil {
		logger.Error("invalid checkRPCVersions argument, either local or peer is nil.")
		return false, nil
	}

	// maxCommonVersion is MIN(local.max, peer.max).
	maxCommonVersion := local.GetMaxRpcVersion()
	if compareRPCVersions(local.GetMaxRpcVersion(), peer.GetMaxRpcVersion()) > 0 {
		maxCommonVersion = peer.GetMaxRpcVersion()
	}

	// minCommonVersion is MAX(local.min, peer.min).
	minCommonVersion := peer.GetMinRpcVersion()
	if compareRPCVersions(local.GetMinRpcVersion(), peer.GetMinRpcVersion()) > 0 {
		minCommonVersion = local.GetMinRpcVersion()
	}

	if compareRPCVersions(maxCommonVersion, minCommonVersion) < 0 {
		return false, nil
	}
	return true, maxCommonVersion
}
