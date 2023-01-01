// Copyright 2020-2021 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package http_test

import (
	"crypto/tls"
	"github.com/stretchr/testify/require"
	nethttp "net/http"
	"testing"
	"time"

	"github.com/influxdata/influxdb-client-go/v2/api/http"
	"github.com/stretchr/testify/assert"
)

func TestDefaultOptions(t *testing.T) {
	opts := http.DefaultOptions()
	assert.Equal(t, (*tls.Config)(nil), opts.TLSConfig())
	assert.Equal(t, uint(20), opts.HTTPRequestTimeout())
	assert.NotNil(t, opts.HTTPClient())
	assert.True(t, opts.OwnHTTPClient())
	transport, ok := opts.HTTPClient().Transport.(*nethttp.Transport)
	require.True(t, ok)
	assert.NotNil(t, transport.Proxy)
	assert.EqualValues(t, "", opts.ApplicationName())
}

func TestOptionsSetting(t *testing.T) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}
	opts := http.DefaultOptions().
		SetTLSConfig(tlsConfig).
		SetHTTPRequestTimeout(50).
		SetApplicationName("Monitor/1.1")
	assert.Equal(t, tlsConfig, opts.TLSConfig())
	assert.Equal(t, uint(50), opts.HTTPRequestTimeout())
	assert.EqualValues(t, "Monitor/1.1", opts.ApplicationName())
	if client := opts.HTTPClient(); assert.NotNil(t, client) {
		assert.Equal(t, 50*time.Second, client.Timeout)
		assert.Equal(t, tlsConfig, client.Transport.(*nethttp.Transport).TLSClientConfig)
		assert.True(t, opts.OwnHTTPClient())
	}

	client := &nethttp.Client{
		Transport: &nethttp.Transport{},
	}
	opts = http.DefaultOptions()
	opts.SetHTTPClient(client)
	assert.Equal(t, client, opts.HTTPClient())
	assert.False(t, opts.OwnHTTPClient())
}
