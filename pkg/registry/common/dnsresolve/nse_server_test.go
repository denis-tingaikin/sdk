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

package dnsresolve_test

import (
	"context"
	"net/url"
	"testing"
	"time"

	"github.com/networkservicemesh/sdk/pkg/tools/clienturlctx"
	"github.com/networkservicemesh/sdk/pkg/tools/sandbox"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/networkservicemesh/api/pkg/api/registry"
	"github.com/stretchr/testify/require"

	"github.com/networkservicemesh/sdk/pkg/registry/common/dnsresolve"
	"github.com/networkservicemesh/sdk/pkg/registry/core/next"
	"github.com/networkservicemesh/sdk/pkg/registry/core/streamchannel"
)

type checkNSEContext struct {
	*testing.T
	expectedURL *url.URL
}

func (c *checkNSEContext) Register(ctx context.Context, ns *registry.NetworkServiceEndpoint) (*registry.NetworkServiceEndpoint, error) {
	require.Equal(c, c.expectedURL, clienturlctx.ClientURL(ctx))
	return next.NetworkServiceEndpointRegistryServer(ctx).Register(ctx, ns)
}

func (c *checkNSEContext) Find(q *registry.NetworkServiceEndpointQuery, s registry.NetworkServiceEndpointRegistry_FindServer) error {
	require.Equal(c, c.expectedURL, clienturlctx.ClientURL(s.Context()))
	return next.NetworkServiceEndpointRegistryServer(s.Context()).Find(q, s)
}

func (c *checkNSEContext) Unregister(ctx context.Context, ns *registry.NetworkServiceEndpoint) (*empty.Empty, error) {
	require.Equal(c, c.expectedURL, clienturlctx.ClientURL(ctx))
	return next.NetworkServiceEndpointRegistryServer(ctx).Unregister(ctx, ns)
}

func Test_DNSEResolve_NSESErver(t *testing.T) {
	const srv = "service1"

	var resolver = new(sandbox.FakeDNSResolver)

	u, err := url.Parse("tcp://127.0.0.1:80")
	require.NoError(t, err)

	resolver.AddSRVEntry("domain1", srv, u)
	resolver.AddCNAMEEntry("kubernetes.default.svc", "kubernetes.default.svc.domain1")

	s := dnsresolve.NewNetworkServiceEndpointRegistryServer(
		dnsresolve.WithNSMgrProxyService(srv),
		dnsresolve.WithResolver(resolver),
	)

	s = next.NewNetworkServiceEndpointRegistryServer(s, &checkNSEContext{T: t, expectedURL: u})

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err = s.Register(ctx, &registry.NetworkServiceEndpoint{Name: "ns-1@domain1"})
	require.Nil(t, err)
	err = s.Find(&registry.NetworkServiceEndpointQuery{NetworkServiceEndpoint: &registry.NetworkServiceEndpoint{Name: "ns-1@domain1"}}, streamchannel.NewNetworkServiceEndpointFindServer((ctx), nil))
	require.Nil(t, err)
	_, err = s.Unregister(ctx, &registry.NetworkServiceEndpoint{Name: "ns-1@domain1"})
	require.Nil(t, err)
}
