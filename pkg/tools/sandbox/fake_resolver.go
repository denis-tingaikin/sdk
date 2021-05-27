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

package sandbox

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"sync"

	"github.com/networkservicemesh/sdk/pkg/registry/common/dnsresolve"
)

// FakeDNSResolver implements dnsresolve.Resolver interface and can be used for logic DNS testing
type FakeDNSResolver struct {
	sync.Mutex
	ports       map[string]string
	hostToCNAME map[string]string
}

// LookupSRV lookups DNS SRV record
func (f *FakeDNSResolver) LookupSRV(ctx context.Context, service, proto, name string) (string, []*net.SRV, error) {
	f.Lock()
	defer f.Unlock()
	if f.ports == nil {
		f.ports = map[string]string{}
	}
	if v, ok := f.ports[name]; ok {
		i, err := strconv.Atoi(v)
		if err != nil {
			return "", nil, err
		}
		return fmt.Sprintf("_%v._%v.%v", service, proto, name), []*net.SRV{{
			Port:   uint16(i),
			Target: name,
		}}, nil
	}
	return "", nil, errors.New("not found")
}

// LookupIPAddr lookups IP address by host
func (f *FakeDNSResolver) LookupIPAddr(_ context.Context, host string) ([]net.IPAddr, error) {
	f.Lock()
	defer f.Unlock()
	if f.ports == nil {
		f.ports = map[string]string{}
	}
	if _, ok := f.ports[host]; ok {
		return []net.IPAddr{{
			IP: net.ParseIP("127.0.0.1"),
		}}, nil
	}
	return nil, errors.New("not found")
}

func (f *FakeDNSResolver) LookupCNAME(ctx context.Context, host string) (cname string, err error) {
	if v, ok := f.hostToCNAME[host]; ok {
		return v, nil
	}

	return "", errors.New("not found")
}

func (f *FakeDNSResolver) AddCNAMEEntry(host, chanme string) {
	if f.hostToCNAME == nil {
		f.hostToCNAME = make(map[string]string)
	}
	f.hostToCNAME[host] = chanme
}

// Register adds new DNS record by passed url.URL
func (f *FakeDNSResolver) AddSRVEntry(name, service string, u *url.URL) {
	if u == nil {
		panic("u cannot be nil")
	}
	f.Lock()
	defer f.Unlock()
	if f.ports == nil {
		f.ports = map[string]string{}
	}
	key := fmt.Sprintf("%v.%v", service, name)
	var err error
	if _, f.ports[key], err = net.SplitHostPort(u.Host); err != nil {
		panic(err.Error())
	}

}

var _ dnsresolve.Resolver = (*FakeDNSResolver)(nil)
