// Copyright 2020 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package http_test

import (
	"crypto/tls"
	"github.com/influxdata/influxdb-client-go/api/http"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDefaultOptions(t *testing.T) {
	opts := http.DefaultOptions()
	assert.Equal(t, (*tls.Config)(nil), opts.TlsConfig())
	assert.Equal(t, uint(20), opts.HttpRequestTimeout())
}

func TestOptionsSetting(t *testing.T) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}
	opts := http.DefaultOptions().
		SetTlsConfig(tlsConfig).
		SetHttpRequestTimeout(50)
	assert.Equal(t, tlsConfig, opts.TlsConfig())
	assert.Equal(t, uint(50), opts.HttpRequestTimeout())
}
