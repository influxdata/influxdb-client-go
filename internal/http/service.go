// Copyright 2020 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

// Package http provides http related servicing  stuff
package http

import (
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"net/url"
	"strconv"

	http2 "github.com/influxdata/influxdb-client-go/api/http"
)

// RequestCallback defines function called after a request is created before any call
type RequestCallback func(req *http.Request)

// ResponseCallback defines function called after a successful response was received
type ResponseCallback func(resp *http.Response) error

// Service handles HTTP operations with taking care of mandatory request headers
type Service interface {
	PostRequest(ctx context.Context, url string, body io.Reader, requestCallback RequestCallback, responseCallback ResponseCallback) *Error
	GetRequest(ctx context.Context, url string, requestCallback RequestCallback, responseCallback ResponseCallback) *Error
	DoHTTPRequest(req *http.Request, requestCallback RequestCallback, responseCallback ResponseCallback) *Error
	DoHTTPRequestWithResponse(req *http.Request, requestCallback RequestCallback) (*http.Response, error)
	SetAuthorization(authorization string)
	Authorization() string
	HTTPClient() *http.Client
	ServerAPIURL() string
	ServerURL() string
}

// service implements Service interface
type service struct {
	serverAPIURL  string
	serverURL     string
	authorization string
	client        *http.Client
}

// NewService creates instance of http Service with given parameters
func NewService(serverURL, authorization string, httpOptions *http2.Options) Service {
	apiURL, err := url.Parse(serverURL)
	serverAPIURL := serverURL
	if err == nil {
		apiURL, err = apiURL.Parse("api/v2/")
		if err == nil {
			serverAPIURL = apiURL.String()
		}
	}
	return &service{
		serverAPIURL:  serverAPIURL,
		serverURL:     serverURL,
		authorization: authorization,
		client: httpOptions.HTTPClient(),
	}
}

func (s *service) ServerAPIURL() string {
	return s.serverAPIURL
}

func (s *service) ServerURL() string {
	return s.serverURL
}

func (s *service) SetAuthorization(authorization string) {
	s.authorization = authorization
}

func (s *service) Authorization() string {
	return s.authorization
}

func (s *service) HTTPClient() *http.Client {
	return s.client
}

func (s *service) PostRequest(ctx context.Context, url string, body io.Reader, requestCallback RequestCallback, responseCallback ResponseCallback) *Error {
	return s.doHTTPRequestWithURL(ctx, http.MethodPost, url, body, requestCallback, responseCallback)
}

func (s *service) doHTTPRequestWithURL(ctx context.Context, method, url string, body io.Reader, requestCallback RequestCallback, responseCallback ResponseCallback) *Error {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return NewError(err)
	}
	return s.DoHTTPRequest(req, requestCallback, responseCallback)
}

func (s *service) DoHTTPRequest(req *http.Request, requestCallback RequestCallback, responseCallback ResponseCallback) *Error {
	resp, err := s.DoHTTPRequestWithResponse(req, requestCallback)
	if err != nil {
		return NewError(err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return s.handleHTTPError(resp)
	}
	if responseCallback != nil {
		err := responseCallback(resp)
		if err != nil {
			return NewError(err)
		}
	}
	return nil
}

func (s *service) DoHTTPRequestWithResponse(req *http.Request, requestCallback RequestCallback) (*http.Response, error) {
	req.Header.Set("Authorization", s.authorization)
	req.Header.Set("User-Agent", UserAgent)
	if requestCallback != nil {
		requestCallback(req)
	}
	return s.client.Do(req)
}

func (s *service) GetRequest(ctx context.Context, url string, requestCallback RequestCallback, responseCallback ResponseCallback) *Error {
	return s.doHTTPRequestWithURL(ctx, http.MethodGet, url, nil, requestCallback, responseCallback)
}

func (s *service) handleHTTPError(r *http.Response) *Error {
	// successful status code range
	if r.StatusCode >= 200 && r.StatusCode < 300 {
		return nil
	}
	defer func() {
		// discard body so connection can be reused
		_, _ = io.Copy(ioutil.Discard, r.Body)
		_ = r.Body.Close()
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
		perror.Err = json.NewDecoder(r.Body).Decode(perror)
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
		default:
			perror.Code = r.Status
			perror.Message = r.Header.Get("X-Influxdb-Error")
		}
	}

	return perror
}
