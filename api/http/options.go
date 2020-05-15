// Copyright 2020 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package http

import (
	"crypto/tls"
)

// Options holds http configuration properties for communicating with InfluxDB server
type Options struct {
	// TLS configuration for secure connection. Default nil
	tlsConfig *tls.Config
	// HTTP request timeout in sec. Default 20
	httpRequestTimeout uint
}

// TlsConfig returns TlsConfig
func (o *Options) TlsConfig() *tls.Config {
	return o.tlsConfig
}

// SetTlsConfig sets TLS configuration for secure connection
func (o *Options) SetTlsConfig(tlsConfig *tls.Config) *Options {
	o.tlsConfig = tlsConfig
	return o
}

// HttpRequestTimeout returns HTTP request timeout
func (o *Options) HttpRequestTimeout() uint {
	return o.httpRequestTimeout
}

// SetHttpRequestTimeout sets HTTP request timeout in sec
func (o *Options) SetHttpRequestTimeout(httpRequestTimeout uint) *Options {
	o.httpRequestTimeout = httpRequestTimeout
	return o
}

// DefaultOptions returns Options object with default values
func DefaultOptions() *Options {
	return &Options{httpRequestTimeout: 20}
}
