// Copyright 2020 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

// Package http holds HTTP options
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

// TLSConfig returns tls.Config
func (o *Options) TLSConfig() *tls.Config {
	return o.tlsConfig
}

// TlsConfig returns TlsConfig
// Deprecated: Use TLSConfig instead.
//lint:ignore ST1003 Deprecated method to be removed in the next release
func (o *Options) TlsConfig() *tls.Config {
	return o.TLSConfig()
}

// SetTLSConfig sets TLS configuration for secure connection
func (o *Options) SetTLSConfig(tlsConfig *tls.Config) *Options {
	o.tlsConfig = tlsConfig
	return o
}

// SetTlsConfig sets TLS configuration for secure connection
// Deprecated: Use SetTLSConfig  instead.
//lint:ignore ST1003 Deprecated method to be removed in the next release
func (o *Options) SetTlsConfig(tlsConfig *tls.Config) *Options {
	return o.SetTLSConfig(tlsConfig)
}

// HTTPRequestTimeout returns HTTP request timeout
func (o *Options) HTTPRequestTimeout() uint {
	return o.httpRequestTimeout
}

// HttpRequestTimeout returns HTTP request timeout.
// Deprecated: Use HTTPRequestTimeout instead.
//lint:ignore ST1003 Deprecated method to be removed in the next release
func (o *Options) HttpRequestTimeout() uint {
	return o.HTTPRequestTimeout()
}

// SetHTTPRequestTimeout sets HTTP request timeout in sec
func (o *Options) SetHTTPRequestTimeout(httpRequestTimeout uint) *Options {
	o.httpRequestTimeout = httpRequestTimeout
	return o
}

// SetHttpRequestTimeout sets HTTP request timeout in sec.
// Deprecated: Use SetHTTPRequestTimeout instead.
//lint:ignore ST1003 Deprecated method to be removed in the next release
func (o *Options) SetHttpRequestTimeout(httpRequestTimeout uint) *Options {
	return o.SetHTTPRequestTimeout(httpRequestTimeout)
}

// DefaultOptions returns Options object with default values
func DefaultOptions() *Options {
	return &Options{httpRequestTimeout: 20}
}
