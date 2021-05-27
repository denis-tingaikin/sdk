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

package nsmgrproxy_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/networkservicemesh/api/pkg/api/networkservice"
	"github.com/networkservicemesh/api/pkg/api/networkservice/mechanisms/cls"
	"github.com/networkservicemesh/api/pkg/api/networkservice/mechanisms/kernel"
	"github.com/networkservicemesh/api/pkg/api/registry"

	"github.com/networkservicemesh/sdk/pkg/tools/sandbox"
)

func TestNSMGR_InterdomainUseCase(t *testing.T) {
	//	t.Cleanup(func() { goleak.VerifyNone(t) })

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	// ctx = log.WithLog(ctx, logruslogger.New(ctx))
	// log.EnableTracing(true)

	var dnsServer = new(sandbox.FakeDNSResolver)

	cluster1 := sandbox.NewBuilder(t).
		SetNodesCount(1).
		SetContext(ctx).
		SetDNSResolver(dnsServer).
		SetDNSDomainName("cluster1").
		Build()

	cluster2 := sandbox.NewBuilder(t).
		SetNodesCount(1).
		SetDNSDomainName("cluster2").
		SetDNSResolver(dnsServer).
		SetContext(ctx).
		Build()

	nsReg := &registry.NetworkService{
		Name: "my-service-interdomain",
	}

	_, err := cluster2.Nodes[0].NSRegistryClient.Register(ctx, nsReg)
	require.NoError(t, err)

	nseReg := &registry.NetworkServiceEndpoint{
		Name:                "final-endpoint",
		NetworkServiceNames: []string{nsReg.Name},
	}

	_, err = cluster2.Nodes[0].NewEndpoint(ctx, nseReg, sandbox.GenerateTestToken)
	require.NoError(t, err)

	nsc := cluster1.Nodes[0].NewClient(ctx, sandbox.GenerateTestToken)

	request := &networkservice.NetworkServiceRequest{
		MechanismPreferences: []*networkservice.Mechanism{
			{Cls: cls.LOCAL, Type: kernel.MECHANISM},
		},
		Connection: &networkservice.Connection{
			Id:             "1",
			NetworkService: fmt.Sprint(nsReg.Name, "@", cluster2.Name),
			Context:        &networkservice.ConnectionContext{},
		},
	}

	conn, err := nsc.Request(ctx, request)
	require.NoError(t, err)
	require.NotNil(t, conn)

	require.Equal(t, 10, len(conn.Path.PathSegments))

	// Simulate refresh from client.

	refreshRequest := request.Clone()
	refreshRequest.Connection = conn.Clone()

	conn, err = nsc.Request(ctx, refreshRequest)
	require.NoError(t, err)
	require.NotNil(t, conn)
	require.Equal(t, 10, len(conn.Path.PathSegments))
}

func TestNSMGR_FloatingInterdomainUseCase(t *testing.T) {
	//t.Cleanup(func() { goleak.VerifyNone(t) })

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	var dnsServer = new(sandbox.FakeDNSResolver)

	cluster1 := sandbox.NewBuilder(t).
		SetNodesCount(1).
		SetContext(ctx).
		SetDNSResolver(dnsServer).
		SetDNSDomainName("cluster1").
		Build()

	cluster2 := sandbox.NewBuilder(t).
		SetNodesCount(1).
		SetDNSDomainName("cluster2").
		SetDNSResolver(dnsServer).
		SetContext(ctx).
		Build()

	floating := sandbox.NewBuilder(t).
		SetNodesCount(1).
		SetDNSDomainName("floating.domain").
		SetDNSResolver(dnsServer).
		SetNSMgrSupplier(nil).
		SetNSMgrProxySupplier(nil).
		SetRegistryProxySupplier(nil).
		SetNodeSetup(nil).
		SetContext(ctx).
		Build()

	nsReg := &registry.NetworkService{
		Name: "my-service-interdomain@" + floating.Name,
	}

	_, err := cluster2.Nodes[0].NSRegistryClient.Register(ctx, nsReg)
	require.NoError(t, err)

	nseReg := &registry.NetworkServiceEndpoint{
		Name:                "final-endpoint@" + floating.Name,
		NetworkServiceNames: []string{"my-service-interdomain"},
	}

	_, err = cluster2.Nodes[0].NewEndpoint(ctx, nseReg, sandbox.GenerateTestToken)
	require.NoError(t, err)

	nsc := cluster1.Nodes[0].NewClient(ctx, sandbox.GenerateTestToken)

	request := &networkservice.NetworkServiceRequest{
		MechanismPreferences: []*networkservice.Mechanism{
			{Cls: cls.LOCAL, Type: kernel.MECHANISM},
		},
		Connection: &networkservice.Connection{
			Id:             "1",
			NetworkService: fmt.Sprint(nsReg.Name),
			Context:        &networkservice.ConnectionContext{},
		},
	}

	conn, err := nsc.Request(ctx, request)
	require.NoError(t, err)
	require.NotNil(t, conn)

	require.Equal(t, 10, len(conn.Path.PathSegments))

	// Simulate refresh from client.

	refreshRequest := request.Clone()
	refreshRequest.Connection = conn.Clone()

	conn, err = nsc.Request(ctx, refreshRequest)
	require.NoError(t, err)
	require.NotNil(t, conn)
	require.Equal(t, 10, len(conn.Path.PathSegments))
}
