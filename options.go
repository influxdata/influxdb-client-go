// Copyright 2020 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package influxdb2

import (
	"crypto/tls"
	"github.com/influxdata/influxdb-client-go/api/http"
	"github.com/influxdata/influxdb-client-go/api/write"
	"time"
)

// Options holds configuration properties for communicating with InfluxDB server
type Options struct {
	// LogLevel to filter log messages. Each level mean to log all categories bellow. 0 error, 1 - warning, 2 - info, 3 - debug
	logLevel uint
	// Writing options
	writeOptions *write.Options
	// Http options
	httpOptions *http.Options
}

// BatchSize returns size of batch
func (o *Options) BatchSize() uint {
	return o.WriteOptions().BatchSize()
}

// SetBatchSize sets number of points sent in single request
func (o *Options) SetBatchSize(batchSize uint) *Options {
	o.WriteOptions().SetBatchSize(batchSize)
	return o
}

// FlushInterval returns flush interval in ms
func (o *Options) FlushInterval() uint {
	return o.WriteOptions().FlushInterval()
}

// SetFlushInterval sets flush interval in ms in which is buffer flushed if it has not been already written
func (o *Options) SetFlushInterval(flushIntervalMs uint) *Options {
	o.WriteOptions().SetFlushInterval(flushIntervalMs)
	return o
}

// RetryInterval returns the retry interval in ms
func (o *Options) RetryInterval() uint {
	return o.WriteOptions().RetryInterval()
}

// SetRetryInterval sets retry interval in ms, which is set if not sent by server
func (o *Options) SetRetryInterval(retryIntervalMs uint) *Options {
	o.WriteOptions().SetRetryInterval(retryIntervalMs)
	return o
}

// MaxRetries returns maximum count of retry attempts of failed writes
func (o *Options) MaxRetries() uint {
	return o.WriteOptions().MaxRetries()
}

// SetMaxRetries sets maximum count of retry attempts of failed writes
func (o *Options) SetMaxRetries(maxRetries uint) *Options {
	o.WriteOptions().SetMaxRetries(maxRetries)
	return o
}

// RetryBufferLimit returns retry buffer limit
func (o *Options) RetryBufferLimit() uint {
	return o.WriteOptions().RetryBufferLimit()
}

// SetRetryBufferLimit sets maximum number of points to keep for retry. Should be multiple of BatchSize.
func (o *Options) SetRetryBufferLimit(retryBufferLimit uint) *Options {
	o.WriteOptions().SetRetryBufferLimit(retryBufferLimit)
	return o
}

// LogLevel returns log level
func (o *Options) LogLevel() uint {
	return o.logLevel
}

// SetLogLevel set level to filter log messages. Each level mean to log all categories bellow. 0 error, 1 - warning, 2 - info, 3 - debug
// Debug level will print also content of writen batches
func (o *Options) SetLogLevel(logLevel uint) *Options {
	o.logLevel = logLevel
	return o
}

// Precision returns time precision for writes
func (o *Options) Precision() time.Duration {
	return o.WriteOptions().Precision()
}

// SetPrecision sets time precision to use in writes for timestamp. In unit of duration: time.Nanosecond, time.Microsecond, time.Millisecond, time.Second
func (o *Options) SetPrecision(precision time.Duration) *Options {
	o.WriteOptions().SetPrecision(precision)
	return o
}

// UseGZip returns true if write request are gzip`ed
func (o *Options) UseGZip() bool {
	return o.WriteOptions().UseGZip()
}

// SetUseGZip specifies whether to use GZip compression in write requests.
func (o *Options) SetUseGZip(useGZip bool) *Options {
	o.WriteOptions().SetUseGZip(useGZip)
	return o
}

// TLSConfig returns TLS config
func (o *Options) TLSConfig() *tls.Config {
	return o.HTTPOptions().TLSConfig()
}

// TlsConfig returns TLS config.
// Deprecated: Use TLSConfig instead.
func (o *Options) TlsConfig() *tls.Config {
	return o.TLSConfig()
}

// SetTLSConfig sets TLS configuration for secure connection
func (o *Options) SetTLSConfig(tlsConfig *tls.Config) *Options {
	o.HTTPOptions().SetTLSConfig(tlsConfig)
	return o
}

// SetTlsConfig sets TLS configuration for secure connection.
// Deprecated: Use SetTLSConfig instead.
func (o *Options) SetTlsConfig(tlsConfig *tls.Config) *Options {
	return o.SetTLSConfig(tlsConfig)
}

// HTTPRequestTimeout returns HTTP request timeout
func (o *Options) HTTPRequestTimeout() uint {
	return o.HTTPOptions().HTTPRequestTimeout()
}

// HttpRequestTimeout returns HTTP request timeout.
// Deprecated: Use HTTPRequestTimeout instead.
func (o *Options) HttpRequestTimeout() uint {
	return o.HTTPRequestTimeout()
}

// SetHTTPRequestTimeout sets HTTP request timeout in sec
func (o *Options) SetHTTPRequestTimeout(httpRequestTimeout uint) *Options {
	o.HTTPOptions().SetHTTPRequestTimeout(httpRequestTimeout)
	return o
}

// SetHttpRequestTimeout sets HTTP request timeout in sec
// Deprecated: Use SetHTTPRequestTimeout instead
func (o *Options) SetHttpRequestTimeout(httpRequestTimeout uint) *Options {
	return o.SetHTTPRequestTimeout(httpRequestTimeout)
}

// WriteOptions returns write related options
func (o *Options) WriteOptions() *write.Options {
	if o.writeOptions == nil {
		o.writeOptions = write.DefaultOptions()
	}
	return o.writeOptions
}

// HTTPOptions returns HTTP related options
func (o *Options) HTTPOptions() *http.Options {
	if o.httpOptions == nil {
		o.httpOptions = http.DefaultOptions()
	}
	return o.httpOptions
}

// HttpOptions returns http related options
// Deprecated: Use HTTPOptions instead
func (o *Options) HttpOptions() *http.Options {
	return o.HTTPOptions()
}

// AddDefaultTag adds a default tag. DefaultTags are added to each written point.
// If a tag with the same key already exist it is overwritten.
// If a point already defines such a tag, it is left unchanged
func (o *Options) AddDefaultTag(key, value string) *Options {
	o.WriteOptions().AddDefaultTag(key, value)
	return o
}

// DefaultOptions returns Options object with default values
func DefaultOptions() *Options {
	return &Options{logLevel: 0, writeOptions: write.DefaultOptions(), httpOptions: http.DefaultOptions()}
}
