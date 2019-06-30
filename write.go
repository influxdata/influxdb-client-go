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
func (c *Client) Write(ctx context.Context, bucket, org string, m...Metric) (err error) {
	tries := uint64(0)
	return c.write(ctx, bucket, org, &tries, m...)
}

func parseWriteError(r io.Reader) (*genericRespError, error) {
	werr := &genericRespError{}
	if err := json.NewDecoder(r).Decode(&werr); err != nil {
		return nil, err
	}
	return werr, nil
}

func (c *Client) write(ctx context.Context, bucket, org string, triesPtr *uint64, m ...Metric) error {
	buf := &bytes.Buffer{}
	e := lp.NewEncoder(buf)
	cleanup := func() {}
	defer func() { cleanup() }()
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
	cleanup = func() {
		r := io.LimitReader(resp.Body, 1<<16) // we limit it because it is usually better to just reuse the body, but sometimes it isn't worth it.
		// throw away the rest of the body so the connection can be reused even if there is still stuff on the wire.
		ioutil.ReadAll(r) // we don't care about the error here, it is just to empty the tcp buffer
		resp.Body.Close()
	}

	switch resp.StatusCode {
	case http.StatusOK, http.StatusNoContent:
	case http.StatusTooManyRequests:
		err = &genericRespError{
			Code:    resp.Status,
			Message: "too many requests too fast",
		}
		cleanup()
		if err2 := c.backoff(triesPtr, resp, err); err2 != nil {
			return err2
		}
		cleanup = func() {}
		goto doRequest
	case http.StatusServiceUnavailable:
		err = &genericRespError{
			Code:    resp.Status,
			Message: "service temporarily unavaliable",
		}
		cleanup()
		if err2 := c.backoff(triesPtr, resp, err); err2 != nil {
			return err2
		}
		cleanup = func() {}
		goto doRequest
	default:
		gwerr, err := parseWriteError(resp.Body)
		if err != nil {
			return err
		}

		return gwerr
	}
	// we don't defer and close till here, because of the retries.
	defer func() {
		r := io.LimitReader(resp.Body, 1<<16) // we limit it because it is usually better to just reuse the body, but sometimes it isn't worth it.
		_, err := ioutil.ReadAll(r)           // throw away the rest of the body so the connection gets reused.
		err2 := resp.Body.Close()
		if err == nil && err2 != nil {
			err = err2
		}
	}()
	e.FailOnFieldErr(c.errOnFieldErr)
	return err
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
