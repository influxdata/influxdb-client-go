// Copyright 2020 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

// Package influxdb2 provides API for using InfluxDB client in Go.
// It's intended to use with InfluxDB 2 server. WriteApi, QueryApi and Health work also with InfluxDB 1.8
package influxdb2

import (
	"context"
	"errors"
	"strings"
	"sync"

	"github.com/influxdata/influxdb-client-go/api"
	"github.com/influxdata/influxdb-client-go/domain"
	ihttp "github.com/influxdata/influxdb-client-go/internal/http"
	"github.com/influxdata/influxdb-client-go/internal/log"
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
	// ServerURL returns the url of the server url client talks to
	ServerURL() string
	// ServerURL returns the url of the server url client talks to
	// Deprecated: Use ServerURL instead.
	ServerUrl() string
	// WriteAPI returns the asynchronous, non-blocking, Write client
	WriteAPI(org, bucket string) api.WriteAPI
	// WriteApi returns the asynchronous, non-blocking, Write client
	// Deprecated: Use WriteAPI instead
	WriteApi(org, bucket string) api.WriteApi
	// WriteAPIBlocking returns the synchronous, blocking, Write client
	WriteAPIBlocking(org, bucket string) api.WriteAPIBlocking
	// WriteApi returns the synchronous, blocking, Write client.
	// Deprecated: Use WriteAPIBlocking instead.
	WriteApiBlocking(org, bucket string) api.WriteApiBlocking
	// QueryAPI returns Query client
	QueryAPI(org string) api.QueryAPI
	// QueryApi returns Query client
	// Deprecated: Use QueryAPI instead.
	QueryApi(org string) api.QueryApi
	// AuthorizationsAPI returns Authorizations API client.
	AuthorizationsAPI() api.AuthorizationsAPI
	// AuthorizationsApi returns Authorizations API client.
	// Deprecated: Use AuthorizationsAPI instead.
	AuthorizationsApi() api.AuthorizationsApi
	// OrganizationsAPI returns Organizations API client
	OrganizationsAPI() api.OrganizationsAPI
	// OrganizationsApi returns Organizations API client.
	// Deprecated: Use OrganizationsAPI instead.
	OrganizationsApi() api.OrganizationsApi
	// UsersAPI returns Users API client.
	UsersAPI() api.UsersAPI
	// UsersApi returns Users API client.
	// Deprecated: Use UsersAPI instead.
	UsersApi() api.UsersApi
	// DeleteAPI returns Delete API client
	DeleteAPI() api.DeleteAPI
	// DeleteApi returns Delete API client.
	// Deprecated: Use DeleteAPI instead.
	DeleteApi() api.DeleteApi
	// BucketsAPI returns Buckets API client
	BucketsAPI() api.BucketsAPI
	// BucketsApi returns Buckets API client.
	// Deprecated: Use BucketsAPI instead.
	BucketsApi() api.BucketsApi
	// LabelsAPI returns Labels API client
	LabelsAPI() api.LabelsAPI
	// LabelsApi returns Labels API client;
	// Deprecated: Use LabelsAPI instead.
	LabelsApi() api.LabelsApi
}

// clientImpl implements Client interface
type clientImpl struct {
	serverURL   string
	options     *Options
	writeAPIs   []api.WriteAPI
	lock        sync.Mutex
	httpService ihttp.Service
	apiClient   *domain.ClientWithResponses
	authAPI     api.AuthorizationsAPI
	//lint:ignore ST1003 Field for deprecated API to be removed in the next release
	authApi api.AuthorizationsApi
	orgAPI  api.OrganizationsAPI
	//lint:ignore ST1003 Field for deprecated API to be removed in the next release
	orgApi api.OrganizationsApi
	//lint:ignore ST1003 Field for deprecated API to be removed in the next release
	usersApi  api.UsersApi
	usersAPI  api.UsersAPI
	deleteAPI api.DeleteAPI
	//lint:ignore ST1003 Field for deprecated API to be removed in the next release
	deleteApi  api.DeleteApi
	bucketsAPI api.BucketsAPI
	//lint:ignore ST1003 Field for deprecated API to be removed in the next release
	bucketsApi api.BucketsApi
	//lint:ignore ST1003 Field for deprecated API to be removed in the next release
	labelsApi api.LabelsApi
	labelsAPI api.LabelsAPI
}

// NewClient creates Client for connecting to given serverURL with provided authentication token, with the default options.
// Authentication token can be empty in case of connecting to newly installed InfluxDB server, which has not been set up yet.
// In such case Setup will set authentication token
func NewClient(serverURL string, authToken string) Client {
	return NewClientWithOptions(serverURL, authToken, DefaultOptions())
}

// NewClientWithOptions creates Client for connecting to given serverURL with provided authentication token
// and configured with custom Options
// Authentication token can be empty in case of connecting to newly installed InfluxDB server, which has not been set up yet.
// In such case Setup will set authentication token
func NewClientWithOptions(serverURL string, authToken string, options *Options) Client {
	normServerURL := serverURL
	if !strings.HasSuffix(normServerURL, "/") {
		// For subsequent path parts concatenation, url has to end with '/'
		normServerURL = serverURL + "/"
	}
	service := ihttp.NewService(normServerURL, "Token "+authToken, options.httpOptions)
	client := &clientImpl{
		serverURL:   serverURL,
		options:     options,
		writeAPIs:   make([]api.WriteAPI, 0, 5),
		httpService: service,
		apiClient:   domain.NewClientWithResponses(service),
	}
	log.Log.SetDebugLevel(client.Options().LogLevel())
	log.Log.Infof("Using URL '%s', token '%s'", serverURL, authToken)
	return client
}
func (c *clientImpl) Options() *Options {
	return c.options
}

func (c *clientImpl) ServerURL() string {
	return c.serverURL
}

//lint:ignore ST1003 Deprecated method to be removed in the next release
func (c *clientImpl) ServerUrl() string {
	return c.serverURL
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
		Password:           &password,
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

//lint:ignore ST1003 Deprecated method to be removed in the next release
func (c *clientImpl) WriteApi(org, bucket string) api.WriteApi {
	return c.WriteAPI(org, bucket)
}

func (c *clientImpl) WriteAPI(org, bucket string) api.WriteAPI {
	w := api.NewWriteAPI(org, bucket, c.httpService, c.options.writeOptions)
	c.writeAPIs = append(c.writeAPIs, w)
	return w
}

//lint:ignore ST1003 Deprecated method to be removed in the next release
func (c *clientImpl) WriteApiBlocking(org, bucket string) api.WriteApiBlocking {
	return c.WriteAPIBlocking(org, bucket)
}

func (c *clientImpl) WriteAPIBlocking(org, bucket string) api.WriteAPIBlocking {
	w := api.NewWriteAPIBlocking(org, bucket, c.httpService, c.options.writeOptions)
	return w
}

func (c *clientImpl) Close() {
	for _, w := range c.writeAPIs {
		w.Close()
	}
}

func (c *clientImpl) QueryAPI(org string) api.QueryAPI {
	return api.NewQueryAPI(org, c.httpService)
}

//lint:ignore ST1003 Deprecated method to be removed in the next release
func (c *clientImpl) QueryApi(org string) api.QueryApi {
	return c.QueryAPI(org)
}

func (c *clientImpl) AuthorizationsAPI() api.AuthorizationsAPI {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.authAPI == nil {
		c.authAPI = api.NewAuthorizationsAPI(c.apiClient)
	}
	return c.authAPI
}

//lint:ignore ST1003 Deprecated method to be removed in the next release
func (c *clientImpl) AuthorizationsApi() api.AuthorizationsApi {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.authApi == nil {
		c.authApi = api.NewAuthorizationsApi(c.apiClient)
	}
	return c.authApi
}

func (c *clientImpl) OrganizationsAPI() api.OrganizationsAPI {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.orgAPI == nil {
		c.orgAPI = api.NewOrganizationsAPI(c.apiClient)
	}
	return c.orgAPI
}

//lint:ignore ST1003 Deprecated method to be removed in the next release
func (c *clientImpl) OrganizationsApi() api.OrganizationsApi {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.orgApi == nil {
		c.orgApi = api.NewOrganizationsApi(c.apiClient)
	}
	return c.orgApi
}

func (c *clientImpl) UsersAPI() api.UsersAPI {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.usersAPI == nil {
		c.usersAPI = api.NewUsersAPI(c.apiClient)
	}
	return c.usersAPI
}

//lint:ignore ST1003 Deprecated method to be removed in the next release
func (c *clientImpl) UsersApi() api.UsersApi {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.usersApi == nil {
		c.usersApi = api.NewUsersApi(c.apiClient)
	}
	return c.usersApi
}

func (c *clientImpl) DeleteAPI() api.DeleteAPI {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.deleteAPI == nil {
		c.deleteAPI = api.NewDeleteAPI(c.apiClient)
	}
	return c.deleteAPI
}

//lint:ignore ST1003 Deprecated method to be removed in the next release
func (c *clientImpl) DeleteApi() api.DeleteApi {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.deleteApi == nil {
		c.deleteApi = api.NewDeleteApi(c.apiClient)
	}
	return c.deleteApi
}

func (c *clientImpl) BucketsAPI() api.BucketsAPI {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.bucketsAPI == nil {
		c.bucketsAPI = api.NewBucketsAPI(c.apiClient)
	}
	return c.bucketsAPI
}

//lint:ignore ST1003 Deprecated method to be removed in the next release
func (c *clientImpl) BucketsApi() api.BucketsApi {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.bucketsApi == nil {
		c.bucketsApi = api.NewBucketsApi(c.apiClient)
	}
	return c.bucketsApi
}

func (c *clientImpl) LabelsAPI() api.LabelsAPI {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.labelsAPI == nil {
		c.labelsAPI = api.NewLabelsAPI(c.apiClient)
	}
	return c.labelsAPI
}

//lint:ignore ST1003 Deprecated method to be removed in the next release
func (c *clientImpl) LabelsApi() api.LabelsApi {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.labelsApi == nil {
		c.labelsApi = api.NewLabelsApi(c.apiClient)
	}
	return c.labelsApi
}
