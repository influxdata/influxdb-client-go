package influxdb

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"net/url"
	"path"
	"strconv"

	"github.com/influxdata/influxdb-client-go/internal/gzip"
	lp "github.com/influxdata/line-protocol"
)

// Write writes metrics to a bucket, and org.  It retries intelligently.
// If the write is too big, it retries again, after breaking the payloads into two requests.
func (c *Client) Write(ctx context.Context, bucket, org string, m ...Metric) (n int, err error) {
	var (
		buf = &bytes.Buffer{}
		e   = lp.NewEncoder(buf)
	)

	e.FailOnFieldErr(c.errOnFieldErr)

	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	default:
	}

	for i := range m {
		if _, err := e.Encode(m[i]); err != nil {
			return 0, err
		}
	}

	req, err := c.makeWriteRequest(bucket, org, buf)
	if err != nil {
		return 0, err
	}

	resp, err := c.httpClient.Do(req.WithContext(ctx))
	if err != nil {
		return 0, err
	}

	defer func() {
		// discard body so connection can be reused
		_, _ = io.Copy(ioutil.Discard, resp.Body)
		resp.Body.Close()
	}()

	eerr, err := parseWriteError(resp)
	if err != nil {
		return 0, err
	}

	if eerr != nil {
		return 0, eerr
	}

	return len(m), nil
}

func parseInt32(v string) (int32, error) {
	retry, err := strconv.ParseInt(v, 10, 32)
	if err != nil {
		return 0, err
	}

	return int32(retry), nil
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
	u.RawQuery = params.Encode()

	switch loc.Scheme {
	case "http", "https":
		u.Path = path.Join(u.Path, "/write")
	case "unix":
	default:
		return "", fmt.Errorf("unsupported scheme: %q", u.Scheme)
	}

	return u.String(), nil
}

func parseWriteError(r *http.Response) (err *Error, perr error) {
	// successful status code range
	if r.StatusCode >= 200 && r.StatusCode < 300 {
		return nil, nil
	}

	err = &Error{}
	if v := r.Header.Get("Retry-After"); v != "" {
		if retry, perr := parseInt32(v); perr == nil {
			err.RetryAfter = &retry
		}
	}

	switch r.StatusCode {
	case http.StatusTooManyRequests:
		err.Code = ETooManyRequests
		err.Message = "exceeded rate limit"
	case http.StatusServiceUnavailable:
		err.Code = EUnavailable
		err.Message = "service temporarily unavailable"
	default:
		// json encoded error
		typ, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if typ == "application/json" {
			perr = json.NewDecoder(r.Body).Decode(err)
			return
		}

		// plain error body type
		var body []byte
		body, perr = ioutil.ReadAll(r.Body)
		if perr != nil {
			return
		}

		err.Code = r.Status
		err.Message = string(body)
	}

	return
}
