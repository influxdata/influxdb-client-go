// Copyright 2020 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

// package influxdb2 provides API for using InfluxDB client in Go
// It's intended to use with InfluxDB 2 server
package influxdb2

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"net"
	"net/http"
	"net/url"
	"path"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/influxdata/influxdb-client-go/domain"
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
	// Close ensures all ongoing asynchronous write clients finish
	Close()
	// Options returns the options associated with client
	Options() *Options
	// ServerUrl returns the url of the server url client talks to
	ServerUrl() string
	// Setup sends request to initialise new InfluxDB server with user, org and bucket, and data retention period
	// Retention period of zero will result to infinite retention
	// and returns details about newly created entities along with the authorization object
	Setup(ctx context.Context, username, password, org, bucket string, retentionPeriodHours int) (*domain.OnboardingResponse, error)
	// Ready checks InfluxDB server is running
	Ready(ctx context.Context) (bool, error)
	// Internal  method for handling posts
	postRequest(ctx context.Context, url string, body io.Reader, requestCallback RequestCallback, responseCallback ResponseCallback) *Error
}

// client implements InfluxDBClient interface
type client struct {
	serverUrl     string
	authorization string
	options       *Options
	writeApis     []WriteApi
	httpClient    *http.Client
	lock          sync.Mutex
}

// Http operation callbacks
type RequestCallback func(req *http.Request)
type ResponseCallback func(req *http.Response) error

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
		serverUrl:     serverUrl,
		authorization: "Token " + authToken,
		httpClient: &http.Client{
			Timeout: time.Second * 20,
			Transport: &http.Transport{
				DialContext: (&net.Dialer{
					Timeout: 5 * time.Second,
				}).DialContext,
				TLSHandshakeTimeout: 5 * time.Second,
				TLSClientConfig:     options.TlsConfig(),
			},
		},
		options:   options,
		writeApis: make([]WriteApi, 0, 5),
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
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, readyUrl.String(), nil)
	if err != nil {
		return false, err
	}
	req.Header.Set("User-Agent", userAgent())
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK, nil
}

func (c *client) WriteApi(org, bucket string) WriteApi {
	w := newWriteApiImpl(org, bucket, c)
	c.writeApis = append(c.writeApis, w)
	return w
}

func (c *client) WriteApiBlocking(org, bucket string) WriteApiBlocking {
	w := newWriteApiBlockingImpl(org, bucket, c)
	return w
}

func (c *client) Close() {
	for _, w := range c.writeApis {
		w.Close()
	}
}

func (c *client) QueryApi(org string) QueryApi {
	return &queryApiImpl{
		org:    org,
		client: c,
	}
}

func (c *client) postRequest(ctx context.Context, url string, body io.Reader, requestCallback RequestCallback, responseCallback ResponseCallback) *Error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
	if err != nil {
		return NewError(err)
	}
	req.Header.Set("Authorization", c.authorization)
	req.Header.Set("User-Agent", userAgent())
	if requestCallback != nil {
		requestCallback(req)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return NewError(err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return c.handleHttpError(resp)
	}
	if responseCallback != nil {
		err := responseCallback(resp)
		if err != nil {
			return NewError(err)
		}
	}
	return nil
}

func (c *client) handleHttpError(r *http.Response) *Error {
	// successful status code range
	if r.StatusCode >= 200 && r.StatusCode < 300 {
		return nil
	}
	defer func() {
		// discard body so connection can be reused
		_, _ = io.Copy(ioutil.Discard, r.Body)
		r.Body.Close()
	}()

	perror := NewError(nil)
	perror.StatusCode = r.StatusCode
	if v := r.Header.Get("Retry-After"); v != "" {
		r, err := strconv.ParseUint(v, 10, 32)
		if err == nil {
			perror.RetryAfter = uint(r)
		}
	}
	// json encoded error
	ctype, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if ctype == "application/json" {
		err := json.NewDecoder(r.Body).Decode(perror)
		perror.Err = err
		return perror
	} else {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			perror.Err = err
			return perror
		}

		perror.Code = r.Status
		perror.Message = string(body)
	}

	if perror.Code == "" && perror.Message == "" {
		switch r.StatusCode {
		case http.StatusTooManyRequests:
			perror.Code = "too many requests"
			perror.Message = "exceeded rate limit"
		case http.StatusServiceUnavailable:
			perror.Code = "unavailable"
			perror.Message = "service temporarily unavailable"
		}
	}
	return perror
}

// Keeps once created User-Agent string
var userAgentCache string

// userAgent does lazy user-agent string initialisation
func userAgent() string {
	if userAgentCache == "" {
		userAgentCache = fmt.Sprintf("influxdb-client-go/%s  (%s; %s)", Version, runtime.GOOS, runtime.GOARCH)
	}
	return userAgentCache
}
