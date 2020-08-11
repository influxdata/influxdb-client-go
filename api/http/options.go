// Copyright 2020 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

// Package http holds HTTP options
package http

import (
	"crypto/tls"
	"net"
	"net/http"
	"time"
)

// Options holds http configuration properties for communicating with InfluxDB server
type Options struct {
	// HTTP client. Default is http.DefaultClient.
	httpClient *http.Client
	// TLS configuration for secure connection. Default nil
	tlsConfig *tls.Config
	// HTTP request timeout in sec. Default 20
	httpRequestTimeout uint
}

// HTTPClient returns the http.Client that is configured to be used
// for HTTP requests. It will return the one that has been set using
// SetHTTPClient or it will construct a default client using the
// other configured options.
func (o *Options) HTTPClient() *http.Client {
	if o.httpClient == nil {
		o.httpClient = &http.Client{
			Timeout: time.Second * time.Duration(o.HTTPRequestTimeout()),
			Transport: &http.Transport{
				DialContext: (&net.Dialer{
					Timeout: 5 * time.Second,
				}).DialContext,
				TLSHandshakeTimeout: 5 * time.Second,
				TLSClientConfig:     o.TLSConfig(),
			},
		}
	}
	return o.httpClient
}

// SetHTTPClient will configure the http.Client that is used
// for HTTP requests. If set to nil, an HTTPClient will be
// generated.
//
// Setting the HTTPClient will cause the other HTTP options
// to be ignored.
func (o *Options) SetHTTPClient(c *http.Client) {
	o.httpClient = c
}

// TLSConfig returns tls.Config
func (o *Options) TLSConfig() *tls.Config {
	return o.tlsConfig
}

// SetTLSConfig sets TLS configuration for secure connection
func (o *Options) SetTLSConfig(tlsConfig *tls.Config) *Options {
	o.tlsConfig = tlsConfig
	return o
}

// HTTPRequestTimeout returns HTTP request timeout
func (o *Options) HTTPRequestTimeout() uint {
	return o.httpRequestTimeout
}

// SetHTTPRequestTimeout sets HTTP request timeout in sec
func (o *Options) SetHTTPRequestTimeout(httpRequestTimeout uint) *Options {
	o.httpRequestTimeout = httpRequestTimeout
	return o
}

// DefaultOptions returns Options object with default values
func DefaultOptions() *Options {
	return &Options{httpRequestTimeout: 20}
}
