package client

import (
	"context"
	"fmt"
	"io"
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

type Client struct {
	rand            *rand.Rand
	httpClient      *http.Client
	contentEncoding string
	transport       *http.Transport
	url             url.URL
	password        string
	username        string
	token           string
	org             string
	maxRetries      uint
	errOnFieldErr   bool
}

type Batch struct {
	lp.Encoder
}

// WithPasswordAndUser will allow the Client to generate a token or session from a username and pass when it needs one.
func (c *Client) WithPasswordAndUser(password, username string) {
	panic("NOT IMPLEMENTED")
}

// Ping checks that status of cluster, and will always return 0 time and no
// error for UDP clients.
func (c *Client) Ping(timeout time.Duration) (time.Duration, string, error) {
	panic("NOT IMPLEMENTED")

}

func (c *Client) Write(ctx context.Context, bucket string, m ...lp.Metric) (err error) {
	r, w := io.Pipe()
	e := lp.NewEncoder(w)
	req, err := c.makeWriteRequest(r)
	if err != nil {
		return err
	}
	tries := uint(0)
doRequest:
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		err2 := resp.Body.Close()
		if err == nil && err2 != nil {
			err = err2
		}
	}()
	switch resp.StatusCode {
	case http.StatusBadRequest,
		http.StatusUnauthorized,
		http.StatusForbidden,
		http.StatusRequestEntityTooLarge:
	case http.StatusTooManyRequests, http.StatusServiceUnavailable:
		retryAfter := resp.Header.Get("Retry-After")
		retry, _ := strconv.Atoi(retryAfter) // we ignore the error here because an error already means retry is 0.
		retryTime := time.Duration(retry) * time.Second
		if retry == 0 { // if we didn't get a Retry-After or it is zero, instead switch to exponential backoff
			retryTime = time.Duration(c.rand.Int63n(((1 << tries) - 1) * 10 * int(time.Microsecond)))
		}
		if retryTime > defaultMaxWait {
			retryTime = defaultMaxWait
		}
		time.Sleep(time.Duration(retry) * time.Second)
		if c.maxRetries == -1 || tries < c.maxRetries {
			tries++
			goto doRequest
		}
	}

	e.FailOnFieldErr(c.errOnFieldErr)
	for i := range m {
		if _, err := e.Encode(m[i]); err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) makeWriteRequest(body io.Reader) (*http.Request, error) {
	var err error
	if c.contentEncoding == "gzip" {
		body, err = gzip.CompressWithGzip(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest("POST", c.url.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "text/plain; charset=utf-8")

	if c.contentEncoding == "gzip" {
		req.Header.Set("Content-Encoding", "gzip")
	}

	return req, nil
}

// func (c *Client) writeBatch(ctx context.Context, bucket string, metrics... lp.Metric) error {
// 	url, err := makeWriteURL(c.url, c.org, bucket)
// 	if err != nil {
// 		return err
// 	}

// 	reader := influx.NewReader(metrics, c.serializer)
// 	req, err := c.makeWriteRequest(url, reader)
// 	if err != nil {
// 		return err
// 	}

// 	resp, err := c.client.Do(req.WithContext(ctx))
// 	if err != nil {
// 		return err
// 	}
// 	defer resp.Body.Close()

// 	if resp.StatusCode == http.StatusNoContent {
// 		return nil
// 	}

// 	writeResp := &genericRespError{}
// 	err = json.NewDecoder(resp.Body).Decode(writeResp)
// 	desc := writeResp.Error()
// 	if err != nil {
// 		desc = resp.Status
// 	}

// 	switch resp.StatusCode {
// 	case http.StatusBadRequest, http.StatusUnauthorized,
// 		http.StatusForbidden, http.StatusRequestEntityTooLarge:
// 		log.Printf("E! [outputs.influxdb_v2] Failed to write metric: %s\n", desc)
// 		return nil
// 	case http.StatusTooManyRequests, http.StatusServiceUnavailable:
// 		retryAfter := resp.Header.Get("Retry-After")
// 		retry, err := strconv.Atoi(retryAfter)
// 		if err != nil {
// 			retry = 0
// 		}
// 		if retry > defaultMaxWait {
// 			retry = defaultMaxWait
// 		}
// 		c.retryTime = time.Now().Add(time.Duration(retry) * time.Second)
// 		return fmt.Errorf("Waiting %ds for server before sending metric again", retry)
// 	}

// 	// This is only until platform spec is fully implemented. As of the
// 	// time of writing, there is no error body returned.
// 	if xErr := resp.Header.Get("X-Influx-Error"); xErr != "" {
// 		desc = fmt.Sprintf("%s; %s", desc, xErr)
// 	}

// 	return &APIError{
// 		StatusCode:  resp.StatusCode,
// 		Title:       resp.Status,
// 		Description: desc,
// 	}
// }

// func (c *Client) addHeaders(req *http.Request) {
// 	for header, value := range c.Headers {
// 		req.Header.Set(header, value)
// 	}
// }

func makeWriteURL(loc url.URL, org, bucket string) (string, error) {
	params := url.Values{}
	params.Set("bucket", bucket)
	params.Set("org", org)

	switch loc.Scheme {
	case "http", "https":
		loc.Path = path.Join(loc.Path, "/api/v2/write")
	default:
		return "", fmt.Errorf("unsupported scheme: %q", loc.Scheme)
	}
	loc.RawQuery = params.Encode()
	return loc.String(), nil
}
