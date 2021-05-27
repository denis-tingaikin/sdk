package seturl

import (
	"context"
	"net/url"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/networkservicemesh/api/pkg/api/registry"
	"github.com/networkservicemesh/sdk/pkg/registry/core/next"
)

type seturlNSEServer struct {
	u *url.URL
}

func (s *setURLNSEServer) Send(nse *registry.NetworkServiceEndpoint) error {
	nse.Url = s.u.String()
	return s.NetworkServiceEndpointRegistry_FindServer.Send(nse)
}

type setURLNSEServer struct {
	u *url.URL
	registry.NetworkServiceEndpointRegistry_FindServer
}

func (n *seturlNSEServer) Register(ctx context.Context, service *registry.NetworkServiceEndpoint) (*registry.NetworkServiceEndpoint, error) {
	pop := service.Url
	service.Url = n.u.String()
	resp, err := next.NetworkServiceEndpointRegistryServer(ctx).Register(ctx, service)

	if resp != nil {
		resp.Url = pop
	}

	return resp, err
}

func (n *seturlNSEServer) Find(query *registry.NetworkServiceEndpointQuery, server registry.NetworkServiceEndpointRegistry_FindServer) error {
	return next.NetworkServiceEndpointRegistryServer(server.Context()).Find(query, &setURLNSEServer{NetworkServiceEndpointRegistry_FindServer: server, u: n.u})
}

func (n *seturlNSEServer) Unregister(ctx context.Context, service *registry.NetworkServiceEndpoint) (*empty.Empty, error) {
	service.Url = n.u.String()
	return next.NetworkServiceEndpointRegistryServer(ctx).Unregister(ctx, service)
}

func NewNetworkServiceEndpointRegistryServer(u *url.URL) registry.NetworkServiceEndpointRegistryServer {
	if u == nil {
		panic("u can not be nil")
	}
	return &seturlNSEServer{u: u}
}
