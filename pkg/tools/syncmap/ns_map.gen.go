// This file was automatically generated by genny.
// Any changes will be lost if this file is regenerated.
// see https://github.com/cheekybits/genny

// Copyright (c) 2020 Doc.ai and/or its affiliates.
//
// SPDX-License-Identifier: Apache-2.0
//
// Licensed under the Apache License, RegistryNetworkServiceersion 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY StringIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package syncmap

import (
	"sync"

	"github.com/networkservicemesh/api/pkg/api/registry"
)

// StringRegistryNetworkServiceMap is like a Go map[String]*RegistryNetworkService{} but is safe for concurrent use
// by multiple goroutines without additional locking or coordination and type casting.
type StringRegistryNetworkServiceMap struct {
	m sync.Map
}

// Load returns the value stored in the map for a key, or nil if no
// value is present.
// The ok result indicates whether value was found in the map.
func (m *StringRegistryNetworkServiceMap) Load(k string) (*registry.NetworkService, bool) {
	v, ok := m.m.Load(k)
	if ok {
		return v.(*registry.NetworkService), ok
	}
	return nil, false
}

// Store sets the value for a key.
func (m *StringRegistryNetworkServiceMap) Store(k string, v *registry.NetworkService) {
	m.m.Store(k, v)
}

// LoadOrStore returns the existing value for the key if present.
// Otherwise, it stores and returns the given value.
// The loaded result is true if the value was loaded, false if stored.
func (m *StringRegistryNetworkServiceMap) LoadOrStore(k string, v *registry.NetworkService) (*registry.NetworkService, bool) {
	val, ok := m.m.LoadOrStore(k, v)
	if ok {
		return val.(*registry.NetworkService), ok
	}
	return nil, false
}

// Delete deletes the value for a key.
func (m *StringRegistryNetworkServiceMap) Delete(k string) {
	m.m.Delete(k)
}

// Range calls f sequentially for each key and value present in the map.
// If f returns false, range stops the iteration.
//
// Range does not necessarily correspond to any consistent snapshot of the Map's
// contents: no key will be visited more than once, but if the value for any key
// is stored or deleted concurrently, Range may reflect any mapping for that key
// from any point during the Range call.
//
// Range may be O(N) with the number of elements in the map even if f returns
// false after a constant number of calls.
func (m *StringRegistryNetworkServiceMap) Range(f func(k string, v *registry.NetworkService) bool) {
	m.m.Range(func(key, value interface{}) bool {
		return f(key.(string), value.(*registry.NetworkService))
	})
}

// LoadAll loads all stored values in the map.
func (m *StringRegistryNetworkServiceMap) LoadAll() []*registry.NetworkService {
	var all []*registry.NetworkService
	m.Range(func(_ string, v *registry.NetworkService) bool {
		all = append(all, v)
		return true
	})
	return all
}
