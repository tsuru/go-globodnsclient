// Copyright 2021 tsuru authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package globodns_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	globodns "github.com/tsuru/go-globodnsclient"
)

func TestClient_BindExport(t *testing.T) {
	tests := map[string]struct {
		handler       http.Handler
		expected      globodns.ScheduleExport
		expectedError string
	}{
		"export successfully scheduled": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				assert.Equal(t, "/bind9/schedule_export.json", r.URL.Path)

				r.Header.Set("Content-Type", "application/json")
				fmt.Fprintf(w, `{"output": "BIND export scheduled for 29 Oct 17:43", "schedule_date": "2021-10-29 17:43:00 -0300"}`)
			}),

			expected: globodns.ScheduleExport{
				Output:       "BIND export scheduled for 29 Oct 17:43",
				ScheduleDate: time.Date(2021, 10, 29, 17, 43, 0, 0, time.Local),
			},
		},

		"server returns an error": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusForbidden)
				fmt.Fprintf(w, `{"error":"NOT AUTHORIZED"}`)
			}),
			expectedError: `globodns: unexpected HTTP status code: Code: 403 Body: {"error":"NOT AUTHORIZED"}`,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(tt.handler)
			defer server.Close()

			client, err := globodns.New(nil, server.URL)
			require.NoError(t, err)

			got, err := client.Bind.Export(context.TODO())
			if tt.expectedError != "" {
				assert.EqualError(t, err, tt.expectedError)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, *got)
		})
	}
}
