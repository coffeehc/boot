package commons

import (
	"context"
	"errors"
	"fmt"
	"github.com/coffeehc/base/log"
	"github.com/coffeehc/boot/component/grpcx/alts/altsproto"
	"github.com/coffeehc/boot/component/grpcx/alts/internal"
	"github.com/coffeehc/boot/component/grpcx/alts/internal/conn"
	"go.uber.org/zap"
	"golang.org/x/sync/semaphore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"io"
	"net"
	"time"
)

const (
	// The maximum byte size of receive frames.
	frameLimit              = 64 * 1024 // 64 KB
	rekeyRecordProtocolName = "ALTSRP_GCM_AES128_REKEY"
)

var (
	hsProtocol      = altsproto.HandshakeProtocol_ALTS
	appProtocols    = []string{"grpc"}
	recordProtocols = []string{rekeyRecordProtocolName}
	keyLength       = map[string]int{
		rekeyRecordProtocolName: 44,
	}
	altsRecordFuncs = map[string]internal.ALTSRecordFunc{
		// ALTS handshaker protocols.
		rekeyRecordProtocolName: func(side int, keyData []byte) (internal.ALTSRecordCrypto, error) {
			return conn.NewAES128GCMRekey(side, keyData)
		},
	}
	// control number of concurrent created (but not closed) handshakes.
	clientHandshakes = semaphore.NewWeighted(1000)
	serverHandshakes = semaphore.NewWeighted(1000)

	// errOutOfBound occurs when the handshake service returns a consumed
	// bytes value larger than the buffer that was passed to it originally.
	errOutOfBound = errors.New("handshake service consumed bytes value is out-of-bound")
)

func init() {
	for protocol, f := range altsRecordFuncs {
		if err := conn.RegisterProtocol(protocol, f); err != nil {
			panic(err)
		}
	}
}

// ClientHandshakerOptions contains the client handshaker options that can
// provided by the caller.
type ClientHandshakerOptions struct {
	ClientIdentity        *altsproto.Identity
	TargetName            string
	TargetServiceAccounts []string
	RPCVersions           *altsproto.RpcProtocolVersions
	ServiceName           string
}

// ServerHandshakerOptions contains the server handshaker options that can
// provided by the caller.
type ServerHandshakerOptions struct {
	RPCVersions           *altsproto.RpcProtocolVersions
	TargetServiceAccounts []string
	ServiceName           string
}

// DefaultClientHandshakeOptions returns the default client handshaker options.
func DefaultClientHandshakeOptions(serviceName string) *ClientHandshakerOptions {
	return &ClientHandshakerOptions{
		ServiceName: serviceName,
	}
}

// DefaultServerHandshakeOptions returns the default client handshaker options.
func DefaultServerHandshakeOptions(serviceName string) *ServerHandshakerOptions {
	return &ServerHandshakerOptions{ServiceName: serviceName}
}

// altsHandshake is used to complete an ALTS handshake between client and
// server. This handshaker talks to the ALTS handshaker service in the metadata
// server.
type altsHandshake struct {
	// RPC altsServiceStream used to access the ALTS Handshaker service.
	altsServiceStream altsproto.HandshakerService_DoHandshakeClient
	// the connection to the peer.
	targetGRPConn net.Conn
	// a virtual connection to the ALTS handshaker service.
	altsHandshakeServiceConn *grpc.ClientConn
	// client handshake options.
	clientHandshakeOptions *ClientHandshakerOptions
	// server handshake options.
	serverHandshakeOptions *ServerHandshakerOptions
	// defines the side doing the handshake, client or server.
	side int
}

func NewClientHandshake(ctx context.Context, conn *grpc.ClientConn, c net.Conn, opts *ClientHandshakerOptions) (internal.Handshaker, error) {
	return &altsHandshake{
		altsServiceStream:        nil,
		targetGRPConn:            c,
		altsHandshakeServiceConn: conn,
		clientHandshakeOptions:   opts,
		side:                     internal.ClientSide,
	}, nil
}

func NewServerHandshake(ctx context.Context, conn *grpc.ClientConn, c net.Conn, opts *ServerHandshakerOptions) (internal.Handshaker, error) {
	return &altsHandshake{
		altsServiceStream:        nil,
		targetGRPConn:            c,
		altsHandshakeServiceConn: conn,
		serverHandshakeOptions:   opts,
		side:                     internal.ServerSide,
	}, nil
}

func (h *altsHandshake) ClientHandshake(ctx context.Context) (net.Conn, credentials.AuthInfo, error) {
	if err := clientHandshakes.Acquire(ctx, 1); err != nil {
		log.Error("握手头读取失败", zap.Error(err))
		return nil, nil, err
	}
	defer clientHandshakes.Release(1)

	if h.side != internal.ClientSide {
		return nil, nil, errors.New("only handshakers created using NewClientHandshake can perform a client handshaker")
	}
	if h.altsServiceStream == nil {
		// 创建到认知服务器的链接
		stream, err := altsproto.NewHandshakerServiceClient(h.altsHandshakeServiceConn).DoHandshake(ctx)
		if err != nil {
			log.Error("创建握手流失败", zap.Error(err))
			return nil, nil, fmt.Errorf("failed to establish altsServiceStream to ALTS handshaker service: %v", err)
		}
		h.altsServiceStream = stream
	}
	// Create target identities from service account list.
	targetIdentities := make([]*altsproto.Identity, 0, len(h.clientHandshakeOptions.TargetServiceAccounts))
	for _, account := range h.clientHandshakeOptions.TargetServiceAccounts {
		targetIdentities = append(targetIdentities, &altsproto.Identity{
			IdentityOneof: &altsproto.Identity_ServiceAccount{
				ServiceAccount: account,
			},
		})
	}
	req := &altsproto.HandshakerReq{
		ReqOneof: &altsproto.HandshakerReq_ClientStart{
			ClientStart: &altsproto.StartClientHandshakeReq{
				HandshakeSecurityProtocol: hsProtocol,
				ApplicationProtocols:      appProtocols,
				RecordProtocols:           recordProtocols,
				TargetIdentities:          targetIdentities,
				LocalIdentity:             h.clientHandshakeOptions.ClientIdentity,
				TargetName:                h.clientHandshakeOptions.TargetName,
				RpcVersions:               h.clientHandshakeOptions.RPCVersions,
			},
		},
	}
	conn, result, err := h.doHandshake(req)
	if err != nil {
		log.Error("实际握手失败", zap.Error(err))
		return nil, nil, err
	}
	authInfo := NewAuthInfo(result)
	return conn, authInfo, nil
}

// ServerHandshake starts and completes a server ALTS handshake for GCP. Once
// done, ServerHandshake returns a secure connection.
func (h *altsHandshake) ServerHandshake(ctx context.Context) (net.Conn, credentials.AuthInfo, error) {
	if err := serverHandshakes.Acquire(ctx, 1); err != nil {
		return nil, nil, err
	}
	defer serverHandshakes.Release(1)
	if h.side != internal.ServerSide {
		return nil, nil, errors.New("only handshakers created using NewServerHandshake can perform a server handshaker")
	}
	// TODO(matthewstevenson88): Change unit tests to use public APIs so
	// that h.altsServiceStream can unconditionally be set based on h.clientConn.
	if h.altsServiceStream == nil {
		// 创建到认证服务器的链接
		stream, err := altsproto.NewHandshakerServiceClient(h.altsHandshakeServiceConn).DoHandshake(ctx)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to establish altsServiceStream to ALTS handshaker service: %v", err)
		}
		h.altsServiceStream = stream
	}

	p := make([]byte, frameLimit)
	n, err := h.targetGRPConn.Read(p)
	if err != nil {
		return nil, nil, err
	}
	localIdentities := make([]*altsproto.Identity, 0, len(h.serverHandshakeOptions.TargetServiceAccounts))
	for _, account := range h.serverHandshakeOptions.TargetServiceAccounts {
		localIdentities = append(localIdentities, &altsproto.Identity{
			IdentityOneof: &altsproto.Identity_ServiceAccount{
				ServiceAccount: account,
			},
			ServiceName: h.serverHandshakeOptions.ServiceName,
		})
	}
	// Prepare server parameters.
	params := make(map[int32]*altsproto.ServerHandshakeParameters)
	params[int32(altsproto.HandshakeProtocol_ALTS)] = &altsproto.ServerHandshakeParameters{
		RecordProtocols: recordProtocols,
		LocalIdentities: localIdentities,
	}
	req := &altsproto.HandshakerReq{
		ReqOneof: &altsproto.HandshakerReq_ServerStart{
			ServerStart: &altsproto.StartServerHandshakeReq{
				ApplicationProtocols: appProtocols,
				HandshakeParameters:  params,
				InBytes:              p[:n],
				RpcVersions:          h.serverHandshakeOptions.RPCVersions,
			},
		},
	}
	conn, result, err := h.doHandshake(req)
	if err != nil {
		return nil, nil, err
	}
	authInfo := NewAuthInfo(result)
	return conn, authInfo, nil
}

func (h *altsHandshake) doHandshake(req *altsproto.HandshakerReq) (net.Conn, *altsproto.HandshakerResult, error) {
	resp, err := h.accessHandshakeService(req)
	if err != nil {
		return nil, nil, err
	}
	// Check of the returned status is an error.
	if resp.GetStatus() != nil {
		if got, want := resp.GetStatus().Code, uint32(codes.OK); got != want {
			return nil, nil, fmt.Errorf("%v", resp.GetStatus().Details)
		}
	}
	var extra []byte
	if req.GetServerStart() != nil {
		if resp.GetBytesConsumed() > uint32(len(req.GetServerStart().GetInBytes())) {
			return nil, nil, errOutOfBound
		}
		extra = req.GetServerStart().GetInBytes()[resp.GetBytesConsumed():]
	}
	result, extra, err := h.processUntilDone(resp, extra)
	if err != nil {
		log.Error("", zap.Any("side", h.side), zap.Any("RemoteAddr", h.targetGRPConn.RemoteAddr()), zap.Error(err))
		return nil, nil, err
	}
	// The handshaker returns a 128 bytes key. It should be truncated based
	// on the returned record protocol.
	keyLen, ok := keyLength[result.RecordProtocol]
	if !ok {
		return nil, nil, fmt.Errorf("unknown resulted record protocol %v", result.RecordProtocol)
	}
	sc, err := conn.NewConn(h.targetGRPConn, h.side, result.GetRecordProtocol(), result.KeyData[:keyLen], extra)
	if err != nil {
		return nil, nil, err
	}
	return sc, result, nil
}

// 向认证服务器发出请求并获得返回信息
func (h *altsHandshake) accessHandshakeService(req *altsproto.HandshakerReq) (*altsproto.HandshakerResp, error) {
	if err := h.altsServiceStream.Send(req); err != nil {
		return nil, err
	}
	resp, err := h.altsServiceStream.Recv()
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// processUntilDone processes the handshake until the handshaker service returns
// the results. Handshaker service takes care of frame parsing, so we read
// whatever received from the network and send it to the handshaker service.
func (h *altsHandshake) processUntilDone(resp *altsproto.HandshakerResp, extra []byte) (*altsproto.HandshakerResult, []byte, error) {
	var lastWriteTime time.Time
	for {
		if len(resp.OutFrames) > 0 {
			lastWriteTime = time.Now()
			if _, err := h.targetGRPConn.Write(resp.OutFrames); err != nil {
				return nil, nil, err
			}
		}
		if resp.Result != nil {
			return resp.Result, extra, nil
		}
		buf := make([]byte, frameLimit)
		n, err := h.targetGRPConn.Read(buf)
		if err != nil && err != io.EOF {
			return nil, nil, err
		}
		//log.Debug("Read frameLimit", zap.Error(err))
		// If there is nothing to send to the handshaker service, and
		// nothing is received from the peer, then we are stuck.
		// This covers the case when the peer is not responding. Note
		// that handshaker service connection issues are caught in
		// accessHandshakeService before we even get here.
		if len(resp.OutFrames) == 0 && n == 0 {
			return nil, nil, internal.PeerNotRespondingError
		}
		// Append extra bytes from the previous interaction with the
		// handshaker service with the current buffer read from targetGRPConn.
		p := append(extra, buf[:n]...)
		// Compute the time elapsed since the last write to the peer.
		timeElapsed := time.Since(lastWriteTime)
		timeElapsedMs := uint32(timeElapsed.Milliseconds())
		// From here on, p and extra point to the same slice.
		resp, err = h.accessHandshakeService(&altsproto.HandshakerReq{
			ReqOneof: &altsproto.HandshakerReq_Next{
				Next: &altsproto.NextHandshakeMessageReq{
					InBytes:          p,
					NetworkLatencyMs: timeElapsedMs,
				},
			},
		})
		if err != nil {
			return nil, nil, err
		}
		// Set extra based on handshaker service response.
		if resp.GetBytesConsumed() > uint32(len(p)) {
			return nil, nil, errOutOfBound
		}
		extra = p[resp.GetBytesConsumed():]
	}
}

// Close terminates the Handshaker. It should be called when the caller obtains
// the secure connection.
func (h *altsHandshake) Close() {
	if h.altsServiceStream != nil {
		h.altsServiceStream.CloseSend()
	}
}

// ResetConcurrentHandshakeSemaphoreForTesting resets the handshake semaphores
// to allow numberOfAllowedHandshakes concurrent handshakes each.
func ResetConcurrentHandshakeSemaphoreForTesting(numberOfAllowedHandshakes int64) {
	clientHandshakes = semaphore.NewWeighted(numberOfAllowedHandshakes)
	serverHandshakes = semaphore.NewWeighted(numberOfAllowedHandshakes)
}
