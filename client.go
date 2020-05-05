// Copyright 2020 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

// Package influxdb2 provides API for using InfluxDB client in Go.
// It's intended to use with InfluxDB 2 server. WriteApi, QueryApi and Health work also with InfluxDB 1.8
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

// Client provides API to communicate with InfluxDBServer.
// There two APIs for writing, WriteApi and WriteApiBlocking.
// WriteApi provides asynchronous, non-blocking, methods for writing time series data.
// WriteApiBlocking provides blocking methods for writing time series data.
type Client interface {
	// WriteApi returns the asynchronous, non-blocking, Write client
	WriteApi(org, bucket string) WriteApi
	// WriteApi returns the synchronous, blocking, Write client
	WriteApiBlocking(org, bucket string) WriteApiBlocking
	// QueryApi returns Query client
	QueryApi(org string) QueryApi
	// AuthorizationsApi returns Authorizations API client
	AuthorizationsApi() api.AuthorizationsApi
	// OrganizationsApi returns Organizations API client
	OrganizationsApi() api.OrganizationsApi
	// UsersApi returns Users API client
	UsersApi() api.UsersApi
	// Close ensures all ongoing asynchronous write clients finish
	Close()
	// Options returns the options associated with client
	Options() *Options
	// ServerUrl returns the url of the server url client talks to
	ServerUrl() string
	// Setup sends request to initialise new InfluxDB server with user, org and bucket, and data retention period
	// and returns details about newly created entities along with the authorization object.
	// Retention period of zero will result to infinite retention.
	Setup(ctx context.Context, username, password, org, bucket string, retentionPeriodHours int) (*domain.OnboardingResponse, error)
	// Ready checks InfluxDB server is running. It doesn't validate authentication params.
	Ready(ctx context.Context) (bool, error)
	// Health returns an InfluxDB server health check result. Read the HealthCheck.Status field to get server status.
	// Health doesn't validate authentication params.
	Health(ctx context.Context) (*domain.HealthCheck, error)
}

// clientImpl implements Client interface
type clientImpl struct {
	serverUrl   string
	options     *Options
	writeApis   []WriteApi
	lock        sync.Mutex
	httpService ihttp.Service
	apiClient   *domain.ClientWithResponses
	authApi     api.AuthorizationsApi
	orgApi      api.OrganizationsApi
	usersApi    api.UsersApi
}

// NewClient creates Client for connecting to given serverUrl with provided authentication token, with the default options.
// Authentication token can be empty in case of connecting to newly installed InfluxDB server, which has not been set up yet.
// In such case Setup will set authentication token
func NewClient(serverUrl string, authToken string) Client {
	return NewClientWithOptions(serverUrl, authToken, DefaultOptions())
}

// NewClientWithOptions creates Client for connecting to given serverUrl with provided authentication token
// and configured with custom Options
// Authentication token can be empty in case of connecting to newly installed InfluxDB server, which has not been set up yet.
// In such case Setup will set authentication token
func NewClientWithOptions(serverUrl string, authToken string, options *Options) Client {
	service := ihttp.NewService(serverUrl, "Token "+authToken, options.tlsConfig, options.httpRequestTimeout)
	client := &clientImpl{
		serverUrl:   serverUrl,
		options:     options,
		writeApis:   make([]WriteApi, 0, 5),
		httpService: service,
		apiClient:   domain.NewClientWithResponses(service),
	}
	return client
}
func (c *clientImpl) Options() *Options {
	return c.options
}

func (c *clientImpl) ServerUrl() string {
	return c.serverUrl
}

func (c *clientImpl) Ready(ctx context.Context) (bool, error) {
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

func (c *clientImpl) WriteApi(org, bucket string) WriteApi {
	w := newWriteApiImpl(org, bucket, c.httpService, c)
	c.writeApis = append(c.writeApis, w)
	return w
}

func (c *clientImpl) WriteApiBlocking(org, bucket string) WriteApiBlocking {
	w := newWriteApiBlockingImpl(org, bucket, c.httpService, c)
	return w
}

func (c *clientImpl) Close() {
	for _, w := range c.writeApis {
		w.Close()
	}
}

func (c *clientImpl) QueryApi(org string) QueryApi {
	return newQueryApi(org, c.httpService, c)
}

func (c *clientImpl) AuthorizationsApi() api.AuthorizationsApi {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.authApi == nil {
		c.authApi = api.NewAuthorizationApi(c.httpService)
	}
	return c.authApi
}

func (c *clientImpl) OrganizationsApi() api.OrganizationsApi {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.orgApi == nil {
		c.orgApi = api.NewOrganizationsApi(c.httpService)
	}
	return c.orgApi
}

func (c *clientImpl) UsersApi() api.UsersApi {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.usersApi == nil {
		c.usersApi = api.NewUsersApi(c.httpService)
	}
	return c.usersApi
}
