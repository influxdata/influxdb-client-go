// Copyright 2020 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

// package influxdb2 provides API for using InfluxDB client in Go
// It's intended to use with InfluxDB 2 server
package influxdb2

import (
	"context"
	"github.com/influxdata/influxdb-client-go/api"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"sync"

	"github.com/influxdata/influxdb-client-go/domain"
	ihttp "github.com/influxdata/influxdb-client-go/internal/http"
)

// InfluxDBClient provides API to communicate with InfluxDBServer
// There two APIs for writing, WriteApi and WriteApiBlocking.
// WriteApi provides asynchronous, non-blocking, methods for writing time series data.
// WriteApiBlocking provides blocking methods for writing time series data
type InfluxDBClient interface {
	// WriteApi returns the asynchronous, non-blocking, Write client.
	WriteApi(org, bucket string) WriteApi
	// WriteApi returns the synchronous, blocking, Write client.
	WriteApiBlocking(org, bucket string) WriteApiBlocking
	// QueryApi returns Query client
	QueryApi(org string) QueryApi
	// AuthorizationsApi returns Authorizations API client
	AuthorizationsApi() api.AuthorizationsApi
	// OrganizationsApi returns Organizations API client
	OrganizationsApi() api.OrganizationsApi
	// Close ensures all ongoing asynchronous write clients finish
	Close()
	// Options returns the options associated with client
	Options() *Options
	// serverUrl returns the url of the server url client talks to
	ServerUrl() string
	// Setup sends request to initialise new InfluxDB server with user, org and bucket, and data retention period
	// Retention period of zero will result to infinite retention
	// and returns details about newly created entities along with the authorization object
	Setup(ctx context.Context, username, password, org, bucket string, retentionPeriodHours int) (*domain.OnboardingResponse, error)
	// Ready checks InfluxDB server is running
	Ready(ctx context.Context) (bool, error)
}

// client implements InfluxDBClient interface
type client struct {
	serverUrl   string
	options     *Options
	writeApis   []WriteApi
	lock        sync.Mutex
	httpService ihttp.Service
}

// NewClient creates InfluxDBClient for connecting to given serverUrl with provided authentication token, with default options
// Authentication token can be empty in case of connecting to newly installed InfluxDB server, which has not been set up yet.
// In such case Setup will set authentication token
func NewClient(serverUrl string, authToken string) InfluxDBClient {
	return NewClientWithOptions(serverUrl, authToken, DefaultOptions())
}

// NewClientWithOptions creates InfluxDBClient for connecting to given serverUrl with provided authentication token
// and configured with custom Options
// Authentication token can be empty in case of connecting to newly installed InfluxDB server, which has not been set up yet.
// In such case Setup will set authentication token
func NewClientWithOptions(serverUrl string, authToken string, options *Options) InfluxDBClient {
	client := &client{
		serverUrl:   serverUrl,
		options:     options,
		writeApis:   make([]WriteApi, 0, 5),
		httpService: ihttp.NewService(serverUrl, "Token "+authToken, options.tlsConfig, options.httpRequestTimeout),
	}
	return client
}
func (c *client) Options() *Options {
	return c.options
}

func (c *client) ServerUrl() string {
	return c.serverUrl
}

func (c *client) Ready(ctx context.Context) (bool, error) {
	readyUrl, err := url.Parse(c.serverUrl)
	if err != nil {
		return false, err
	}
	readyUrl.Path = path.Join(readyUrl.Path, "ready")
	readyRes := false
	perror := c.httpService.GetRequest(ctx, readyUrl.String(), nil, func(resp *http.Response) error {
		// discard body so connection can be reused
		_, _ = io.Copy(ioutil.Discard, resp.Body)
		_ = resp.Body.Close()
		readyRes = resp.StatusCode == http.StatusOK
		return nil
	})
	if perror != nil {
		return false, perror
	}
	return readyRes, nil
}

func (c *client) WriteApi(org, bucket string) WriteApi {
	w := newWriteApiImpl(org, bucket, c.httpService, c)
	c.writeApis = append(c.writeApis, w)
	return w
}

func (c *client) WriteApiBlocking(org, bucket string) WriteApiBlocking {
	w := newWriteApiBlockingImpl(org, bucket, c.httpService, c)
	return w
}

func (c *client) Close() {
	for _, w := range c.writeApis {
		w.Close()
	}
}

func (c *client) QueryApi(org string) QueryApi {
	return newQueryApi(org, c.httpService, c)
}

func (c *client) AuthorizationsApi() api.AuthorizationsApi {
	return api.NewAuthorizationApi(c.httpService)
}

func (c *client) OrganizationsApi() api.OrganizationsApi {
	return api.NewOrganizationsApi(c.httpService)
}
