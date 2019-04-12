package influxdb

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
)

// QueryCSV returns the result of a flux query.
// TODO: annotations optionally
func (c *Client) QueryCSV(flux string, org string) (io.ReadCloser, error) {
	qURL, err := c.makeQueryURL(org)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", qURL, bytes.NewBufferString(flux))
	if err != nil {
		return nil, err
	}
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
		if resp.ContentLength > 0 {
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
	cleanup = func() {} // we don't want to close the body  if we got a status code in the 2xx range.
	return resp.Body, ErrUnimplemented
}

func (c *Client) makeQueryURL(org string) (string, error) {
	qu, err := url.Parse(c.url.String())
	if err != nil {
		return "", err
	}
	qu.Path = path.Join(qu.Path, "query")

	params := qu.Query()
	params.Set("org", org)
	return qu.String(), nil
}
