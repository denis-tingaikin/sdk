package storeurl

import (
	"context"
	"net/url"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/networkservicemesh/api/pkg/api/registry"
	"github.com/networkservicemesh/sdk/pkg/registry/core/next"
	"github.com/networkservicemesh/sdk/pkg/tools/stringurl"
)

type storeurl struct {
	m *stringurl.Map
}

type urlstockFindServer struct {
	m *stringurl.Map
	registry.NetworkServiceEndpointRegistry_FindServer
}

func (n *storeurl) Register(ctx context.Context, service *registry.NetworkServiceEndpoint) (*registry.NetworkServiceEndpoint, error) {
	return next.NetworkServiceEndpointRegistryServer(ctx).Register(ctx, service)
}

func (n *storeurl) Find(query *registry.NetworkServiceEndpointQuery, server registry.NetworkServiceEndpointRegistry_FindServer) error {
	return next.NetworkServiceEndpointRegistryServer(server.Context()).Find(query, &urlstockFindServer{NetworkServiceEndpointRegistry_FindServer: server, m: n.m})
}

func (n *storeurl) Unregister(ctx context.Context, service *registry.NetworkServiceEndpoint) (*empty.Empty, error) {
	return next.NetworkServiceEndpointRegistryServer(ctx).Unregister(ctx, service)
}

func NewNetworkServiceEndpointRegistryServer(m *stringurl.Map) registry.NetworkServiceEndpointRegistryServer {
	if m == nil {
		panic("m can not be nil")
	}
	return &storeurl{m: m}
}

func (s *urlstockFindServer) Send(nse *registry.NetworkServiceEndpoint) error {
	u, err := url.Parse(nse.Url)
	if err != nil {
		return err
	}
	s.m.Store(nse.Name, u)
	return s.NetworkServiceEndpointRegistry_FindServer.Send(nse)
}
