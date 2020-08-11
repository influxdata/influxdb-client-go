// Copyright 2020 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

// Package influxdb2 provides API for using InfluxDB client in Go.
// It's intended to use with InfluxDB 2 server. WriteAPI, QueryAPI and Health work also with InfluxDB 1.8
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
// There two APIs for writing, WriteAPI and WriteAPIBlocking.
// WriteAPI provides asynchronous, non-blocking, methods for writing time series data.
// WriteAPIBlocking provides blocking methods for writing time series data.
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
	// WriteAPIBlocking returns the synchronous, blocking, Write client
	WriteAPIBlocking(org, bucket string) api.WriteAPIBlocking
	// QueryAPI returns Query client
	QueryAPI(org string) api.QueryAPI
	// AuthorizationsAPI returns Authorizations API client.
	AuthorizationsAPI() api.AuthorizationsAPI
	// OrganizationsAPI returns Organizations API client
	OrganizationsAPI() api.OrganizationsAPI
	// UsersAPI returns Users API client.
	UsersAPI() api.UsersAPI
	// DeleteAPI returns Delete API client
	DeleteAPI() api.DeleteAPI
	// BucketsAPI returns Buckets API client
	BucketsAPI() api.BucketsAPI
	// LabelsAPI returns Labels API client
	LabelsAPI() api.LabelsAPI
}

// clientImpl implements Client interface
type clientImpl struct {
	serverURL     string
	options       *Options
	writeAPIs     map[string]api.WriteAPI
	syncWriteAPIs map[string]api.WriteAPIBlocking
	lock          sync.Mutex
	httpService   ihttp.Service
	apiClient     *domain.ClientWithResponses
	authAPI       api.AuthorizationsAPI
	orgAPI        api.OrganizationsAPI
	usersAPI      api.UsersAPI
	deleteAPI     api.DeleteAPI
	bucketsAPI    api.BucketsAPI
	labelsAPI     api.LabelsAPI
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
		serverURL:     serverURL,
		options:       options,
		writeAPIs:     make(map[string]api.WriteAPI, 5),
		syncWriteAPIs: make(map[string]api.WriteAPIBlocking, 5),
		httpService:   service,
		apiClient:     domain.NewClientWithResponses(service),
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

func (c *clientImpl) WriteAPI(org, bucket string) api.WriteAPI {
	c.lock.Lock()
	defer c.lock.Unlock()
	key := org + "_" + bucket
	if _, ok := c.writeAPIs[key]; !ok {
		w := api.NewWriteAPI(org, bucket, c.httpService, c.options.writeOptions)
		c.writeAPIs[key] = w
	}
	return c.writeAPIs[key]
}

func (c *clientImpl) WriteAPIBlocking(org, bucket string) api.WriteAPIBlocking {
	c.lock.Lock()
	defer c.lock.Unlock()
	key := org + "_" + bucket
	if _, ok := c.syncWriteAPIs[key]; !ok {
		w := api.NewWriteAPIBlocking(org, bucket, c.httpService, c.options.writeOptions)
		c.syncWriteAPIs[key] = w
	}
	return c.syncWriteAPIs[key]
}

func (c *clientImpl) Close() {
	for key, w := range c.writeAPIs {
		w.Close()
		delete(c.writeAPIs, key)
	}
	for key := range c.syncWriteAPIs {
		delete(c.syncWriteAPIs, key)
	}
}

func (c *clientImpl) QueryAPI(org string) api.QueryAPI {
	return api.NewQueryAPI(org, c.httpService)
}

func (c *clientImpl) AuthorizationsAPI() api.AuthorizationsAPI {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.authAPI == nil {
		c.authAPI = api.NewAuthorizationsAPI(c.apiClient)
	}
	return c.authAPI
}

func (c *clientImpl) OrganizationsAPI() api.OrganizationsAPI {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.orgAPI == nil {
		c.orgAPI = api.NewOrganizationsAPI(c.apiClient)
	}
	return c.orgAPI
}

func (c *clientImpl) UsersAPI() api.UsersAPI {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.usersAPI == nil {
		c.usersAPI = api.NewUsersAPI(c.apiClient)
	}
	return c.usersAPI
}

func (c *clientImpl) DeleteAPI() api.DeleteAPI {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.deleteAPI == nil {
		c.deleteAPI = api.NewDeleteAPI(c.apiClient)
	}
	return c.deleteAPI
}

func (c *clientImpl) BucketsAPI() api.BucketsAPI {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.bucketsAPI == nil {
		c.bucketsAPI = api.NewBucketsAPI(c.apiClient)
	}
	return c.bucketsAPI
}

func (c *clientImpl) LabelsAPI() api.LabelsAPI {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.labelsAPI == nil {
		c.labelsAPI = api.NewLabelsAPI(c.apiClient)
	}
	return c.labelsAPI
}
