// Copyright 2021 tsuru authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package globodns_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	globodns "github.com/tsuru/go-globodnsclient"
)

func TestClient_DomainList(t *testing.T) {
	var count int

	tests := map[string]struct {
		handler       http.HandlerFunc
		parameters    globodns.ListDomainRequest
		expected      []globodns.Domain
		expectedError string
	}{
		"domain per page < 0": {
			parameters: globodns.ListDomainRequest{
				DomainsPerPage: -1,
			},
			expectedError: "globodns: domains per page cannot be negative",
		},

		"page < 0": {
			parameters: globodns.ListDomainRequest{
				Page: -100,
			},
			expectedError: "globodns: page cannot be negative",
		},

		"servers returns an error": {
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "some internal error")
			},
			expectedError: "globodns: unexpected HTTP status code: Code: 500 Body: some internal error",
		},

		"listing domains from a sigle page": {
			handler: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, url.Values{
					"query":    []string{"*.example.com"},
					"page":     []string{"100"},
					"per_page": []string{"25"},
					"reverse":  []string{"true"},
					"view":     []string{"all"},
				}, r.URL.Query())

				fmt.Fprintf(w, `[{"domain": {"id": 1, "name": "internal.example.com", "ttl": "1080", "authority_type": "M", "addressing_type": "M", "view_id": 10, "notes": "some note"}}]`)
			},
			parameters: globodns.ListDomainRequest{
				Query:          "*.example.com",
				Page:           100,
				DomainsPerPage: 25,
				Reverse:        func(b bool) *bool { return &b }(true),
				View:           "all",
			},
			expected: []globodns.Domain{
				{
					ID:             1,
					Name:           "internal.example.com",
					TTL:            "1080",
					AuthorityType:  "M",
					AddressingType: "M",
					ViewID:         10,
					Notes:          func(s string) *string { return &s }("some note"),
				},
			},
		},

		"listing all domains": {
			handler: func(w http.ResponseWriter, r *http.Request) {
				defer func() { count++ }()

				assert.Equal(t, "1", r.URL.Query().Get("per_page"))

				if count == 0 {
					assert.Equal(t, "1", r.URL.Query().Get("page"))
					fmt.Fprintf(w, `[{"domain": {"id": 1, "name": "example.com", "ttl": "1080"}}]`)
					return
				}

				if count == 1 {
					assert.Equal(t, "2", r.URL.Query().Get("page"))
					fmt.Fprintf(w, `[{"domain": {"id": 2, "name": "example.test", "ttl": "1080"}}]`)
					return
				}

				if count == 2 {
					fmt.Fprintf(w, `[]`)
					return
				}

				require.Fail(t, "should not pass here 4 times")
			},
			parameters: globodns.ListDomainRequest{
				DomainsPerPage: 1,
			},
			expected: []globodns.Domain{
				{ID: 1, Name: "example.com", TTL: "1080"},
				{ID: 2, Name: "example.test", TTL: "1080"},
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			count = 0

			server := httptest.NewServer(tt.handler)
			defer server.Close()

			client, err := globodns.New(nil, server.URL)
			require.NoError(t, err)

			got, err := client.Domain.List(context.TODO(), tt.parameters)
			if tt.expectedError != "" {
				assert.EqualError(t, err, tt.expectedError)
				return
			}

			assert.Equal(t, tt.expected, got)
		})
	}
}
