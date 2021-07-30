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

func TestClient_RecordList(t *testing.T) {
	var count int

	tests := map[string]struct {
		handler       http.HandlerFunc
		domainID      int
		params        *globodns.ListRecordsParameters
		expected      []globodns.Record
		expectedError string
	}{
		"domain id < 0": {
			domainID:      -100,
			expectedError: "globodns: domain ID cannot be negative",
		},

		"records per page < 0": {
			domainID:      10,
			params:        &globodns.ListRecordsParameters{PerPage: -10},
			expectedError: "globodns: records per page cannot be negative",
		},

		"page < 0": {
			domainID:      10,
			params:        &globodns.ListRecordsParameters{Page: -10},
			expectedError: "globodns: page cannot be negative",
		},

		"servers returns an error": {
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "some internal error")
			},
			expectedError: "globodns: unexpected HTTP status code: Code: 500 Body: some internal error",
		},

		"list records from a single page": {
			handler: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, url.Values{
					"query":    []string{"*"},
					"page":     []string{"100"},
					"per_page": []string{"25"},
					"reverse":  []string{"true"},
				}, r.URL.Query())

				fmt.Fprintf(w, `[
	{"a":   {"name": "www", "content": "169.196.100.100"}},
	{"mx":  {"name": "@", "content": "mail", "ttl": 86400}},
	{"a":   {"name": "mail", "content": "169.196.100.101"}},
	{"txt": {"name": "@", "content": "my TXT record!!!", "ttl": 60}} ]`)
			},
			domainID: 10,
			params: &globodns.ListRecordsParameters{
				Page:    100,
				PerPage: 25,
				Query:   "*",
				Reverse: func(b bool) *bool { return &b }(true),
			},
			expected: []globodns.Record{
				{Name: "www", Content: "169.196.100.100", Type: "A"},
				{Name: "@", Content: "mail", TTL: func(n int) *int { return &n }(86400), Type: "MX"},
				{Name: "mail", Content: "169.196.100.101", Type: "A"},
				{Name: "@", Content: "my TXT record!!!", TTL: func(n int) *int { return &n }(60), Type: "TXT"},
			},
		},

		"list all resources": {
			handler: func(w http.ResponseWriter, r *http.Request) {
				defer func() { count++ }()

				assert.Equal(t, "1", r.URL.Query().Get("per_page"))

				if count == 0 {
					assert.Equal(t, "1", r.URL.Query().Get("page"))
					fmt.Fprintf(w, `[{"a":   {"name": "www", "content": "169.196.100.100"}}]`)
					return
				}

				if count == 1 {
					assert.Equal(t, "2", r.URL.Query().Get("page"))
					fmt.Fprintf(w, `[{"mx":  {"name": "@", "content": "mail", "ttl": 86400}}]`)
					return
				}

				if count == 2 {
					assert.Equal(t, "3", r.URL.Query().Get("page"))
					fmt.Fprintf(w, `[{"a":   {"name": "mail", "content": "169.196.100.101"}}]`)
					return
				}

				if count == 3 {
					assert.Equal(t, "4", r.URL.Query().Get("page"))
					fmt.Fprintf(w, `[{"txt": {"name": "@", "content": "my TXT record!!!", "ttl": 60}}]`)
					return
				}

				if count == 4 {
					assert.Equal(t, "5", r.URL.Query().Get("page"))
					fmt.Fprintf(w, `[]`)
					return
				}

				require.Fail(t, "should not pass here")
			},
			params: &globodns.ListRecordsParameters{PerPage: 1},
			expected: []globodns.Record{
				{Name: "www", Content: "169.196.100.100", Type: "A"},
				{Name: "@", Content: "mail", TTL: func(n int) *int { return &n }(86400), Type: "MX"},
				{Name: "mail", Content: "169.196.100.101", Type: "A"},
				{Name: "@", Content: "my TXT record!!!", TTL: func(n int) *int { return &n }(60), Type: "TXT"},
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

			got, err := client.Record.List(context.TODO(), tt.domainID, tt.params)
			if tt.expectedError != "" {
				assert.EqualError(t, err, tt.expectedError)
				return
			}

			assert.Equal(t, tt.expected, got)
		})
	}
}
