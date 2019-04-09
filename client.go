package client

import (
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
	"time"

	"github.com/influxdata/influxdb-client-go/internal/gzip"
	lp "github.com/influxdata/line-protocol"
)

const defaultMaxWait = 10 * time.Second

// Client is a client for writing to influx.
type Client struct {
	httpClient           *http.Client
	contentEncoding      string
	gzipCompressionLevel int
	url                  *url.URL
	password             string
	username             string
	token                string
	org                  string
	maxRetries           int
	errOnFieldErr        bool
	userAgent            string
	authorization        string
	maxLineBytes         int
}

// NewClient creates a new Client.  If httpClient is nil, it will use an http client with sane defaults.
func NewClient(httpClient *http.Client, options ...Option) (*Client, error) {
	c := &Client{
		httpClient:      httpClient,
		contentEncoding: "gzip",
	}
	if c.httpClient == nil {
		c.httpClient = defaultHTTPClient()
	}
	c.url, _ = url.Parse(`http://127.0.0.1:9999/api/v2`)
	c.userAgent = ua()
	if c.token != "" {
		c.authorization = "Token " + c.token
	}
	for i := range options {
		if err := options[i](c); err != nil {
			return nil, err
		}
	}
	if c.token == "" {
		return nil, errors.New("a token is required, use WithToken(\"the_token\")")
	}
	return c, nil
}

// Ping checks the status of cluster.
func (c *Client) Ping(ctx context.Context) (time.Duration, string, error) {
	ts := time.Now()
	req, err := http.NewRequest("GET", c.url.String()+"/ready", nil)
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

func (c *Client) Write(ctx context.Context, bucket, org string, m ...lp.Metric) (err error) {
	r, w := io.Pipe()
	e := lp.NewEncoder(w)
	req, err := c.makeWriteRequest(bucket, org, r)
	if err != nil {
		return err
	}
	tries := uint(0)
doRequest:
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	switch resp.StatusCode {
	case http.StatusOK, http.StatusNoContent:
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
	case http.StatusRequestEntityTooLarge:
		resp.Body.Close()
		return &genericRespError{
			Code:    resp.Status,
			Message: "write has been rejected because the payload is too large. Error message returns max size supported. All data in body was rejected and not written",
		}
	case http.StatusTooManyRequests:
		err = &genericRespError{
			Code:    resp.Status,
			Message: "token is temporarily over quota, failed more than max retries",
		}
	case http.StatusServiceUnavailable:
		err = &genericRespError{
			Code:    resp.Status,
			Message: "server is temporarily unavailable to accept writes, failed more than max retries",
		}
		retryAfter := resp.Header.Get("Retry-After")
		retry, _ := strconv.Atoi(retryAfter) // we ignore the error here because an error already means retry is 0.
		retryTime := time.Duration(retry) * time.Second
		if retry == 0 { // if we didn't get a Retry-After or it is zero, instead switch to exponential backoff
			retryTime = time.Duration(rand.Int63n(((1 << tries) - 1) * 10 * int64(time.Microsecond)))
		}
		if retryTime > defaultMaxWait {
			retryTime = defaultMaxWait
		}
		time.Sleep(time.Duration(retry) * time.Second)
		resp.Body.Close()
		if c.maxRetries == -1 || int(tries) < c.maxRetries {
			tries++
			goto doRequest
		}
		return err
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

func (c *Client) makeWriteRequest(bucket, org string, body io.Reader) (*http.Request, error) {
	var err error
	if c.contentEncoding == "gzip" {
		body, err = gzip.CompressWithGzip(body)
		if err != nil {
			return nil, err
		}
	}
	url, err := makeWriteURL(c.url, org, bucket)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", url, body)
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
	params := url.Values{}
	params.Set("bucket", bucket)
	params.Set("org", org)

	switch loc.Scheme {
	case "http", "https":
		loc.Path = path.Join(loc.Path, "/api/v2/write")
	case "unix":
	default:
		return "", fmt.Errorf("unsupported scheme: %q", loc.Scheme)
	}
	loc.RawQuery = params.Encode()
	return loc.String(), nil
}
