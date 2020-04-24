// Copyright 2020 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package influxdb2

import (
	"context"
	http2 "github.com/influxdata/influxdb-client-go/internal/http"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestUserAgent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		if r.Header.Get("User-Agent") == http2.UserAgent {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))

	defer server.Close()
	c := NewClient(server.URL, "x")
	ready, err := c.Ready(context.Background())
	assert.True(t, ready)
	assert.Nil(t, err)

	err = c.WriteApiBlocking("o", "b").WriteRecord(context.Background(), "a,a=a a=1i")
	assert.Nil(t, err)
}

func TestServerError429(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.Header().Set("Retry-After", "1")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte(`{"code":"too many requests", "message":"exceeded rate limit"}`))
	}))

	defer server.Close()
	c := NewClient(server.URL, "x")
	err := c.WriteApiBlocking("o", "b").WriteRecord(context.Background(), "a,a=a a=1i")
	require.NotNil(t, err)
	perror, ok := err.(*http2.Error)
	require.True(t, ok)
	require.NotNil(t, perror)
	assert.Equal(t, "too many requests", perror.Code)
	assert.Equal(t, "exceeded rate limit", perror.Message)
}

func TestServerErrorNonJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`internal server error`))
	}))

	defer server.Close()
	c := NewClient(server.URL, "x")
	err := c.WriteApiBlocking("o", "b").WriteRecord(context.Background(), "a,a=a a=1i")
	require.NotNil(t, err)
	perror, ok := err.(*http2.Error)
	require.True(t, ok)
	require.NotNil(t, perror)
	assert.Equal(t, "500 Internal Server Error", perror.Code)
	assert.Equal(t, "internal server error", perror.Message)
}
