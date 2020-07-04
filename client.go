// Copyright 2020 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

// Package influxdb2 provides API for using InfluxDB client in Go.
// It's intended to use with InfluxDB 2 server. WriteApi, QueryApi and Health work also with InfluxDB 1.8
package influxdb2

import (
	"context"
	"errors"
	"sync"

	"github.com/influxdata/influxdb-client-go/api"
	"github.com/influxdata/influxdb-client-go/api/log"

	"github.com/influxdata/influxdb-client-go/domain"
	ihttp "github.com/influxdata/influxdb-client-go/internal/http"
)

// Client provides API to communicate with InfluxDBServer.
// There two APIs for writing, WriteApi and WriteApiBlocking.
// WriteApi provides asynchronous, non-blocking, methods for writing time series data.
// WriteApiBlocking provides blocking methods for writing time series data.
type Client interface {
	// Setup sends request to initialise new InfluxDB server with user, org and bucket, and data retention period
	// and returns details about newly created entities along with the authorization object.
	// Retention period of zero will result to infinite retention.
	Setup(ctx context.Context, username, password, org, bucket string, retentionPeriodHours int) (*domain.OnboardingResponse, error)
	// Ready checks InfluxDB server is running. It doesn't validate authentication params.
	Ready(ctx context.Context) (bool, error)
	// Health returns an InfluxDB server health check result. Read the HealthCheck.Status field to get server status.
	// Health doesn't validate authentication params.
	Health(ctx context.Context) (*domain.HealthCheck, error)
	// Close ensures all ongoing asynchronous write clients finish
	Close()
	// Options returns the options associated with client
	Options() *Options
	// ServerUrl returns the url of the server url client talks to
	ServerUrl() string
	// WriteApi returns the asynchronous, non-blocking, Write client
	WriteApi(org, bucket string) api.WriteApi
	// WriteApi returns the synchronous, blocking, Write client
	WriteApiBlocking(org, bucket string) api.WriteApiBlocking
	// QueryApi returns Query client
	QueryApi(org string) api.QueryApi
	// AuthorizationsApi returns Authorizations API client
	AuthorizationsApi() api.AuthorizationsApi
	// OrganizationsApi returns Organizations API client
	OrganizationsApi() api.OrganizationsApi
	// UsersApi returns Users API client
	UsersApi() api.UsersApi
	// DeleteApi returns Delete API client
	DeleteApi() api.DeleteApi
	// BucketsApi returns Buckets API client
	BucketsApi() api.BucketsApi
	// LabelsApi returns Labels API client
	LabelsApi() api.LabelsApi
}

// clientImpl implements Client interface
type clientImpl struct {
	serverUrl   string
	options     *Options
	writeApis   []api.WriteApi
	lock        sync.Mutex
	httpService ihttp.Service
	apiClient   *domain.ClientWithResponses
	authApi     api.AuthorizationsApi
	orgApi      api.OrganizationsApi
	usersApi    api.UsersApi
	deleteApi   api.DeleteApi
	bucketsApi  api.BucketsApi
	labelsApi   api.LabelsApi
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
	service := ihttp.NewService(serverUrl, "Token "+authToken, options.httpOptions)
	client := &clientImpl{
		serverUrl:   serverUrl,
		options:     options,
		writeApis:   make([]api.WriteApi, 0, 5),
		httpService: service,
		apiClient:   domain.NewClientWithResponses(service),
	}
	log.Log.SetDebugLevel(client.Options().LogLevel())
	return client
}
func (c *clientImpl) Options() *Options {
	return c.options
}

func (c *clientImpl) ServerUrl() string {
	return c.serverUrl
}

func (c *clientImpl) Ready(ctx context.Context) (bool, error) {
	params := &domain.GetReadyParams{}
	response, err := c.apiClient.GetReadyWithResponse(ctx, params)
	if err != nil {
		return false, err
	}
	if response.JSONDefault != nil {
		return false, domain.DomainErrorToError(response.JSONDefault, response.StatusCode())
	}
	return true, nil
}

func (c *clientImpl) Setup(ctx context.Context, username, password, org, bucket string, retentionPeriodHours int) (*domain.OnboardingResponse, error) {
	if username == "" || password == "" {
		return nil, errors.New("a username and password is required for a setup")
	}
	c.lock.Lock()
	defer c.lock.Unlock()
	params := &domain.PostSetupParams{}
	body := &domain.PostSetupJSONRequestBody{
		Bucket:             bucket,
		Org:                org,
		Password:           password,
		RetentionPeriodHrs: &retentionPeriodHours,
		Username:           username,
	}
	response, err := c.apiClient.PostSetupWithResponse(ctx, params, *body)
	if err != nil {
		return nil, err
	}
	if response.JSONDefault != nil {
		return nil, domain.DomainErrorToError(response.JSONDefault, response.StatusCode())
	}
	c.httpService.SetAuthorization("Token " + *response.JSON201.Auth.Token)
	return response.JSON201, nil
}

func (c *clientImpl) Health(ctx context.Context) (*domain.HealthCheck, error) {
	params := &domain.GetHealthParams{}
	response, err := c.apiClient.GetHealthWithResponse(ctx, params)
	if err != nil {
		return nil, err
	}
	if response.JSONDefault != nil {
		return nil, domain.DomainErrorToError(response.JSONDefault, response.StatusCode())
	}
	if response.JSON503 != nil {
		//unhealthy server
		return response.JSON503, nil
	}

	return response.JSON200, nil
}

func (c *clientImpl) WriteApi(org, bucket string) api.WriteApi {
	w := api.NewWriteApiImpl(org, bucket, c.httpService, c.options.writeOptions)
	c.writeApis = append(c.writeApis, w)
	return w
}

func (c *clientImpl) WriteApiBlocking(org, bucket string) api.WriteApiBlocking {
	w := api.NewWriteApiBlockingImpl(org, bucket, c.httpService, c.options.writeOptions)
	return w
}

func (c *clientImpl) Close() {
	for _, w := range c.writeApis {
		w.Close()
	}
}

func (c *clientImpl) QueryApi(org string) api.QueryApi {
	return api.NewQueryApi(org, c.httpService)
}

func (c *clientImpl) AuthorizationsApi() api.AuthorizationsApi {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.authApi == nil {
		c.authApi = api.NewAuthorizationApi(c.apiClient)
	}
	return c.authApi
}

func (c *clientImpl) OrganizationsApi() api.OrganizationsApi {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.orgApi == nil {
		c.orgApi = api.NewOrganizationsApi(c.apiClient)
	}
	return c.orgApi
}

func (c *clientImpl) UsersApi() api.UsersApi {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.usersApi == nil {
		c.usersApi = api.NewUsersApi(c.apiClient)
	}
	return c.usersApi
}

func (c *clientImpl) DeleteApi() api.DeleteApi {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.deleteApi == nil {
		c.deleteApi = api.NewDeleteApi(c.apiClient)
	}
	return c.deleteApi
}

func (c *clientImpl) BucketsApi() api.BucketsApi {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.bucketsApi == nil {
		c.bucketsApi = api.NewBucketsApi(c.apiClient)
	}
	return c.bucketsApi
}

func (c *clientImpl) LabelsApi() api.LabelsApi {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.labelsApi == nil {
		c.labelsApi = api.NewLabelsApi(c.apiClient)
	}
	return c.labelsApi
}
