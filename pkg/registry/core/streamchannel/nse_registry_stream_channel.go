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

package streamchannel

import (
	"context"
	"github.com/networkservicemesh/api/pkg/api/registry"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

func NewNetworkServiceEndpointFindClient(ctx context.Context, recvCh <-chan *registry.NetworkServiceEndpoint) registry.NetworkServiceEndpointRegistry_FindClient {
	return &networkServiceEndpointRegistryFindClient{
		ctx:    ctx,
		recvCh: recvCh,
	}
}

type networkServiceEndpointRegistryFindClient struct {
	grpc.ClientStream
	err    error
	recvCh <-chan *registry.NetworkServiceEndpoint
	ctx    context.Context
}

func (c *networkServiceEndpointRegistryFindClient) Recv() (*registry.NetworkServiceEndpoint, error) {
	res, ok := <-c.recvCh
	if !ok {
		err := errors.New("recv channel has been closed")
		if c.err == nil {
			return nil, err
		}
		return res, errors.Wrap(c.err, err.Error())
	}
	return res, c.err
}

func (c *networkServiceEndpointRegistryFindClient) Context() context.Context {
	return c.ctx
}

var _ registry.NetworkServiceEndpointRegistry_FindClient = &networkServiceEndpointRegistryFindClient{}

func NewNetworkServiceEndpointFindServer(ctx context.Context, sendCh chan<- *registry.NetworkServiceEndpoint) registry.NetworkServiceEndpointRegistry_FindServer {
	return &networkServiceEndpointRegistryFindServer{
		ctx:    ctx,
		sendCh: sendCh,
	}
}

type networkServiceEndpointRegistryFindServer struct {
	grpc.ServerStream
	ctx    context.Context
	sendCh chan<- *registry.NetworkServiceEndpoint
}

func (s *networkServiceEndpointRegistryFindServer) Send(endpoint *registry.NetworkServiceEndpoint) error {
	s.sendCh <- endpoint
	return nil
}

func (s *networkServiceEndpointRegistryFindServer) Context() context.Context {
	return s.ctx
}

var _ registry.NetworkServiceEndpointRegistry_FindServer = &networkServiceEndpointRegistryFindServer{}
