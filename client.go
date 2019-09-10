package influxdb

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"
)

// TODO(docmerlin): change the generator so we don't have to hand edit the generated code
//go:generate go run scripts/buildclient.go

// Client is a client for writing to influx.
type Client struct {
	httpClient       *http.Client
	contentEncoding  string
	compressionLevel int
	url              *url.URL
	password         string
	username         string
	l                sync.Mutex
	errOnFieldErr    bool
	userAgent        string
	authorization    string // the Authorization header
	maxLineBytes     int
}

// New creates a new Client.
// The client is concurrency safe, so feel free to use it and abuse it to your heart's content.
func New(connection string, token string, options ...Option) (*Client, error) {
	c := &Client{
		contentEncoding:  "gzip",
		compressionLevel: 4,
		errOnFieldErr:    true,
		authorization:    "Token " + token,
	}

	if connection == "" {
		connection = `http://127.0.0.1:9999`
	}

	u, err := url.Parse(connection)
	if err != nil {
		return nil, fmt.Errorf("Error: could not parse url: %v", err)
	}
	u.Path = `/api/v2`

	c.url = u

	c.userAgent = userAgent()
	for i := range options {
		// check for incompatible options
		if options[i].name == "WithGZIP" {
			for j := range options {
				if options[j].name == "WithNoCompression" {
					return nil, errors.New("the WithGzip is incompatible with the WithNoCompression option")
				}
			}
		}
		if err := options[i].f(c); err != nil {
			return nil, err
		}
	}

	if c.httpClient == nil {
		c.httpClient = defaultHTTPClient()
	}
	if c.authorization == "" && !(c.username != "" || c.password != "") {
		return nil, errors.New("a token or a username and password is required, pass a token to New(), or use WithUserAndPass(\"the_username\",\"the_password\")")
	}
	return c, nil
}

// Ping checks the status of cluster.
func (c *Client) Ping(ctx context.Context) error {
	// deep copy c.url, because we have an entirely different path
	pingURL, _ := url.Parse(c.url.String()) // we don't check the error here, because it just came from an already parsed url.URL object.
	pingURL.Path = "/ready"
	req, err := http.NewRequest(http.MethodGet, pingURL.String(), nil)
	if err != nil {
		return err
	}

	req = req.WithContext(ctx)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// we shouldn't see this
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		var err = errors.New(string(body))
		return err
	}

	return nil
}

// Close closes any idle connections on the Client.
func (c *Client) Close() error {
	c.httpClient.CloseIdleConnections()
	return nil // we do this, so it qualifies as a closer.
}
