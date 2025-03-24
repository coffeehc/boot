/*
 *
 * Copyright 2018 gRPC authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

// Package service manages connections between the VM application and the ALTS
// handshaker service.
package service

import (
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/keepalive"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	// mu guards altsCenterConnMap and hsDialer.
	mu sync.Mutex
	// hsConn represents a mapping from a hypervisor handshaker service address
	// to a corresponding connection to a hypervisor handshaker service
	// instance.
	altsCenterConnMap = make(map[string]*grpc.ClientConn)
)

// Dial dials the handshake service in the hypervisor. If a connection has
// already been established, this function returns it. Otherwise, a new
// connection is created.
func Dial(atlsCenterAddress string) (*grpc.ClientConn, error) {
	mu.Lock()
	defer mu.Unlock()
	altsCenterConn, ok := altsCenterConnMap[atlsCenterAddress]
	if !ok {
		// Create a new connection to the handshaker service. Note that
		// this connection stays open until the application is closed.
		var err error
		altsCenterConn, err = grpc.NewClient(atlsCenterAddress,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithConnectParams(grpc.ConnectParams{
				Backoff: backoff.Config{
					BaseDelay:  time.Millisecond * 300, // 第一次失败重试前后需等待多久
					Multiplier: 1.2,                    // 在失败的重试后乘以的倍数
					Jitter:     0.2,                    // 随机抖动因子
					MaxDelay:   time.Second * 2,        // backoff上限
				},
				MinConnectTimeout: time.Second * 60,
			}),
			grpc.WithDefaultCallOptions(
				grpc.UseCompressor("gzip"),
				grpc.WaitForReady(true),
				grpc.MaxCallRecvMsgSize(1024*1024*8),
				grpc.MaxCallSendMsgSize(1024*1024*2),
			),
			grpc.WithKeepaliveParams(keepalive.ClientParameters{
				Time:                time.Second * 120,
				Timeout:             time.Second * 60,
				PermitWithoutStream: true,
			}),
			grpc.WithUserAgent("alts client"),
			//grpc.ConnectionTimeout(time.Second*5),
		)
		if err != nil {
			return nil, err
		}
		altsCenterConnMap[atlsCenterAddress] = altsCenterConn
	}
	return altsCenterConn, nil
}
