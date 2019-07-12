package influxdb

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"

	lp "github.com/influxdata/line-protocol"
)

// Write writes metrics to a bucket, and org.  It retries intelligently.
// If the write is too big, it retries again, after breaking the payloads into two requests.
func (c *Client) Write(ctx context.Context, org Organisation, bucket Bucket, m ...Metric) (n int, err error) {
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

	req, err := c.makeWriteRequest(string(bucket), string(org), buf)
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

	switch resp.StatusCode {
	case http.StatusOK, http.StatusNoContent:
	case http.StatusTooManyRequests:
		err = &genericRespError{
			Code:    resp.Status,
			Message: "too many requests too fast",
		}
	case http.StatusServiceUnavailable:
		err = &genericRespError{
			Code:    resp.Status,
			Message: "service temporarily unavaliable",
		}
	default:
		gwerr, err := parseWriteError(resp.Body)
		if err != nil {
			return 0, err
		}

		return 0, gwerr
	}

	return len(m), err
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

func parseWriteError(r io.Reader) (*genericRespError, error) {
	werr := &genericRespError{}
	if err := json.NewDecoder(r).Decode(&werr); err != nil {
		return nil, err
	}
	return werr, nil
}
