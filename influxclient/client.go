// Copyright 2021 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

// Package influxclient provides client for InfluxDB server.
package influxclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"net/url"
	"strings"

	"github.com/influxdata/influxdb-client-go/influxclient/model"
)

const (
	DefaultBatchSize = 5000
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

	// Organization is name or ID of organization where data (buckets, users, tasks, etc.) belongs to
	Organization string

	// BatchSize holds the default batch size used by PointWriter.
	// If it's zero, DefaultBatchSize will be used.
	// Note that this can be overridden with PointWriter.SetBatchSize.
	BatchSize int

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
	// generated server client
	apiClient *model.Client
}

// httpParams holds parameters for creating an HTTP request
type httpParams struct {
	// URL of server endpoint
	endpointURL *url.URL
	// Params to be added to URL
	queryParams url.Values
	// HTTP request method, eg. POST
	httpMethod string
	// HTTP request headers
	headers http.Header
	// HTTP POST/PUT body
	body io.Reader
}

// apiCallDelegate delegates generated API client calls to client
type apiCallDelegate struct {
	// Client
	c *Client
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
	c.apiURL, err = url.Parse(serverAddress)
	if err != nil {
		return nil, fmt.Errorf("error parsing server URL: %w", err)
	}
	// Create API client
	c.apiClient, err = model.NewClient(c.apiURL.String(), &apiCallDelegate{c})
	if err != nil {
		return nil, fmt.Errorf("error creating server API client: %w", err)
	}
	// Update server API URL
	c.apiURL, err = url.Parse(c.apiClient.APIEndpoint)
	if err != nil {
		return nil, fmt.Errorf("error parsing API endpoint URL: %w", err)
	}
	return c, nil
}

// APIClient returns generates API client
func (c *Client) APIClient() *model.Client {
	return c.apiClient
}

// makeAPICall issues an HTTP request to InfluxDB server API url according to parameters.
// Additionally, sets Authorization header and User-Agent.
// It returns http.Response or error. Error can be a *ServerError if server responded with error.
func (c *Client) makeAPICall(ctx context.Context, params httpParams) (*http.Response, error) {
	// copy URL
	urlObj := *params.endpointURL
	urlObj.RawQuery = params.queryParams.Encode()

	fullURL := urlObj.String()

	req, err := http.NewRequestWithContext(ctx, params.httpMethod, fullURL, params.body)
	if err != nil {
		return nil, fmt.Errorf("error calling %s: %v", fullURL, err)
	}
	for k, v := range params.headers {
		for _, i := range v {
			req.Header.Add(k, i)
		}
	}
	req.Header.Set("User-Agent", userAgent)
	if c.authorization != "" {
		req.Header.Add("Authorization", c.authorization)
	}

	resp, err := c.params.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error calling %s: %v", fullURL, err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, c.resolveHTTPError(resp)
	}

	return resp, nil
}

// resolveHTTPError parses server error response and returns error with human-readable message
func (c *Client) resolveHTTPError(r *http.Response) error {
	// successful status code range
	if r.StatusCode >= 200 && r.StatusCode < 300 {
		return nil
	}

	var httpError struct {
		ServerError
		// Error message of InfluxDB 1 error
		Error string `json:"error"`
	}

	httpError.StatusCode = r.StatusCode

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		httpError.Message = fmt.Sprintf("cannot read error response:: %v", err)
	}
	ctype, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if ctype == "application/json" {
		err := json.Unmarshal(body, &httpError)
		if err != nil {
			httpError.Message = fmt.Sprintf("cannot decode error response: %v", err)
		}
		if httpError.Message == "" && httpError.Code == "" {
			httpError.Message = httpError.Error
		}
	}
	if httpError.Message == "" {
		//TODO: "This could be a large piece of unreadable body; we might be able to do better than this"
		if len(body) > 0 {
			httpError.Message = string(body)
		} else {
			httpError.Message = r.Status
		}
	}

	return &httpError.ServerError
}

// Do makes API call for generated client
func (d *apiCallDelegate) Do(req *http.Request) (*http.Response, error) {
	queryParams := req.URL.Query()
	req.URL.RawQuery = ""
	return d.c.makeAPICall(req.Context(), httpParams{
		endpointURL: req.URL,
		headers: req.Header,
		httpMethod: req.Method,
		body: req.Body,
		queryParams: queryParams,
	})
}

// AuthorizationsAPI returns a value that can be used to interact with the
// authorization-related parts of the InfluxDB API.
func (c *Client) AuthorizationsAPI() *AuthorizationsAPI {
	return newAuthorizationsAPI(c.apiClient)
}

// BucketsAPI returns a value that can be used to interact with the
// bucket-related parts of the InfluxDB API.
func (c *Client) BucketsAPI() *BucketsAPI {
	return newBucketsAPI(c.apiClient)
}

// LabelsAPI returns a value that can be used to interact with the
// label-related parts of the InfluxDB API.
func (c *Client) LabelsAPI() *LabelsAPI {
	return newLabelsAPI(c.apiClient)
}

// OrganizationAPI returns a value that can be used to interact with the
// organization-related parts of the InfluxDB API.
func (c *Client) OrganizationAPI() *OrganizationAPI {
	return newOrganizationAPI(c.apiClient)
}

// TasksAPI returns a value that can be used to interact with the
// task-related parts of the InfluxDB API.
func (c *Client) TasksAPI() *TasksAPI {
	return newTasksAPI(c.apiClient)
}

// UsersAPI returns a value that can be used to interact with the
// user-related parts of the InfluxDB API.
func (c *Client) UsersAPI() *UsersAPI {
	return newUsersAPI(c.apiClient)
}
