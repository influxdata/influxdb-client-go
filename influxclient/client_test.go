// Copyright 2021 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package influxclient

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	_, err := New(Params{})
	require.Error(t, err)
	assert.Equal(t, "empty server URL", err.Error())

	_, err = New(Params{ServerURL: "http@localhost:8086"})
	require.Error(t, err)
	assert.Equal(t, "error parsing server URL: parse \"http@localhost:8086/\": first path segment in URL cannot contain colon", err.Error())

	c, err := New(Params{ServerURL: "http://localhost:8086"})
	require.NoError(t, err)
	assert.Equal(t, "http://localhost:8086", c.params.ServerURL)
	assert.Equal(t, "http://localhost:8086/api/v2/", c.apiURL.String())
	assert.Equal(t, "", c.authorization)

	_, err = New(Params{ServerURL: "localhost\n"})
	if assert.Error(t, err) {
		assert.True(t, strings.HasPrefix(err.Error(), "error parsing server URL:"))
	}

	c, err = New(Params{ServerURL: "http://localhost:8086", AuthToken: "my-token"})
	require.NoError(t, err)
	assert.Equal(t, "Token my-token", c.authorization)
	assert.EqualValues(t, DefaultWriteParams, c.params.WriteParams)
}

func TestURLs(t *testing.T) {
	urls := []struct {
		serverURL    string
		serverAPIURL string
	}{
		{"http://host:8086", "http://host:8086/api/v2/"},
		{"http://host:8086/", "http://host:8086/api/v2/"},
		{"http://host:8086/path", "http://host:8086/path/api/v2/"},
		{"http://host:8086/path/", "http://host:8086/path/api/v2/"},
		{"http://host:8086/path1/path2/path3", "http://host:8086/path1/path2/path3/api/v2/"},
		{"http://host:8086/path1/path2/path3/", "http://host:8086/path1/path2/path3/api/v2/"},
	}
	for _, turl := range urls {
		t.Run(turl.serverURL, func(t *testing.T) {
			c, err := New(Params{ServerURL: turl.serverURL})
			require.NoError(t, err)
			assert.Equal(t, turl.serverURL, c.params.ServerURL)
			assert.Equal(t, turl.serverAPIURL, c.apiURL.String())
		})
	}
}

func TestResolveErrorMessage(t *testing.T) {
	errMsg := "compilation failed: error at @1:170-1:171: invalid expression @1:167-1:168: |"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(400)
		_, _ = w.Write([]byte(`{"code":"invalid","message":"` + errMsg + `"}`))
	}))
	defer ts.Close()
	client, err := New(Params{ServerURL: ts.URL})
	require.NoError(t, err)
	turl, err := url.Parse(ts.URL)
	require.NoError(t, err)
	res, err := client.makeAPICall(context.Background(), httpParams{
		endpointURL: turl,
		queryParams: nil,
		httpMethod:  "GET",
		headers:     nil,
		body:        nil,
	})
	assert.Nil(t, res)
	require.Error(t, err)
	assert.Equal(t, "invalid: "+errMsg, err.Error())
}

func TestResolveErrorHTML(t *testing.T) {
	html := `<html><body><h1>Not found</h1></body></html>`
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/html")
		w.WriteHeader(404)
		_, _ = w.Write([]byte(html))
	}))
	defer ts.Close()
	client, err := New(Params{ServerURL: ts.URL})
	require.NoError(t, err)
	turl, err := url.Parse(ts.URL)
	require.NoError(t, err)
	res, err := client.makeAPICall(context.Background(), httpParams{
		endpointURL: turl,
		queryParams: nil,
		httpMethod:  "GET",
		headers:     nil,
		body:        nil,
	})
	assert.Nil(t, res)
	require.Error(t, err)
	assert.Equal(t, html, err.Error())
}

func TestResolveErrorV1(t *testing.T) {
	errMsg := "compilation failed: error at @1:170-1:171: invalid expression @1:167-1:168: |"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(400)
		_, _ = w.Write([]byte(`{"error": "` + errMsg + `"}`))
	}))
	defer ts.Close()
	client, err := New(Params{ServerURL: ts.URL})
	require.NoError(t, err)
	turl, err := url.Parse(ts.URL)
	require.NoError(t, err)
	res, err := client.makeAPICall(context.Background(), httpParams{
		endpointURL: turl,
		queryParams: nil,
		httpMethod:  "GET",
		headers:     nil,
		body:        nil,
	})
	assert.Nil(t, res)
	require.Error(t, err)
	assert.Equal(t, errMsg, err.Error())
}

func TestResolveErrorNoError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	}))
	defer ts.Close()
	client, err := New(Params{ServerURL: ts.URL})
	require.NoError(t, err)
	turl, err := url.Parse(ts.URL)
	require.NoError(t, err)
	res, err := client.makeAPICall(context.Background(), httpParams{
		endpointURL: turl,
		queryParams: nil,
		httpMethod:  "GET",
		headers:     nil,
		body:        nil,
	})
	assert.Nil(t, res)
	require.Error(t, err)
	assert.Equal(t, `500 Internal Server Error`, err.Error())
}

func TestReadyOk(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{
    "status": "ready",
    "started": "2021-02-24T12:13:37.681813026Z",
    "up": "5713h41m50.256128486s"
}`))
	}))
	defer ts.Close()
	client, err := New(Params{ServerURL: ts.URL, AuthToken: ""})
	require.NoError(t, err)
	dur, err := client.Ready(context.Background())
	require.NoError(t, err)
	exp := 5713*time.Hour + 41*time.Minute + 50*time.Second + 256128486*time.Nanosecond
	assert.Equal(t, exp, dur)
}

func TestReadyInvalidJson(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{
    "status": "ready",
}`))
	}))
	defer ts.Close()
	client, err := New(Params{ServerURL: ts.URL, AuthToken: ""})
	require.NoError(t, err)
	dur, err := client.Ready(context.Background())
	assert.Error(t, err)
	assert.Zero(t, dur)
}

func TestReadyInvalidDuration(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{
    "status": "ready",
    "started": "2021-02-24T12:13:37.681813026Z",
    "up": "1t"
}`))
	}))
	defer ts.Close()
	client, err := New(Params{ServerURL: ts.URL, AuthToken: ""})
	require.NoError(t, err)
	dur, err := client.Ready(context.Background())
	assert.Error(t, err)
	assert.Zero(t, dur)
}

func TestReadyHtml(t *testing.T) {
	html := `<!doctype html><html lang="en"><body><div id="react-root" data-basepath=""></div><script src="/static/39f7ddd770.js"></script></body></html>`
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/html")
		w.WriteHeader(200)
		w.Write([]byte(html))
	}))
	defer ts.Close()
	client, err := New(Params{ServerURL: ts.URL, AuthToken: ""})
	require.NoError(t, err)
	check, err := client.Ready(context.Background())
	require.Error(t, err)
	assert.Zero(t, check)
	assert.Equal(t, "error calling Ready: invalid character '<' looking for beginning of value", err.Error())
}

func TestReadyFail(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(503)
		w.Write([]byte{})
	}))
	defer ts.Close()
	client, err := New(Params{ServerURL: ts.URL, AuthToken: ""})
	require.NoError(t, err)
	dur, err := client.Ready(context.Background())
	require.Error(t, err)
	assert.Zero(t, dur)
}

func TestHealthOk(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{"name":"influxdb", "message":"ready for queries and writes", "status":"pass", "checks":[], "version": "2.0.4", "commit": "4e7a59bb9a"}`))
	}))
	defer ts.Close()
	client, err := New(Params{ServerURL: ts.URL, AuthToken: ""})
	require.NoError(t, err)
	check, err := client.Health(context.Background())
	require.NoError(t, err)
	require.NotNil(t, check)
	assert.Equal(t, "influxdb", check.Name)
	assert.Equal(t, "pass", string(check.Status))
	if assert.NotNil(t, check.Message) {
		assert.Equal(t, "ready for queries and writes", *check.Message)
	}
	if assert.NotNil(t, check.Commit) {
		assert.Equal(t, "4e7a59bb9a", *check.Commit)
	}
	if assert.NotNil(t, check.Version) {
		assert.Equal(t, "2.0.4", *check.Version)
	}
	if assert.NotNil(t, check.Checks) {
		assert.Len(t, *check.Checks, 0)
	}
}

func TestHealthInvalidJson(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{"name":"influxdb",}`))
	}))
	defer ts.Close()
	client, err := New(Params{ServerURL: ts.URL, AuthToken: ""})
	require.NoError(t, err)
	check, err := client.Health(context.Background())
	assert.Error(t, err)
	assert.Nil(t, check)
}

func TestHealthHtml(t *testing.T) {
	html := `<!doctype html><html lang="en"><body><div id="react-root" data-basepath=""></div><script src="/static/39f7ddd770.js"></script></body></html>`
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/html")
		w.WriteHeader(200)
		w.Write([]byte(html))
	}))
	defer ts.Close()
	client, err := New(Params{ServerURL: ts.URL, AuthToken: ""})
	require.NoError(t, err)
	check, err := client.Health(context.Background())
	require.Error(t, err)
	assert.Nil(t, check)
	assert.Equal(t, "error calling Health: invalid character '<' looking for beginning of value", err.Error())
}

func TestHealthFail(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(503)
		w.Write([]byte{})
	}))
	defer ts.Close()
	client, err := New(Params{ServerURL: ts.URL, AuthToken: ""})
	require.NoError(t, err)
	check, err := client.Health(context.Background())
	require.Error(t, err)
	require.Nil(t, check)
}

func TestPingOk(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(204)
		w.Write([]byte{})
	}))
	defer ts.Close()
	client, err := New(Params{ServerURL: ts.URL, AuthToken: ""})
	require.NoError(t, err)
	err = client.Ping(context.Background())
	require.NoError(t, err)
}

func TestPingFail(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(503)
		w.Write([]byte{})
	}))
	defer ts.Close()
	client, err := New(Params{ServerURL: ts.URL, AuthToken: ""})
	require.NoError(t, err)
	err = client.Ping(context.Background())
	require.Error(t, err)
}

func TestDeletePointsOk(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(204)
	}))
	defer ts.Close()
	client, err := New(Params{ServerURL: ts.URL, AuthToken: ""})
	require.NoError(t, err)
	start, err := time.Parse(time.RFC3339, "2019-08-24T14:15:00Z")
	require.NoError(t, err)
	stop, err := time.Parse(time.RFC3339, "2019-08-25T14:15:00Z")
	require.NoError(t, err)
	err = client.DeletePoints(context.Background(), &DeleteParams{
		Org: "my-org",
		Bucket: "my-bucket",
		Predicate: `_measurement="sensorData"`,
		Start: start,
		Stop: stop,
	})
	require.NoError(t, err)
}

func TestDeletePointsFail(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(204)
	}))
	defer ts.Close()
	client, err := New(Params{ServerURL: ts.URL, AuthToken: ""})
	require.NoError(t, err)
	err = client.DeletePoints(context.Background(), nil)
	assert.Error(t, err)
	err = client.DeletePoints(context.Background(), &DeleteParams{})
	assert.Error(t, err)
	err = client.DeletePoints(context.Background(), &DeleteParams{
		Org: "my-org",
	})
	assert.Error(t, err)
	err = client.DeletePoints(context.Background(), &DeleteParams{
		Bucket: "my-bucket",
	})
	assert.Error(t, err)
	err = client.DeletePoints(context.Background(), &DeleteParams{
		Org: "my-org",
		Bucket: "my-bucket",
	})
	assert.Error(t, err)
	err = client.DeletePoints(context.Background(), &DeleteParams{
		Org: "my-org",
		Bucket: "my-bucket",
		Start: time.Now(),
	})
	assert.Error(t, err)
	err = client.DeletePoints(context.Background(), &DeleteParams{
		Org: "my-org",
		Bucket: "my-bucket",
		Stop: time.Now(),
	})
	assert.Error(t, err)
}
