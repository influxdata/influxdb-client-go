// Copyright 2021 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

// Package influxclient provides client for InfluxDB server.
package influxclient

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// Params holds the parameters for creating a new client.
// The only mandatory field is ServerURL. AuthToken is also important
// if authentication was not done outside this client.
type Params struct {
	// ServerURL holds the URL of the InfluxDB server to connect to.
	// This must be non-empty. E.g. http://localhost:8086
	ServerURL string

	// AuthToken holds the authorization token for the API.
	// This can be obtained through the GUI web browser interface.
	AuthToken string

	// HTTPClient is used to make API requests.
	//
	// This can be used to specify a custom TLS configuration
	// (TLSClientConfig), a custom request timeout (Timeout),
	// or other customization as required.
	//
	// It HTTPClient is nil, http.DefaultClient will be used.
	HTTPClient *http.Client
}

// Client implements an InfluxDB client.
type Client struct {
	// Configuration params.
	params Params
	// Pre-created Authorization HTTP header value.
	authorization string
	// Cached base server API URL.
	apiURL *url.URL
}

// New creates new Client with given Params, where ServerURL and AuthToken are mandatory.
func New(params Params) (*Client, error) {
	c := &Client{params: params}
	if params.ServerURL == "" {
		return nil, errors.New("empty server URL")
	}
	if c.params.AuthToken != "" {
		c.authorization = "Token " + c.params.AuthToken
	}
	if c.params.HTTPClient == nil {
		c.params.HTTPClient = http.DefaultClient
	}

	serverAddress := params.ServerURL
	if !strings.HasSuffix(serverAddress, "/") {
		// For subsequent path parts concatenation, url has to end with '/'
		serverAddress = params.ServerURL + "/"
	}
	var err error
	// Prepare server API URL
	c.apiURL, err = url.Parse(serverAddress + "api/v2/")
	if err != nil {
		return nil, fmt.Errorf("error parsing server URL: %w", err)
	}
	return c, nil
}
