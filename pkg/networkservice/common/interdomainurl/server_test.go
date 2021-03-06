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

package interdomainurl_test

import (
	"context"
	"fmt"
	"net/url"
	"testing"

	"github.com/networkservicemesh/api/pkg/api/networkservice"
	"github.com/stretchr/testify/require"

	"github.com/networkservicemesh/sdk/pkg/networkservice/common/interdomainurl"
	"github.com/networkservicemesh/sdk/pkg/networkservice/core/next"
	"github.com/networkservicemesh/sdk/pkg/networkservice/utils/checks/checkcontext"
	"github.com/networkservicemesh/sdk/pkg/networkservice/utils/checks/checkrequest"
	"github.com/networkservicemesh/sdk/pkg/tools/clienturlctx"
)

const (
	nseName   = "nse1"
	domainURL = "tcp://127.0.0.1:5000"
)

func TestInterdomainURLServer_Register(t *testing.T) {
	expected, err := url.Parse(domainURL)
	require.NoError(t, err)

	s := next.NewNetworkServiceServer(
		interdomainurl.NewServer(),
		checkcontext.NewServer(t, func(t *testing.T, ctx context.Context) {
			require.Equal(t, *expected, *clienturlctx.ClientURL(ctx))
		}),
		checkrequest.NewServer(t, func(t *testing.T, request *networkservice.NetworkServiceRequest) {
			require.Equal(t, nseName, request.Connection.NetworkServiceEndpointName)
		}),
	)

	conn, err := s.Request(context.Background(), &networkservice.NetworkServiceRequest{
		Connection: &networkservice.Connection{
			NetworkServiceEndpointName: fmt.Sprintf("%s@%s", nseName, domainURL),
		},
	})
	require.NoError(t, err)
	require.NotNil(t, conn)

	require.Equal(t, fmt.Sprintf("%s@%s", nseName, domainURL), conn.NetworkServiceEndpointName)
}
