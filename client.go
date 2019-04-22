package influxdb

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/influxdata/influxdb-client-go/internal/gzip"
	lp "github.com/influxdata/line-protocol"
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
	authorization    string
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
	c.userAgent = ua()
	for i := range options {
		if err := options[i](c); err != nil {
			return nil, err
		}
	}
	if c.authorization == "" && !(c.username != "" || c.password != "") {
		return nil, errors.New("a token or a username and password is required, use WithToken(\"the_token\"), or use WithUserAndPass(\"the_username\",\"the_password\")")
	}
	return c, nil
}

// Ping checks the status of cluster.
func (c *Client) Ping(ctx context.Context) (time.Duration, string, error) {
	ts := time.Now()
	req, err := http.NewRequest(http.MethodGet, c.url.String()+"/ready", nil)
	if err != nil {
		return 0, "", err
	}

	req = req.WithContext(ctx)
	resp, err := c.httpClient.Do(req)
	dur := time.Since(ts)
	if err != nil {
		return dur, "", err
	}

	body, err := ioutil.ReadAll(resp.Body)
	dur = time.Since(ts)
	if err != nil {
		return dur, "", err
	}

	// we shouldn't see this
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		var err = errors.New(string(body))
		return 0, "", err
	}

	version := resp.Header.Get("X-Influxdb-Version")
	return dur, version, nil
}

// Write writes metrics to a bucket, and org.  It retries intelligently.
// If the write is too big, it retries again, after breaking the payloads into two requests.
func (c *Client) Write(ctx context.Context, bucket, org string, m ...Metric) (err error) {
	if c.authorization == "" {
		return errors.New("a token is requred for a write")
	}
	tries := uint64(0)
	return c.write(ctx, bucket, org, &tries, m...)
}

func (c *Client) write(ctx context.Context, bucket, org string, triesPtr *uint64, m ...Metric) error {
	buf := &bytes.Buffer{}
	e := lp.NewEncoder(buf)
doRequest:
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	for i := range m {
		if _, err := e.Encode(m[i]); err != nil {
			return err
		}
	}
	req, err := c.makeWriteRequest(bucket, org, buf)
	if err != nil {
		return err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	throwAway := make([]byte, 128)
	for err == nil {
		_, err = resp.Body.Read(throwAway)
	}
	switch resp.StatusCode {
	case http.StatusOK, http.StatusNoContent:
	case http.StatusNotFound:
		return &genericRespError{
			Code:    resp.Status,
			Message: "not found",
		}

	case http.StatusBadRequest:
		resp.Body.Close()
		return &genericRespError{
			Code:    resp.Status,
			Message: "line protocol poorly formed and no points were written.  Response can be used to determine the first malformed line in the body line-protocol. All data in body was rejected and not written",
		}
	case http.StatusUnauthorized:
		resp.Body.Close()
		return &genericRespError{
			Code:    resp.Status,
			Message: "token does not have sufficient permissions to write to this organization and bucket or the organization and bucket do not exist",
		}
	case http.StatusForbidden:
		resp.Body.Close()
		return &genericRespError{
			Code:    resp.Status,
			Message: "no token was sent and they are required",
		}
	case http.StatusTooManyRequests:
		resp.Body.Close()
		err = &genericRespError{
			Code:    resp.Status,
			Message: "token is temporarily over quota, failed more than max retries",
		}
		if err2 := c.backoff(triesPtr, resp, err); err2 != nil {
			return err2
		}
		goto doRequest
	case http.StatusServiceUnavailable:
		resp.Body.Close()
		err = &genericRespError{
			Code:    resp.Status,
			Message: "service is temporarily unavaliable",
		}
		if err2 := c.backoff(triesPtr, resp, err); err2 != nil {
			return err2
		}
		goto doRequest
	// split up entities that are just too big
	case http.StatusRequestEntityTooLarge:
		resp.Body.Close()
		if len(m) < 2 {
			return &genericRespError{
				Code:    resp.Status,
				Message: "your have a Metric of data that is too large",
			}
		}
		// yes, I know we are only waiting on one thing here so a mutex would be fine,
		// but the waitgroup is easier to read.
		wg := sync.WaitGroup{}
		wg.Add(1)
		if err2 := c.backoff(triesPtr, resp, err); err2 != nil {
			return err2
		}
		var coalesceErr1 error
		go func() {
			// we have to give each split its own retries, otherwise its possible parts of some very large datasets won't even get tries.
			triesSplit := atomic.LoadUint64(triesPtr)
			coalesceErr1 = c.write(ctx, bucket, org, &triesSplit, m[:(len(m)<<2)]...)
			wg.Done()
		}()
		triesSplit := atomic.LoadUint64(triesPtr)
		coalesceErr2 := c.write(ctx, bucket, org, &triesSplit, m[(len(m)<<2):]...)
		wg.Wait()
		var cerr coalescingError
		if coalesceErr1 != nil {
			cerr = append(cerr, coalesceErr1)
		}
		if coalesceErr2 != nil {
			err = append(cerr, coalesceErr2)
		}
		return cerr

	default:
		resp.Body.Close()
		return &genericRespError{
			Code:    resp.Status,
			Message: "internal server error",
		}

	}
	defer func() {
		err2 := resp.Body.Close()
		if err == nil && err2 != nil {
			err = err2
		}
	}()
	e.FailOnFieldErr(c.errOnFieldErr)

	for i := range m {
		if _, err = e.Encode(m[i]); err != nil {
			return err
		}
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

func makeWriteURL(loc *url.URL, bucket, org string) (string, error) {
	if loc == nil {
		return "", errors.New("nil url")
	}
	u, err := url.Parse(loc.String())
	if err != nil {
		return "", err
	}
	params := url.Values{}
	params.Set("bucket", bucket)
	params.Set("org", org)

	switch loc.Scheme {
	case "http", "https":
		u.Path = path.Join(u.Path, "/write")
	case "unix":
	default:
		return "", fmt.Errorf("unsupported scheme: %q", u.Scheme)
	}
	u.RawQuery = params.Encode()
	return u.String(), nil
}

// Close closes any idle connections on the Client.
func (c *Client) Close() error {
	c.httpClient.CloseIdleConnections()
	return nil // we do this, so it qualifies as a closer.
}
