// Copyright (c) 2020-2021 Cisco Systems, Inc.
//
// Copyright (c) 2020-2021 Doc.ai and/or its affiliates.
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

// Package endpoint provides a simple wrapper for building a NetworkServiceServer
package endpoint

import (
	"context"
	"net/url"

	"google.golang.org/grpc"

	"github.com/google/uuid"
	"github.com/networkservicemesh/api/pkg/api/networkservice"

	"github.com/networkservicemesh/sdk/pkg/networkservice/common/authorize"
	"github.com/networkservicemesh/sdk/pkg/networkservice/common/serialize"
	"github.com/networkservicemesh/sdk/pkg/networkservice/common/updatetoken"
	"github.com/networkservicemesh/sdk/pkg/networkservice/utils/metadata"
	"github.com/networkservicemesh/sdk/pkg/tools/grpcutils"

	"github.com/networkservicemesh/sdk/pkg/networkservice/common/monitor"
	"github.com/networkservicemesh/sdk/pkg/networkservice/common/timeout"
	"github.com/networkservicemesh/sdk/pkg/networkservice/common/updatepath"
	"github.com/networkservicemesh/sdk/pkg/networkservice/core/chain"
	"github.com/networkservicemesh/sdk/pkg/tools/token"
)

// Endpoint - aggregates the APIs:
//            - networkservice.NetworkServiceServer
//            - networkservice.MonitorConnectionServer
type Endpoint interface {
	networkservice.NetworkServiceServer
	networkservice.MonitorConnectionServer
	// Register - register the endpoint with *grpc.Server s
	Register(s *grpc.Server)
}

type endpoint struct {
	networkservice.NetworkServiceServer
	networkservice.MonitorConnectionServer
}

type serverOptions struct {
	name                    string
	authorizeServer         networkservice.NetworkServiceServer
	additionalFunctionality []networkservice.NetworkServiceServer
}

// Option modifies server option value
type Option func(o *serverOptions)

// WithName sets name of the NetworkServiceServer
func WithName(name string) Option {
	return func(o *serverOptions) {
		o.name = name
	}
}

// WithAuthorizeServer sets authorization server chain element
func WithAuthorizeServer(authorizeServer networkservice.NetworkServiceServer) Option {
	if authorizeServer == nil {
		panic("Authorize server cannot be nil")
	}
	return func(o *serverOptions) {
		o.authorizeServer = authorizeServer
	}
}

// WithAdditionalFunctionality sets additional NetworkServiceServer chain elements to be included in the chain
func WithAdditionalFunctionality(additionalFunctionality ...networkservice.NetworkServiceServer) Option {
	return func(o *serverOptions) {
		o.additionalFunctionality = additionalFunctionality
	}
}

// NewServer - returns a NetworkServiceMesh client as a chain of the standard Client pieces plus whatever
func NewServer(ctx context.Context, tokenGenerator token.GeneratorFunc, options ...Option) Endpoint {
	opts := &serverOptions{
		name:            "endpoint-" + uuid.New().String(),
		authorizeServer: authorize.NewServer(authorize.Any()),
	}
	for _, opt := range options {
		opt(opts)
	}

	rv := &endpoint{}
	rv.NetworkServiceServer = chain.NewNamedNetworkServiceServer(
		opts.name,
		append([]networkservice.NetworkServiceServer{
			updatepath.NewServer(opts.name),
			opts.authorizeServer,
			serialize.NewServer(),
			// `timeout` uses ctx as a context for the timeout Close and it closes only the subsequent chain, so
			// chain elements before the `timeout` in chain shouldn't make any updates to the Close context and
			// shouldn't be closed on Connection Close.
			timeout.NewServer(ctx),
			metadata.NewServer(),
			monitor.NewServer(ctx, &rv.MonitorConnectionServer),
			updatetoken.NewServer(tokenGenerator),
		}, opts.additionalFunctionality...)...)
	return rv
}

func (e *endpoint) Register(s *grpc.Server) {
	grpcutils.RegisterHealthServices(s, e)
	networkservice.RegisterNetworkServiceServer(s, e)
	networkservice.RegisterMonitorConnectionServer(s, e)
}

// Serve  - serves passed Endpoint on grpc
func Serve(ctx context.Context, listenOn *url.URL, endpoint Endpoint, opt ...grpc.ServerOption) <-chan error {
	server := grpc.NewServer(opt...)
	endpoint.Register(server)

	return grpcutils.ListenAndServe(ctx, listenOn, server)
}
