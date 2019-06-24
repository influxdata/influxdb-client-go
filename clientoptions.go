package influxdb

import (
	"crypto/tls"
	"net/http"
	"net/url"
	"time"
)

// Option is an option for the client config.  If you pass multiple incompatible Options the later one should override.
type Option struct {
	name string
	f    func(*Client) error
}

// HTTPConfig is an https://github.com/influxdata/influxdb1-client compatible client for setting config
// options.  This is here to make transition to the influxdb2 client easy from the old influxdb 1 client library.
// It is recommended that you set the options using the With___ functions instead.
type HTTPConfig struct {
	// Addr should be of the form "http://host:port"
	// or "http://[ipv6-host%zone]:port".
	Addr string

	// Username is the influxdb username, optional.
	Username string

	// Password is the influxdb password, optional.
	Password string

	// UserAgent is the http User Agent, defaults to "InfluxDBClient" plus os and version info.
	UserAgent string

	// Timeout for influxdb writes, if set to zero, it defaults to a 20 second timeout. This is a difference from the influxdb1-client.
	Timeout time.Duration

	// InsecureSkipVerify gets passed to the http client, if true, it will
	// skip https certificate verification. Defaults to false.
	// this currently isn't supported, set on the http client.
	InsecureSkipVerify bool

	// TLSConfig allows the user to set their own TLS config for the HTTP
	// Client. If set, this option overrides InsecureSkipVerify.
	// this currently isn't supported, set on the http client.
	TLSConfig *tls.Config

	// Proxy configures the Proxy function on the HTTP client.
	// this currently isn't supported
	Proxy func(req *http.Request) (*url.URL, error)

	// If HTTPClient is nil, the New Client function will use an http client with sane defaults.
	HTTPClient *http.Client
}

// WithV1Config is an option for setting config in a way that makes it easy to convert from the old influxdb1 client config.
func WithV1Config(conf *HTTPConfig) Option {
	return Option{
		name: "WithV1Config",
		f: func(c *Client) error {
			if conf.Username != "" || conf.Password != "" {
				if err := WithUserAndPass(conf.Username, conf.Password).f(c); err != nil {
					return err
				}
			}
			if conf.UserAgent != "" {
				if err := WithUserAgent(conf.UserAgent).f(c); err != nil {
					return err
				}
			}
			if conf.UserAgent != "" {
				if err := WithUserAgent(conf.UserAgent).f(c); err != nil {
					return err
				}
			}
			if conf.Timeout == 0 {
				c.httpClient.Timeout = conf.Timeout
			}
			if conf.InsecureSkipVerify || conf.TLSConfig != nil || conf.Proxy != nil {
				panic("Unimplemented")
			}
			if conf.HTTPClient != nil {
				if err := WithHTTPClient(conf.HTTPClient).f(c); err != nil {
					return err
				}
			}
			return nil
		},
	}
}

// WithNoCompression returns an option for writing the data to influxdb without compression.
func WithNoCompression() Option {
	return Option{
		name: "WithNoCompression",
		f: func(c *Client) error {
			c.contentEncoding = ""
			return nil
		},
	}
}

// WithGZIP returns an option for setting gzip compression level.
// The default (should this option not be used ) is level 4.
func WithGZIP(n int) Option {
	return Option{
		name: "WithGZIP",
		f: func(c *Client) error {
			c.contentEncoding = "gzip"
			c.compressionLevel = n
			return nil
		},
	}
}

// WithUserAndPass returns an option for setting a username and password, which generates a session for use.
// TODO(docmerlin): session logic.
func WithUserAndPass(username, password string) Option {
	return Option{
		name: "WithUserAndPass",
		f: func(c *Client) error {
			c.username = username
			c.password = password
			return nil
		},
	}
}

// WithUserAgent returns an option for setting a custom useragent string.
func WithUserAgent(ua string) Option {
	return Option{
		name: "WithUserAgent",
		f: func(c *Client) error {
			c.userAgent = ua
			return nil
		},
	}
}

// WithMaxLineBytes returns an option for setting the max length of a line of influx line-protocol in bytes.
func WithMaxLineBytes(n int) Option {
	return Option{
		name: "WithMaxLineBytes",
		f: func(c *Client) error {
			c.maxLineBytes = n
			return nil
		},
	}
}

// WithHTTPClient returns an option for setting a custom HTTP Client
func WithHTTPClient(h *http.Client) Option {
	return Option{
		name: "WithHTTPCLient",
		f: func(c *Client) error {
			c.httpClient = h
			return nil
		},
	}
}
