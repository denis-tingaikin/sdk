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

package sandbox

import (
	"context"
	"net"
	"net/url"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"

	registryapi "github.com/networkservicemesh/api/pkg/api/registry"

	"github.com/networkservicemesh/sdk/pkg/networkservice/chains/nsmgr"
	"github.com/networkservicemesh/sdk/pkg/networkservice/chains/nsmgrproxy"
	"github.com/networkservicemesh/sdk/pkg/networkservice/common/authorize"
	"github.com/networkservicemesh/sdk/pkg/registry/chains/client"
	"github.com/networkservicemesh/sdk/pkg/registry/chains/memory"
	"github.com/networkservicemesh/sdk/pkg/registry/chains/proxydns"
	"github.com/networkservicemesh/sdk/pkg/registry/common/dnsresolve"
	"github.com/networkservicemesh/sdk/pkg/tools/grpcutils"
	"github.com/networkservicemesh/sdk/pkg/tools/log"
	"github.com/networkservicemesh/sdk/pkg/tools/opentracing"
	"github.com/networkservicemesh/sdk/pkg/tools/token"
)

// Builder implements builder pattern for building NSM Domain
type Builder struct {
	require                *require.Assertions
	resources              []context.CancelFunc
	nodesCount             int
	nodesConfig            []*NodeConfig
	DNSDomainName          string
	Resolver               dnsresolve.Resolver
	supplyNSMgr            SupplyNSMgrFunc
	supplyNSMgrProxy       SupplyNSMgrProxyFunc
	supplyRegistry         SupplyRegistryFunc
	supplyRegistryProxy    SupplyRegistryProxyFunc
	setupNode              SetupNodeFunc
	generateTokenFunc      token.GeneratorFunc
	registryExpiryDuration time.Duration
	ctx                    context.Context
}

// NewBuilder creates new SandboxBuilder
func NewBuilder(t *testing.T) *Builder {
	return &Builder{
		nodesCount:             1,
		require:                require.New(t),
		Resolver:               net.DefaultResolver,
		supplyNSMgr:            nsmgr.NewServer,
		DNSDomainName:          "cluster.local",
		supplyRegistry:         memory.NewServer,
		supplyRegistryProxy:    proxydns.NewServer,
		supplyNSMgrProxy:       nsmgrproxy.NewServer,
		setupNode:              defaultSetupNode(t),
		generateTokenFunc:      GenerateTestToken,
		registryExpiryDuration: time.Minute,
	}
}

// Build builds Domain and Supplier
func (b *Builder) Build() *Domain {
	ctx := b.ctx
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
		b.resources = append(b.resources, cancel)
	}
	ctx = log.Join(ctx, log.Empty())

	domain := new(Domain)
	domain.NSMgrProxy = b.newNSMgrProxy(ctx)
	if domain.NSMgrProxy == nil {
		domain.RegistryProxy = b.newRegistryProxy(ctx, &url.URL{})
	} else {
		domain.RegistryProxy = b.newRegistryProxy(ctx, domain.NSMgrProxy.URL)
	}
	if domain.RegistryProxy == nil {
		domain.Registry = b.newRegistry(ctx, nil)
	} else {
		domain.Registry = b.newRegistry(ctx, domain.RegistryProxy.URL)
	}

	for i := 0; i < b.nodesCount; i++ {
		domain.Nodes = append(domain.Nodes, b.newNode(ctx, domain.Registry.URL, b.nodesConfig[i]))
	}

	domain.resources, b.resources = b.resources, nil

	return domain
}

// SetContext sets default context for all chains
func (b *Builder) SetContext(ctx context.Context) *Builder {
	b.ctx = ctx
	b.SetCustomConfig([]*NodeConfig{})
	return b
}

// SetCustomConfig sets custom configuration for nodes
func (b *Builder) SetCustomConfig(config []*NodeConfig) *Builder {
	oldConfig := b.nodesConfig
	b.nodesConfig = nil

	for i := 0; i < b.nodesCount; i++ {
		nodeConfig := &NodeConfig{}
		if i < len(config) && config[i] != nil {
			*nodeConfig = *oldConfig[i]
		}

		customConfig := &NodeConfig{}
		if i < len(config) && config[i] != nil {
			*customConfig = *config[i]
		}

		if customConfig.NsmgrCtx != nil {
			nodeConfig.NsmgrCtx = customConfig.NsmgrCtx
		} else if nodeConfig.NsmgrCtx == nil {
			nodeConfig.NsmgrCtx = b.ctx
		}

		if customConfig.NsmgrGenerateTokenFunc != nil {
			nodeConfig.NsmgrGenerateTokenFunc = customConfig.NsmgrGenerateTokenFunc
		} else if nodeConfig.NsmgrGenerateTokenFunc == nil {
			nodeConfig.NsmgrGenerateTokenFunc = b.generateTokenFunc
		}

		if customConfig.ForwarderCtx != nil {
			nodeConfig.ForwarderCtx = customConfig.ForwarderCtx
		} else if nodeConfig.ForwarderCtx == nil {
			nodeConfig.ForwarderCtx = b.ctx
		}

		if customConfig.ForwarderGenerateTokenFunc != nil {
			nodeConfig.ForwarderGenerateTokenFunc = customConfig.ForwarderGenerateTokenFunc
		} else if nodeConfig.ForwarderGenerateTokenFunc == nil {
			nodeConfig.ForwarderGenerateTokenFunc = b.generateTokenFunc
		}

		b.nodesConfig = append(b.nodesConfig, nodeConfig)
	}
	return b
}

// SetNodesCount sets nodes count
func (b *Builder) SetNodesCount(nodesCount int) *Builder {
	b.nodesCount = nodesCount
	b.SetCustomConfig([]*NodeConfig{})
	return b
}

// SetDNSResolver sets DNS resolver for proxy registries
func (b *Builder) SetDNSResolver(d dnsresolve.Resolver) *Builder {
	b.Resolver = d
	return b
}

// SetTokenGenerateFunc sets function for the token generation
func (b *Builder) SetTokenGenerateFunc(f token.GeneratorFunc) *Builder {
	b.generateTokenFunc = f
	return b
}

// SetRegistryProxySupplier replaces default memory registry supplier to custom function
func (b *Builder) SetRegistryProxySupplier(f SupplyRegistryProxyFunc) *Builder {
	b.supplyRegistryProxy = f
	return b
}

// SetRegistrySupplier replaces default memory registry supplier to custom function
func (b *Builder) SetRegistrySupplier(f SupplyRegistryFunc) *Builder {
	b.supplyRegistry = f
	return b
}

// SetDNSDomainName sets DNS domain name for the building NSM domain
func (b *Builder) SetDNSDomainName(name string) *Builder {
	b.DNSDomainName = name
	return b
}

// SetNSMgrProxySupplier replaces default nsmgr-proxy supplier to custom function
func (b *Builder) SetNSMgrProxySupplier(f SupplyNSMgrProxyFunc) *Builder {
	b.supplyNSMgrProxy = f
	return b
}

// SetNSMgrSupplier replaces default nsmgr supplier to custom function
func (b *Builder) SetNSMgrSupplier(f SupplyNSMgrFunc) *Builder {
	b.supplyNSMgr = f
	return b
}

// SetNodeSetup replaces default node setup to custom function
func (b *Builder) SetNodeSetup(f SetupNodeFunc) *Builder {
	b.setupNode = f
	return b
}

// SetRegistryExpiryDuration replaces registry expiry duration to custom
func (b *Builder) SetRegistryExpiryDuration(registryExpiryDuration time.Duration) *Builder {
	b.registryExpiryDuration = registryExpiryDuration
	return b
}

func (b *Builder) dialContext(ctx context.Context, u *url.URL) *grpc.ClientConn {
	conn, err := grpc.DialContext(ctx, grpcutils.URLToTarget(u), DefaultDialOptions(b.generateTokenFunc)...)
	b.resources = append(b.resources, func() {
		_ = conn.Close()
	})
	b.require.NoError(err, "Can not dial to", u)
	return conn
}

func (b *Builder) newNSMgrProxy(ctx context.Context) *EndpointEntry {
	if b.supplyRegistryProxy == nil {
		return nil
	}
	name := "nsmgr-proxy-" + uuid.New().String()
	mgr := b.supplyNSMgrProxy(ctx, b.generateTokenFunc,
		nsmgrproxy.WithName(name),
		nsmgrproxy.WithDialOptions(DefaultDialOptions(b.generateTokenFunc)...))
	serveURL := &url.URL{Scheme: "tcp", Host: "127.0.0.1:0"}
	serve(ctx, serveURL, mgr.Register)
	log.FromContext(ctx).Infof("%v listen on: %v", name, serveURL)
	return &EndpointEntry{
		Endpoint: mgr,
		URL:      serveURL,
	}
}

// NewNSMgr - starts new Network Service Manager
func (b *Builder) NewNSMgr(ctx context.Context, node *Node, address string, registryURL *url.URL, generateTokenFunc token.GeneratorFunc) (entry *NSMgrEntry, resources []context.CancelFunc) {
	nsmgrCtx, nsmgrCancel := context.WithCancel(ctx)
	b.resources = append(b.resources, nsmgrCancel)

	entry = b.newNSMgr(nsmgrCtx, address, registryURL, generateTokenFunc)

	b.SetupRegistryClients(nsmgrCtx, node)

	resources, b.resources = b.resources, nil
	return
}

func (b *Builder) newNSMgr(ctx context.Context, address string, registryURL *url.URL, generateTokenFunc token.GeneratorFunc) *NSMgrEntry {
	if b.supplyNSMgr == nil {
		panic("nodes without managers are not supported")
	}
	var registryCC *grpc.ClientConn
	if registryURL != nil {
		registryCC = b.dialContext(ctx, registryURL)
	}
	listener, err := net.Listen("tcp", address)
	b.require.NoError(err)
	serveURL := grpcutils.AddressToURL(listener.Addr())
	b.require.NoError(listener.Close())

	nsmgrReg := &registryapi.NetworkServiceEndpoint{
		Name: "nsmgr-" + uuid.New().String(),
		Url:  serveURL.String(),
	}

	mgr := b.supplyNSMgr(ctx, nsmgrReg, authorize.NewServer(authorize.Any()), generateTokenFunc, registryCC, DefaultDialOptions(generateTokenFunc)...)

	serve(ctx, serveURL, mgr.Register)
	log.FromContext(ctx).Infof("%v listen on: %v", nsmgrReg.Name, serveURL)
	return &NSMgrEntry{
		URL:   serveURL,
		Nsmgr: mgr,
	}
}

func serve(ctx context.Context, u *url.URL, register func(server *grpc.Server)) {
	server := grpc.NewServer(opentracing.WithTracing()...)
	register(server)
	errCh := grpcutils.ListenAndServe(ctx, u, server)
	go func() {
		select {
		case <-ctx.Done():
			log.FromContext(ctx).Infof("Stop serve: %v", u.String())
			return
		case err := <-errCh:
			if err != nil {
				log.FromContext(ctx).Fatalf("An error during serve: %v", err.Error())
			}
		}
	}()
}

func (b *Builder) newRegistryProxy(ctx context.Context, nsmgrProxyURL *url.URL) *RegistryEntry {
	if b.supplyRegistryProxy == nil {
		return nil
	}
	result := b.supplyRegistryProxy(ctx, b.Resolver, b.DNSDomainName, nsmgrProxyURL, DefaultDialOptions(b.generateTokenFunc)...)
	serveURL := &url.URL{Scheme: "tcp", Host: "127.0.0.1:0"}
	serve(ctx, serveURL, result.Register)
	log.FromContext(ctx).Infof("registry-proxy-dns listen on: %v", serveURL)
	return &RegistryEntry{
		URL:      serveURL,
		Registry: result,
	}
}

func (b *Builder) newRegistry(ctx context.Context, proxyRegistryURL *url.URL) *RegistryEntry {
	if b.supplyRegistry == nil {
		return nil
	}
	result := b.supplyRegistry(ctx, b.registryExpiryDuration, proxyRegistryURL, DefaultDialOptions(b.generateTokenFunc)...)
	serveURL := &url.URL{Scheme: "tcp", Host: "127.0.0.1:0"}
	serve(ctx, serveURL, result.Register)
	log.FromContext(ctx).Infof("Registry listen on: %v", serveURL)
	return &RegistryEntry{
		URL:      serveURL,
		Registry: result,
	}
}

func (b *Builder) newNode(ctx context.Context, registryURL *url.URL, nodeConfig *NodeConfig) *Node {
	nsmgrEntry := b.newNSMgr(nodeConfig.NsmgrCtx, "127.0.0.1:0", registryURL, nodeConfig.NsmgrGenerateTokenFunc)

	node := &Node{
		ctx:   b.ctx,
		NSMgr: nsmgrEntry,
	}

	b.SetupRegistryClients(nodeConfig.NsmgrCtx, node)

	if b.setupNode != nil {
		b.setupNode(ctx, node, nodeConfig)
	}

	return node
}

// SetupRegistryClients - creates Network Service Registry Clients
func (b *Builder) SetupRegistryClients(ctx context.Context, node *Node) {
	nsmgrCC := b.dialContext(ctx, node.NSMgr.URL)

	node.ForwarderRegistryClient = client.NewNetworkServiceEndpointRegistryInterposeClient(ctx, nsmgrCC)
	node.EndpointRegistryClient = client.NewNetworkServiceEndpointRegistryClient(ctx, nsmgrCC)
	node.NSRegistryClient = client.NewNetworkServiceRegistryClient(nsmgrCC)
}

func defaultSetupNode(t *testing.T) SetupNodeFunc {
	return func(ctx context.Context, node *Node, nodeConfig *NodeConfig) {
		nseReg := &registryapi.NetworkServiceEndpoint{
			Name: "forwarder-" + uuid.New().String(),
		}
		_, err := node.NewForwarder(nodeConfig.ForwarderCtx, nseReg, nodeConfig.ForwarderGenerateTokenFunc)
		require.NoError(t, err)
	}
}
