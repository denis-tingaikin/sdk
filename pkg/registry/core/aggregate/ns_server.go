// Copyright (c) 2020 Doc.ai and/or its affiliates.
//
// SPDX-License-Identifier: Apache-2.0
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package aggregate provides a possible to aggregate few stream clients to single interface
package aggregate

import (
	"context"
	"io"
	"sync"
	"sync/atomic"

	"github.com/networkservicemesh/api/pkg/api/registry"
	"google.golang.org/grpc"
)

type nsAggregateClient struct {
	grpc.ClientStream
	ctx     context.Context
	cancel  func()
	clients []registry.NetworkServiceRegistry_FindClient
	once    sync.Once
	ch      chan *registry.NetworkService
	done    *int32
}

func (c *nsAggregateClient) initMonitoring() {
	for i := 0; i < len(c.clients); i++ {
		client := c.clients[i]
		go func() {
			for ns := range registry.ReadNetworkServiceChannel(client) {
				c.ch <- ns
			}
			if atomic.AddInt32(c.done, 1) == int32(len(c.clients)) {
				c.cancel()
			}
		}()
	}
	if len(c.clients) == 0 {
		c.cancel()
	}
}

func (c *nsAggregateClient) Recv() (*registry.NetworkService, error) {
	c.once.Do(c.initMonitoring)
	select {
	case <-c.ctx.Done():
		return nil, io.EOF
	case v, ok := <-c.ch:
		if !ok {
			return nil, io.EOF
		}
		return v, nil
	}
}

func (c *nsAggregateClient) Context() context.Context {
	return c.ctx
}

// NewNetworkServiceFindClient aggregates few NetworkServiceRegistry_FindClient to single  NetworkServiceRegistry_FindClient
func NewNetworkServiceFindClient(clients ...registry.NetworkServiceRegistry_FindClient) registry.NetworkServiceRegistry_FindClient {
	d := int32(0)
	r := &nsAggregateClient{
		clients: filterNetworkServiceClients(clients),
		ch:      make(chan *registry.NetworkService),
		done:    &d,
	}
	r.ctx, r.cancel = context.WithCancel(context.Background())
	return r
}

func filterNetworkServiceClients(clients []registry.NetworkServiceRegistry_FindClient) []registry.NetworkServiceRegistry_FindClient {
	var result []registry.NetworkServiceRegistry_FindClient
	for _, c := range clients {
		if c == nil {
			continue
		}
		result = append(result, c)
	}
	return result
}