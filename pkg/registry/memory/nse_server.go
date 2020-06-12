package memory

import (
	"context"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/networkservicemesh/api/pkg/api/registry"
	"github.com/networkservicemesh/sdk/pkg/registry/core/next"
	"github.com/networkservicemesh/sdk/pkg/tools/matchutils"
	"github.com/networkservicemesh/sdk/pkg/tools/serialize"
	"sync"
)

type networkServiceEndpointRegistryServer struct {
	networkServices     NetworkServiceEndpointSyncMap
	executor            serialize.Executor
	eventChannels       []chan *registry.NetworkServiceEndpoint
	eventChannelSize    int
	eventChannelsLocker sync.Mutex
}

func (n networkServiceEndpointRegistryServer) Register(ctx context.Context, ns *registry.NetworkServiceEndpoint) (*registry.NetworkServiceEndpoint, error) {
	n.networkServices.Store(ns.Name, ns)
	n.executor.AsyncExec(func() {
		n.eventChannelsLocker.Lock()
		for _, ch := range n.eventChannels {
			ch <- ns
		}
		n.eventChannelsLocker.Unlock()
	})
	return next.NetworkServiceEndpointRegistryServer(ctx).Register(ctx, ns)
}

func (n networkServiceEndpointRegistryServer) Find(query *registry.NetworkServiceEndpointQuery, s registry.NetworkServiceEndpointRegistry_FindServer) error {
	sendAllMatches := func(ns *registry.NetworkServiceEndpoint) error {
		var err error
		n.networkServices.Range(func(key string, value *registry.NetworkServiceEndpoint) bool {
			if matchutils.MatchNetworkServiceEndpoints(query.NetworkServiceEndpoint, value) {
				err = s.Send(value)
				return err == nil
			}
			return true
		})
		return err
	}
	if query.Watch {
		go func() {
			eventCh := make(chan *registry.NetworkServiceEndpoint, n.eventChannelSize)
			n.eventChannelsLocker.Lock()
			n.eventChannels = append(n.eventChannels, eventCh)
			n.eventChannelsLocker.Unlock()
			err := sendAllMatches(query.NetworkServiceEndpoint)
			if err != nil {
				return
			}
			for event := range eventCh {
				if s.Context().Err() != nil {
					return
				}
				if err := s.Send(event); err != nil {
					return
				}
				if matchutils.MatchNetworkServiceEndpoints(query.NetworkServiceEndpoint, event) {
					if err := s.Send(event); err != nil {
						return
					}
				}
			}
		}()
	} else {
		if err := sendAllMatches(query.NetworkServiceEndpoint); err != nil {
			return err
		}
	}
	return next.NetworkServiceEndpointRegistryServer(s.Context()).Find(query, s)
}

func (n networkServiceEndpointRegistryServer) Unregister(ctx context.Context, nse *registry.NetworkServiceEndpoint) (*empty.Empty, error) {
	nse.ExpirationTime.Seconds = 0
	n.networkServices.Delete(nse.Name)
	n.executor.AsyncExec(func() {
		n.eventChannelsLocker.Lock()
		for _, ch := range n.eventChannels {
			ch <- nse
		}
		n.eventChannelsLocker.Unlock()
	})
	return next.NetworkServiceEndpointRegistryServer(ctx).Unregister(ctx, nse)
}

func NewNetworkServiceEndpointRegistryServer() registry.NetworkServiceEndpointRegistryServer {
	return &networkServiceEndpointRegistryServer{}
}
