package writer

import (
	"context"
	"time"
)

// Config is a structure used to configure a point writer
type Config struct {
	ctxt          context.Context
	size          int
	flushInterval time.Duration
	retry         bool
	retryOptions  []RetryOption
}

// Option is a functional option for Configuring point writers
type Option func(*Config)

// Options is a slice of Option
type Options []Option

// Config constructs a default configuration and then
// applies the callee options and returns the config
func (o Options) Config() Config {
	config := Config{
		ctxt:          context.Background(),
		size:          defaultBufferSize,
		flushInterval: defaultFlushInterval,
		retry:         true,
		retryOptions:  []RetryOption{},
	}

	o.Apply(&config)

	return config
}

// Apply calls each option in the slice on options on the provided Config
func (o Options) Apply(c *Config) {
	for _, opt := range o {
		opt(c)
	}
}

// WithContext sets the context.Context used for each flush
func WithContext(ctxt context.Context) Option {
	return func(c *Config) {
		c.ctxt = ctxt
	}
}

// WithBufferSize sets the size of the underlying buffer on the point writer
func WithBufferSize(size int) Option {
	return func(c *Config) {
		c.size = size
	}
}

// WithFlushInterval sets the flush interval on the writer
// The point writer will wait at least this long between flushes
// of the undeyling buffered writer
func WithFlushInterval(interval time.Duration) Option {
	return func(c *Config) {
		c.flushInterval = interval
	}
}

// WithRetries configures automatic retry behavior on specific
// transient error conditions when attempting to Write metrics
// to a client
func WithRetries(options ...RetryOption) Option {
	return func(c *Config) {
		c.retry = true
		c.retryOptions = options
	}
}
