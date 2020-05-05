// Copyright 2020 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package influxdb2

import (
	"crypto/tls"
	"time"
)

// Options holds configuration properties for communicating with InfluxDB server
type Options struct {
	// Maximum number of points sent to server in single request. Default 5000
	batchSize uint
	// Interval, in ms, in which is buffer flushed if it has not been already written (by reaching batch size) . Default 1000ms
	flushInterval uint
	// Default retry interval in ms, if not sent by server. Default 1000ms
	retryInterval uint
	// Maximum count of retry attempts of failed writes
	maxRetries uint
	// Maximum number of points to keep for retry. Should be multiple of BatchSize. Default 10,000
	retryBufferLimit uint
	// DebugLevel to filter log messages. Each level mean to log all categories bellow. 0 error, 1 - warning, 2 - info, 3 - debug
	logLevel uint
	// Precision to use in writes for timestamp. In unit of duration: time.Nanosecond, time.Microsecond, time.Millisecond, time.Second
	// Default time.Nanosecond
	precision time.Duration
	// Whether to use GZip compression in requests. Default false
	useGZip bool
	// TLS configuration for secure connection. Default nil
	tlsConfig *tls.Config
	// HTTP request timeout in sec. Default 20
	httpRequestTimeout uint
}

// BatchSize returns size of batch
func (o *Options) BatchSize() uint {
	return o.batchSize
}

// SetBatchSize sets number of points sent in single request
func (o *Options) SetBatchSize(batchSize uint) *Options {
	o.batchSize = batchSize
	return o
}

// FlushInterval returns flush interval in ms
func (o *Options) FlushInterval() uint {
	return o.flushInterval
}

// SetFlushInterval sets flush interval in ms in which is buffer flushed if it has not been already written
func (o *Options) SetFlushInterval(flushIntervalMs uint) *Options {
	o.flushInterval = flushIntervalMs
	return o
}

// RetryInterval returns the retry interval in ms
func (o *Options) RetryInterval() uint {
	return o.retryInterval
}

// SetRetryInterval sets retry interval in ms, which is set if not sent by server
func (o *Options) SetRetryInterval(retryIntervalMs uint) *Options {
	o.retryInterval = retryIntervalMs
	return o
}

// MaxRetries returns maximum count of retry attempts of failed writes
func (o *Options) MaxRetries() uint {
	return o.maxRetries
}

// SetMaxRetries sets maximum count of retry attempts of failed writes
func (o *Options) SetMaxRetries(maxRetries uint) *Options {
	o.maxRetries = maxRetries
	return o
}

// RetryBufferLimit returns retry buffer limit
func (o *Options) RetryBufferLimit() uint {
	return o.retryBufferLimit
}

// SetRetryBufferLimit sets maximum number of points to keep for retry. Should be multiple of BatchSize.
func (o *Options) SetRetryBufferLimit(retryBufferLimit uint) *Options {
	o.retryBufferLimit = retryBufferLimit
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
	return o.precision
}

// SetPrecision sets time precision to use in writes for timestamp. In unit of duration: time.Nanosecond, time.Microsecond, time.Millisecond, time.Second
func (o *Options) SetPrecision(precision time.Duration) *Options {
	o.precision = precision
	return o
}

// UseGZip returns true if write request are gzip`ed
func (o *Options) UseGZip() bool {
	return o.useGZip
}

// SetUseGZip specifies whether to use GZip compression in write requests.
func (o *Options) SetUseGZip(useGZip bool) *Options {
	o.useGZip = useGZip
	return o
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
	return &Options{batchSize: 5000, maxRetries: 3, retryInterval: 1000, flushInterval: 1000, precision: time.Nanosecond, useGZip: false, retryBufferLimit: 10000, httpRequestTimeout: 20}
}
