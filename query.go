package influxdb

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
)

// QueryCSV returns the result of a flux query.
// TODO: annotations optionally
func (c *Client) QueryCSV(ctx context.Context, flux string, org string) (*QueryCSVResult, error) {
	qURL, err := c.makeQueryURL(org)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", qURL, bytes.NewBufferString(flux))
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Authorization", c.authorization)
	req.Header.Set("Content-Type", "application/vnd.flux")
	resp, err := c.httpClient.Do(req)
	// this is so we can unset the defer later if we don't error.
	cleanup := func() {
		resp.Body.Close()
	}
	defer func() { cleanup() }()
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		r := io.LimitReader(resp.Body, 1<<14) // only support errors that are 16kB long, more than that and something is probably wrong.
		gerr := &genericRespError{Code: resp.Status}
		if resp.ContentLength != 0 {
			if err := json.NewDecoder(r).Decode(gerr); err != nil {
				gerr.Code = resp.Status
				message, err := ioutil.ReadAll(r)
				if err != nil {
					return nil, err
				}
				gerr.Message = string(message)
			}
		}
		return nil, gerr
	}
	cleanup = func() {} // we don't want to close the body if we got a status code in the 2xx range.
	return &QueryCSVResult{ReadCloser: resp.Body}, nil
}

func (c *Client) makeQueryURL(org string) (string, error) {
	qu, err := url.Parse(c.url.String())
	if err != nil {
		return "", err
	}
	qu.Path = path.Join(qu.Path, "query")
	fmt.Println(qu.Path)

	params := qu.Query()
	params.Set("org", org)
	qu.RawQuery = params.Encode()
	fmt.Println(qu.String())
	return qu.String(), nil
}

// QueryCSVResult is the result of a flux query in CSV format
type QueryCSVResult struct {
	io.ReadCloser
}
