// Copyright 2020-2021 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package influxdb2

import (
	"context"
	"fmt"
	ilog "github.com/influxdata/influxdb-client-go/v2/log"
	"log"
	"net/http"
	"net/http/httptest"
	"runtime"
	"strings"
	"testing"
	"time"

	ihttp "github.com/influxdata/influxdb-client-go/v2/api/http"
	"github.com/influxdata/influxdb-client-go/v2/domain"
	http2 "github.com/influxdata/influxdb-client-go/v2/internal/http"
	iwrite "github.com/influxdata/influxdb-client-go/v2/internal/write"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUrls(t *testing.T) {
	urls := []struct {
		serverURL      string
		serverAPIURL   string
		writeURLPrefix string
	}{
		{"http://host:8086", "http://host:8086/api/v2/", "http://host:8086/api/v2/write"},
		{"http://host:8086/", "http://host:8086/api/v2/", "http://host:8086/api/v2/write"},
		{"http://host:8086/path", "http://host:8086/path/api/v2/", "http://host:8086/path/api/v2/write"},
		{"http://host:8086/path/", "http://host:8086/path/api/v2/", "http://host:8086/path/api/v2/write"},
		{"http://host:8086/path1/path2/path3", "http://host:8086/path1/path2/path3/api/v2/", "http://host:8086/path1/path2/path3/api/v2/write"},
		{"http://host:8086/path1/path2/path3/", "http://host:8086/path1/path2/path3/api/v2/", "http://host:8086/path1/path2/path3/api/v2/write"},
	}
	for _, url := range urls {
		t.Run(url.serverURL, func(t *testing.T) {
			c := NewClient(url.serverURL, "x")
			ci := c.(*clientImpl)
			assert.Equal(t, url.serverURL, ci.serverURL)
			assert.Equal(t, url.serverAPIURL, ci.httpService.ServerAPIURL())
			ws := iwrite.NewService("org", "bucket", ci.httpService, c.Options().WriteOptions())
			wu := ws.WriteURL()
			assert.Equal(t, url.writeURLPrefix+"?bucket=bucket&org=org&precision=ns", wu)
		})
	}
}

func TestWriteAPIManagement(t *testing.T) {
	data := []struct {
		org          string
		bucket       string
		expectedCout int
	}{
		{"o1", "b1", 1},
		{"o1", "b2", 2},
		{"o1", "b1", 2},
		{"o2", "b1", 3},
		{"o2", "b2", 4},
		{"o1", "b2", 4},
		{"o1", "b3", 5},
		{"o2", "b2", 5},
	}
	c := NewClient("http://localhost", "x").(*clientImpl)
	for i, d := range data {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			w := c.WriteAPI(d.org, d.bucket)
			assert.NotNil(t, w)
			assert.Len(t, c.writeAPIs, d.expectedCout)
			wb := c.WriteAPIBlocking(d.org, d.bucket)
			assert.NotNil(t, wb)
			assert.Len(t, c.syncWriteAPIs, d.expectedCout)
		})
	}
	c.Close()
	assert.Len(t, c.writeAPIs, 0)
	assert.Len(t, c.syncWriteAPIs, 0)
}

func TestUserAgentBase(t *testing.T) {
	ua := fmt.Sprintf("influxdb-client-go/%s (%s; %s)", Version, runtime.GOOS, runtime.GOARCH)
	assert.Equal(t, ua, http2.UserAgentBase)

}

type doer struct {
	userAgent string
	doer      ihttp.Doer
}

func (d *doer) Do(req *http.Request) (*http.Response, error) {
	req.Header.Set("User-Agent", d.userAgent)
	return d.doer.Do(req)
}

func TestUserAgent(t *testing.T) {
	ua := http2.UserAgentBase
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-time.After(100 * time.Millisecond)
		if r.Header.Get("User-Agent") == ua {
			w.WriteHeader(http.StatusNoContent)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))

	defer server.Close()
	var sb strings.Builder
	log.SetOutput(&sb)
	log.SetFlags(0)
	c := NewClientWithOptions(server.URL, "x", DefaultOptions().SetLogLevel(ilog.WarningLevel))
	assert.True(t, strings.Contains(sb.String(), "Application name is not set"))
	up, err := c.Ping(context.Background())
	require.NoError(t, err)
	assert.True(t, up)

	err = c.WriteAPIBlocking("o", "b").WriteRecord(context.Background(), "a,a=a a=1i")
	assert.NoError(t, err)

	c.Close()
	sb.Reset()
	// Test setting application  name
	c = NewClientWithOptions(server.URL, "x", DefaultOptions().SetApplicationName("Monitor/1.1"))
	ua = fmt.Sprintf("influxdb-client-go/%s (%s; %s) Monitor/1.1", Version, runtime.GOOS, runtime.GOARCH)
	assert.False(t, strings.Contains(sb.String(), "Application name is not set"))
	up, err = c.Ping(context.Background())
	require.NoError(t, err)
	assert.True(t, up)

	err = c.WriteAPIBlocking("o", "b").WriteRecord(context.Background(), "a,a=a a=1i")
	assert.NoError(t, err)
	c.Close()

	ua = "Monitor/1.1"
	opts := DefaultOptions()
	opts.HTTPOptions().SetHTTPDoer(&doer{
		userAgent: ua,
		doer:      http.DefaultClient,
	})

	//Create client with custom user agent setter
	c = NewClientWithOptions(server.URL, "x", opts)
	up, err = c.Ping(context.Background())
	require.NoError(t, err)
	assert.True(t, up)

	err = c.WriteAPIBlocking("o", "b").WriteRecord(context.Background(), "a,a=a a=1i")
	assert.NoError(t, err)
	c.Close()
}

func TestServerError429(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-time.After(100 * time.Millisecond)
		w.Header().Set("Retry-After", "1")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"code":"too many requests", "message":"exceeded rate limit"}`))
	}))

	defer server.Close()
	c := NewClient(server.URL, "x")
	err := c.WriteAPIBlocking("o", "b").WriteRecord(context.Background(), "a,a=a a=1i")
	require.Error(t, err)
	assert.Equal(t, "too many requests: exceeded rate limit", err.Error())
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
	err := c.WriteAPIBlocking("o", "b").WriteRecord(context.Background(), "a,a=a a=1i")
	require.NoError(t, err)
}

func TestServerErrorNonJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-time.After(100 * time.Millisecond)
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`internal server error`))
	}))

	defer server.Close()
	c := NewClient(server.URL, "x")
	//Test non JSON error in custom code
	err := c.WriteAPIBlocking("o", "b").WriteRecord(context.Background(), "a,a=a a=1i")
	require.Error(t, err)
	assert.Equal(t, "500 Internal Server Error: internal server error", err.Error())

	// Test non JSON error from generated code
	params := &domain.GetBucketsParams{}
	b, err := c.APIClient().GetBuckets(context.Background(), params)
	assert.Nil(t, b)
	require.Error(t, err)
	assert.Equal(t, "500 Internal Server Error: internal server error", err.Error())

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
	err := c.WriteAPIBlocking("o", "b").WriteRecord(context.Background(), "a,a=a a=1i")
	require.Error(t, err)
	assert.Equal(t, "404 Not Found: bruh moment", err.Error())
}

func TestServerErrorEmptyBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
	}))

	defer server.Close()
	c := NewClient(server.URL, "x")
	err := c.WriteAPIBlocking("o", "b").WriteRecord(context.Background(), "a,a=a a=1i")
	require.Error(t, err)
	assert.Equal(t, "Unexpected status code 404", err.Error())
}

func TestReadyFail(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`<html></html>`))
	}))

	defer server.Close()
	c := NewClient(server.URL, "x")
	r, err := c.Ready(context.Background())
	assert.Error(t, err)
	assert.Nil(t, r)
}

func TestHealthFail(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`<html></html>`))
	}))

	defer server.Close()
	c := NewClient(server.URL, "x")
	h, err := c.Health(context.Background())
	assert.Error(t, err)
	assert.Nil(t, h)
}
