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

package opa_test

import (
	"context"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/networkservicemesh/api/pkg/api/networkservice"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/networkservicemesh/sdk/pkg/tools/opa"
)

func genJWTWithClaimsWithYear(year int) string {
	return genJWTWithClaims(&jwt.StandardClaims{
		ExpiresAt: time.Date(year, 1, 1, 1, 1, 1, 1, time.UTC).Unix(),
	})
}

func genJWTWithClaims(claims *jwt.StandardClaims) string {
	token, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte("super secret"))
	return token
}

func genConnectionWithTokens(tokens []string) *networkservice.Connection {
	rv := &networkservice.Connection{
		Path: &networkservice.Path{
			PathSegments: []*networkservice.PathSegment{},
		},
	}

	for _, token := range tokens {
		rv.Path.PathSegments = append(rv.Path.PathSegments, &networkservice.PathSegment{
			Token: token,
		})
	}

	return rv
}

func TestNoTokensExpiredPolicy(t *testing.T) {
	nextYear := time.Now().Year() + 1
	lastYear := time.Now().Year() - 1
	suits := []struct {
		name    string
		tokens  []string
		expired bool
	}{
		{
			name: "simple positive test with one token",
			tokens: []string{
				genJWTWithClaimsWithYear(nextYear),
			},
			expired: false,
		},
		{
			name: "positive test with two tokens",
			tokens: []string{
				genJWTWithClaimsWithYear(nextYear),
				genJWTWithClaimsWithYear(nextYear),
			},
			expired: false,
		},
		{
			name: "negative test with expired/not expired tokens",
			tokens: []string{
				genJWTWithClaimsWithYear(nextYear),
				genJWTWithClaimsWithYear(lastYear),
			},
			expired: true,
		},
	}

	p := opa.WithTokensExpiredPolicy()

	for i := range suits {
		s := suits[i]

		t.Run(s.name, func(t *testing.T) {
			conn := genConnectionWithTokens(s.tokens)
			checkResult := func(err error) {
				if s.expired {
					require.NotNil(t, err)
					s, ok := status.FromError(err)
					require.True(t, ok, "error without error status code")
					require.Equal(t, s.Code(), codes.PermissionDenied, "wrong error status code")
					return
				}
				require.Nil(t, err)
			}
			checkResult(p.Check(context.Background(), conn.GetPath()))
		})
	}
}
