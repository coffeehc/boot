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
	clientHandshakes = semaphore.NewWeighted(100)
	serverHandshakes = semaphore.NewWeighted(100)

	// errOutOfBound occurs when the handshake service returns a consumed
	// bytes value larger than the buffer that was passed to it originally.
	errOutOfBound = errors.New("handshaker service consumed bytes value is out-of-bound")
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

// DefaultClientHandshakerOptions returns the default client handshaker options.
func DefaultClientHandshakerOptions(serviceName string) *ClientHandshakerOptions {
	return &ClientHandshakerOptions{
		ServiceName: serviceName,
	}
}

// DefaultServerHandshakerOptions returns the default client handshaker options.
func DefaultServerHandshakerOptions(serviceName string) *ServerHandshakerOptions {
	return &ServerHandshakerOptions{ServiceName: serviceName}
}

// altsHandshaker is used to complete an ALTS handshake between client and
// server. This handshaker talks to the ALTS handshaker service in the metadata
// server.
type altsHandshaker struct {
	// RPC stream used to access the ALTS Handshaker service.
	stream altsproto.HandshakerService_DoHandshakeClient
	// the connection to the peer.
	conn net.Conn
	// a virtual connection to the ALTS handshaker service.
	clientConn *grpc.ClientConn
	// client handshake options.
	clientOpts *ClientHandshakerOptions
	// server handshake options.
	serverOpts *ServerHandshakerOptions
	// defines the side doing the handshake, client or server.
	side int
}

func NewClientHandshaker(ctx context.Context, conn *grpc.ClientConn, c net.Conn, opts *ClientHandshakerOptions) (internal.Handshaker, error) {
	return &altsHandshaker{
		stream:     nil,
		conn:       c,
		clientConn: conn,
		clientOpts: opts,
		side:       internal.ClientSide,
	}, nil
}

func NewServerHandshaker(ctx context.Context, conn *grpc.ClientConn, c net.Conn, opts *ServerHandshakerOptions) (internal.Handshaker, error) {
	return &altsHandshaker{
		stream:     nil,
		conn:       c,
		clientConn: conn,
		serverOpts: opts,
		side:       internal.ServerSide,
	}, nil
}

func (h *altsHandshaker) ClientHandshake(ctx context.Context) (net.Conn, credentials.AuthInfo, error) {
	if err := clientHandshakes.Acquire(ctx, 1); err != nil {
		log.Error("握手头读取失败", zap.Error(err))
		return nil, nil, err
	}
	defer clientHandshakes.Release(1)

	if h.side != internal.ClientSide {
		return nil, nil, errors.New("only handshakers created using NewClientHandshaker can perform a client handshaker")
	}

	// TODO(matthewstevenson88): Change unit tests to use public APIs so
	// that h.stream can unconditionally be set based on h.clientConn.
	if h.stream == nil {
		stream, err := altsproto.NewHandshakerServiceClient(h.clientConn).DoHandshake(ctx)
		if err != nil {
			log.Error("创建握手流失败", zap.Error(err))
			return nil, nil, fmt.Errorf("failed to establish stream to ALTS handshaker service: %v", err)
		}
		h.stream = stream
	}

	// Create target identities from service account list.
	targetIdentities := make([]*altsproto.Identity, 0, len(h.clientOpts.TargetServiceAccounts))
	for _, account := range h.clientOpts.TargetServiceAccounts {
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
				LocalIdentity:             h.clientOpts.ClientIdentity,
				TargetName:                h.clientOpts.TargetName,
				RpcVersions:               h.clientOpts.RPCVersions,
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
func (h *altsHandshaker) ServerHandshake(ctx context.Context) (net.Conn, credentials.AuthInfo, error) {
	if err := serverHandshakes.Acquire(ctx, 1); err != nil {
		return nil, nil, err
	}
	defer serverHandshakes.Release(1)

	if h.side != internal.ServerSide {
		return nil, nil, errors.New("only handshakers created using NewServerHandshaker can perform a server handshaker")
	}

	// TODO(matthewstevenson88): Change unit tests to use public APIs so
	// that h.stream can unconditionally be set based on h.clientConn.
	if h.stream == nil {
		stream, err := altsproto.NewHandshakerServiceClient(h.clientConn).DoHandshake(ctx)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to establish stream to ALTS handshaker service: %v", err)
		}
		h.stream = stream
	}

	p := make([]byte, frameLimit)
	n, err := h.conn.Read(p)
	if err != nil {
		return nil, nil, err
	}
	localIdentities := make([]*altsproto.Identity, 0, len(h.serverOpts.TargetServiceAccounts))
	for _, account := range h.serverOpts.TargetServiceAccounts {
		localIdentities = append(localIdentities, &altsproto.Identity{
			IdentityOneof: &altsproto.Identity_ServiceAccount{
				ServiceAccount: account,
			},
			ServiceName: h.serverOpts.ServiceName,
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
				RpcVersions:          h.serverOpts.RPCVersions,
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

func (h *altsHandshaker) doHandshake(req *altsproto.HandshakerReq) (net.Conn, *altsproto.HandshakerResult, error) {
	resp, err := h.accessHandshakerService(req)
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
		return nil, nil, err
	}
	// The handshaker returns a 128 bytes key. It should be truncated based
	// on the returned record protocol.
	keyLen, ok := keyLength[result.RecordProtocol]
	if !ok {
		return nil, nil, fmt.Errorf("unknown resulted record protocol %v", result.RecordProtocol)
	}
	sc, err := conn.NewConn(h.conn, h.side, result.GetRecordProtocol(), result.KeyData[:keyLen], extra)
	if err != nil {
		return nil, nil, err
	}
	return sc, result, nil
}

func (h *altsHandshaker) accessHandshakerService(req *altsproto.HandshakerReq) (*altsproto.HandshakerResp, error) {
	if err := h.stream.Send(req); err != nil {
		return nil, err
	}
	resp, err := h.stream.Recv()
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// processUntilDone processes the handshake until the handshaker service returns
// the results. Handshaker service takes care of frame parsing, so we read
// whatever received from the network and send it to the handshaker service.
func (h *altsHandshaker) processUntilDone(resp *altsproto.HandshakerResp, extra []byte) (*altsproto.HandshakerResult, []byte, error) {
	var lastWriteTime time.Time
	for {
		if len(resp.OutFrames) > 0 {
			lastWriteTime = time.Now()
			if _, err := h.conn.Write(resp.OutFrames); err != nil {
				return nil, nil, err
			}
		}
		if resp.Result != nil {
			return resp.Result, extra, nil
		}
		buf := make([]byte, frameLimit)
		n, err := h.conn.Read(buf)
		if err != nil && err != io.EOF {
			return nil, nil, err
		}
		// If there is nothing to send to the handshaker service, and
		// nothing is received from the peer, then we are stuck.
		// This covers the case when the peer is not responding. Note
		// that handshaker service connection issues are caught in
		// accessHandshakerService before we even get here.
		if len(resp.OutFrames) == 0 && n == 0 {
			return nil, nil, internal.PeerNotRespondingError
		}
		// Append extra bytes from the previous interaction with the
		// handshaker service with the current buffer read from conn.
		p := append(extra, buf[:n]...)
		// Compute the time elapsed since the last write to the peer.
		timeElapsed := time.Since(lastWriteTime)
		timeElapsedMs := uint32(timeElapsed.Milliseconds())
		// From here on, p and extra point to the same slice.
		resp, err = h.accessHandshakerService(&altsproto.HandshakerReq{
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
func (h *altsHandshaker) Close() {
	if h.stream != nil {
		h.stream.CloseSend()
	}
}

// ResetConcurrentHandshakeSemaphoreForTesting resets the handshake semaphores
// to allow numberOfAllowedHandshakes concurrent handshakes each.
func ResetConcurrentHandshakeSemaphoreForTesting(numberOfAllowedHandshakes int64) {
	clientHandshakes = semaphore.NewWeighted(numberOfAllowedHandshakes)
	serverHandshakes = semaphore.NewWeighted(numberOfAllowedHandshakes)
}
