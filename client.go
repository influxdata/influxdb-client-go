package influxdb

import (
	"context"
	"errors"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/influxdata/influxdb-client-go/internal/gzip"
)

const defaultMaxWait = 10 * time.Second

// Client is a client for writing to influx.
type Client struct {
	httpClient       *http.Client
	contentEncoding  string
	compressionLevel int
	url              *url.URL
	password         string
	username         string
	l                sync.Mutex
	maxRetries       int
	errOnFieldErr    bool
	userAgent        string
	authorization    string // the Authorization header
	maxLineBytes     int
}

// New creates a new Client.  If httpClient is nil, it will use an http client with sane defaults.
// The client is concurrency safe, so feel free to use it and abuse it to your heart's content.
func New(httpClient *http.Client, options ...Option) (*Client, error) {
	c := &Client{
		httpClient:       httpClient,
		contentEncoding:  "gzip",
		compressionLevel: 4,
		errOnFieldErr:    true,
	}
	if c.httpClient == nil {
		c.httpClient = defaultHTTPClient()
	}
	c.url, _ = url.Parse(`http://127.0.0.1:9999/api/v2`)
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
	if c.authorization == "" && !(c.username != "" || c.password != "") {
		return nil, errors.New("a token or a username and password is required, use WithToken(\"the_token\"), or use WithUserAndPass(\"the_username\",\"the_password\")")
	}
	return c, nil
}

// Ping checks the status of cluster.
func (c *Client) Ping(ctx context.Context) error {
	req, err := http.NewRequest(http.MethodGet, c.url.String()+"/ready", nil)
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

// backoff is a helper method for backoff, triesPtr must not be nil.
func (c *Client) backoff(triesPtr *uint64, resp *http.Response, err error) error {
	tries := atomic.LoadUint64(triesPtr)
	if c.maxRetries >= 0 || int(tries) >= c.maxRetries {
		return maxRetriesExceededError{
			err:   err,
			tries: c.maxRetries,
		}
	}
	retry := 0
	if resp != nil {
		retryAfter := resp.Header.Get("Retry-After")
		retry, _ = strconv.Atoi(retryAfter) // we ignore the error here because an error already means retry is 0.
	}
	sleepFor := time.Duration(retry) * time.Second
	if retry == 0 { // if we didn't get a Retry-After or it is zero, instead switch to exponential backoff
		sleepFor = time.Duration(rand.Int63n(((1 << tries) - 1) * 10 * int64(time.Microsecond)))
	}
	if sleepFor > defaultMaxWait {
		sleepFor = defaultMaxWait
	}
	time.Sleep(sleepFor)
	atomic.AddUint64(triesPtr, 1)
	return nil
}

func (c *Client) makeWriteRequest(bucket, org string, body io.Reader) (*http.Request, error) {
	var err error
	if c.contentEncoding == "gzip" {
		body, err = gzip.CompressWithGzip(body, c.compressionLevel)
		if err != nil {
			return nil, err
		}
	}
	u, err := makeWriteURL(c.url, bucket, org)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, u, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "text/plain; charset=utf-8")

	if c.contentEncoding == "gzip" {
		req.Header.Set("Content-Encoding", "gzip")
	}

	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Authorization", c.authorization)

	return req, nil
}

// Close closes any idle connections on the Client.
func (c *Client) Close() error {
	c.httpClient.CloseIdleConnections()
	return nil // we do this, so it qualifies as a closer.
}
