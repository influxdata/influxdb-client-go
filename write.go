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

// Write writes metrics to a bucket, and org. The result n is the number of points written.
func (c *Client) Write(ctx context.Context, bucket, org string, m ...Metric) (n int, err error) {
	var (
		buf = &bytes.Buffer{}
		e   = lp.NewEncoder(buf)
	)

	e.SetFieldTypeSupport(lp.UintSupport)
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

	var req *http.Request

	if c.contentEncoding == "gzip" {
		req, err = NewWriteGzipRequest(c.url, c.userAgent, c.authorization, bucket, org, c.compressionLevel, buf)
	} else {
		req, err = NewWriteRequest(c.url, c.userAgent, c.authorization, bucket, org, buf)
	}
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
		_ = resp.Body.Close()
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

func NewWriteGzipRequest(url *url.URL, userAgent, token, bucket, org string, compressionLevel int, body io.Reader) (*http.Request, error) {
	body, err := gzip.CompressWithGzip(body, compressionLevel)
	if err != nil {
		return nil, err
	}

	u, err := makeWriteURL(url, bucket, org)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, u, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "text/plain; charset=utf-8")
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Authorization", token)
	return req, nil
}

func NewWriteRequest(url *url.URL, userAgent, token, bucket, org string, body io.Reader) (*http.Request, error) {
	u, err := makeWriteURL(url, bucket, org)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, u, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "text/plain; charset=utf-8")
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Authorization", token)
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
	err.StatusCode = r.StatusCode
	switch r.StatusCode {
	case http.StatusRequestEntityTooLarge:
		err.Code = ETooLarge
		err.Message = "tried to write too large a batch"
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
