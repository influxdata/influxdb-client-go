package writer

import "time"

// Config is a structure used to configure a point writer
type Config struct {
	size          int
	flushInterval time.Duration
}

// Option is a functional option for Configuring point writers
type Option func(*Config)

// Options is a slice of Option
type Options []Option

// Config constructs a default configuration and then
// applies the callee options and returns the config
func (o Options) Config() Config {
	config := Config{
		size:          defaultBufferSize,
		flushInterval: defaultFlushInterval,
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
