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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	_, err := New(Params{})
	require.Error(t, err)
	assert.Equal(t, "empty server URL", err.Error())

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
