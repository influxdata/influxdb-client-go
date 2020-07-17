// Copyright 2020 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package influxdb2

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	http2 "github.com/influxdata/influxdb-client-go/internal/http"
	iwrite "github.com/influxdata/influxdb-client-go/internal/write"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUrls(t *testing.T) {
	urls := []struct {
		serverURL      string
		serverAPIURL   string
		writeURLPrefix string
	}{
		{"http://host:9999", "http://host:9999/api/v2/", "http://host:9999/api/v2/write"},
		{"http://host:9999/", "http://host:9999/api/v2/", "http://host:9999/api/v2/write"},
		{"http://host:9999/path", "http://host:9999/path/api/v2/", "http://host:9999/path/api/v2/write"},
		{"http://host:9999/path/", "http://host:9999/path/api/v2/", "http://host:9999/path/api/v2/write"},
		{"http://host:9999/path1/path2/path3", "http://host:9999/path1/path2/path3/api/v2/", "http://host:9999/path1/path2/path3/api/v2/write"},
		{"http://host:9999/path1/path2/path3/", "http://host:9999/path1/path2/path3/api/v2/", "http://host:9999/path1/path2/path3/api/v2/write"},
	}
	for _, url := range urls {
		t.Run(url.serverURL, func(t *testing.T) {
			c := NewClient(url.serverURL, "x")
			ci := c.(*clientImpl)
			assert.Equal(t, url.serverURL, ci.serverURL)
			assert.Equal(t, url.serverAPIURL, ci.httpService.ServerAPIURL())
			ws := iwrite.NewService("org", "bucket", ci.httpService, c.Options().WriteOptions())
			wu, err := ws.WriteURL()
			require.Nil(t, err)
			assert.Equal(t, url.writeURLPrefix+"?bucket=bucket&org=org&precision=ns", wu)
		})
	}
}

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
		_, _ = w.Write([]byte(`{"code":"too many requests", "message":"exceeded rate limit"}`))
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

func TestServerOnPath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/proxy/0:0/influx/api/v2/write" {
			w.WriteHeader(http.StatusNoContent)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(fmt.Sprintf(`{"code":"internal server error", "message":"%s"}`, r.URL.Path)))
		}
	}))

	defer server.Close()
	c := NewClient(server.URL+"/proxy/0:0/influx/", "x")
	err := c.WriteApiBlocking("o", "b").WriteRecord(context.Background(), "a,a=a a=1i")
	require.Nil(t, err)
}

func TestServerErrorNonJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`internal server error`))
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

func TestServerErrorInflux1_8(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Influxdb-Error", "bruh moment")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(404)
		_, _ = w.Write([]byte(`{"error": "bruh moment"}`))
	}))

	defer server.Close()
	c := NewClient(server.URL, "x")
	err := c.WriteApiBlocking("o", "b").WriteRecord(context.Background(), "a,a=a a=1i")
	require.NotNil(t, err)
	perror, ok := err.(*http2.Error)
	require.True(t, ok)
	require.NotNil(t, perror)
	assert.Equal(t, "404 Not Found", perror.Code)
	assert.Equal(t, "bruh moment", perror.Message)
}

func TestServerErrorEmptyBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
	}))

	defer server.Close()
	c := NewClient(server.URL, "x")
	err := c.WriteApiBlocking("o", "b").WriteRecord(context.Background(), "a,a=a a=1i")
	require.NotNil(t, err)
	assert.Equal(t, "Unexpected status code 404", err.Error())
}
