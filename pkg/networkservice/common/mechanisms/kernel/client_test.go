// Copyright (c) 2021 Doc.ai and/or its affiliates.
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

package kernel_test

import (
	"context"
	"testing"

	"github.com/networkservicemesh/api/pkg/api/networkservice"
	"github.com/stretchr/testify/require"

	"github.com/networkservicemesh/sdk/pkg/networkservice/common/mechanisms/kernel"
)

func Test_KernelClient_ShouldNotDoublingMechanims(t *testing.T) {
	c := kernel.NewClient()

	req := &networkservice.NetworkServiceRequest{}

	for i := 0; i < 10; i++ {
		_, err := c.Request(context.Background(), req)
		require.NoError(t, err)
		require.Len(t, req.MechanismPreferences, 1)
	}
}
