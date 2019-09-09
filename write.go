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

	switch resp.StatusCode {
	case http.StatusOK, http.StatusNoContent:
	case http.StatusTooManyRequests:
		return 0, Error{
			Code:    ETooManyRequests,
			Message: "exceeded rate limit",
		}
	case http.StatusServiceUnavailable:
		return 0, Error{
			Code:    EUnavailable,
			Message: "service temporarily unavaliable",
		}
	default:
		werr, err := parseWriteError(resp)
		if err != nil {
			return 0, err
		}

		return 0, werr
	}

	return len(m), nil
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

func parseWriteError(r *http.Response) (err Error, eerr error) {
	typ, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if typ == "application/json" {
		eerr = json.NewDecoder(r.Body).Decode(&err)
		return
	}

	var body []byte
	body, eerr = ioutil.ReadAll(r.Body)
	if eerr != nil {
		return
	}

	return Error{Code: r.Status, Message: string(body)}, nil
}
