package client

import (
	"errors"
	"fmt"
	"io"
	"net/url"
	"path"
)

// QueryCSV returns the result of a flux query.
// TODO: annotations optionally
func (c *Client) QueryCSV(flux string) (io.Reader, error) {
	return nil, ErrUnimplemented
}

func makeQueryURL(loc *url.URL) (string, error) {
	if loc == nil {
		return "", errors.New("nil url")
	}
	params := url.Values{}
	// params.Set("bucket", bucket)
	// params.Set("org", org)

	switch loc.Scheme {
	case "http", "https":
		loc.Path = path.Join(loc.Path, "/api/v2/query")
	case "unix":
	default:
		return "", fmt.Errorf("unsupported scheme: %q", loc.Scheme)
	}
	loc.RawQuery = params.Encode()
	return loc.String(), nil
}
