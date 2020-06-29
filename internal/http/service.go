// Copyright 2020 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package http

import (
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"mime"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"

	http2 "github.com/influxdata/influxdb-client-go/api/http"
)

// Http operation callbacks
type RequestCallback func(req *http.Request)
type ResponseCallback func(resp *http.Response) error

// Service handles HTTP operations with taking care of mandatory request headers
type Service interface {
	PostRequest(ctx context.Context, url string, body io.Reader, requestCallback RequestCallback, responseCallback ResponseCallback) *Error
	GetRequest(ctx context.Context, url string, requestCallback RequestCallback, responseCallback ResponseCallback) *Error
	DoHttpRequest(req *http.Request, requestCallback RequestCallback, responseCallback ResponseCallback) *Error
	DoHttpRequestWithResponse(req *http.Request, requestCallback RequestCallback) (*http.Response, error)
	SetAuthorization(authorization string)
	Authorization() string
	HttpClient() *http.Client
	ServerApiUrl() string
	ServerUrl() string
}

// serviceImpl implements Service interface
type serviceImpl struct {
	serverApiUrl  string
	serverUrl     string
	authorization string
	client        *http.Client
}

// NewService creates instance of http Service with given parameters
func NewService(serverUrl, authorization string, httpOptions *http2.Options) Service {
	apiUrl, err := url.Parse(serverUrl)
	serverApiUrl := serverUrl
	if err == nil {
		apiUrl, err = apiUrl.Parse("/api/v2/")
		if err == nil {
			serverApiUrl = apiUrl.String()
		}
	}
	return &serviceImpl{
		serverApiUrl:  serverApiUrl,
		serverUrl:     serverUrl,
		authorization: authorization,
		client: &http.Client{
			Timeout: time.Second * time.Duration(httpOptions.HttpRequestTimeout()),
			Transport: &http.Transport{
				DialContext: (&net.Dialer{
					Timeout: 5 * time.Second,
				}).DialContext,
				TLSHandshakeTimeout: 5 * time.Second,
				TLSClientConfig:     httpOptions.TlsConfig(),
			},
		},
	}
}

func (s *serviceImpl) ServerApiUrl() string {
	return s.serverApiUrl
}

func (s *serviceImpl) ServerUrl() string {
	return s.serverUrl
}

func (s *serviceImpl) SetAuthorization(authorization string) {
	s.authorization = authorization
}

func (s *serviceImpl) Authorization() string {
	return s.authorization
}

func (s *serviceImpl) HttpClient() *http.Client {
	return s.client
}

func (s *serviceImpl) PostRequest(ctx context.Context, url string, body io.Reader, requestCallback RequestCallback, responseCallback ResponseCallback) *Error {
	return s.doHttpRequestWithUrl(ctx, http.MethodPost, url, body, requestCallback, responseCallback)
}

func (s *serviceImpl) doHttpRequestWithUrl(ctx context.Context, method, url string, body io.Reader, requestCallback RequestCallback, responseCallback ResponseCallback) *Error {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return NewError(err)
	}
	return s.DoHttpRequest(req, requestCallback, responseCallback)
}

func (s *serviceImpl) DoHttpRequest(req *http.Request, requestCallback RequestCallback, responseCallback ResponseCallback) *Error {
	resp, err := s.DoHttpRequestWithResponse(req, requestCallback)
	if err != nil {
		return NewError(err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return s.handleHttpError(resp)
	}
	if responseCallback != nil {
		err := responseCallback(resp)
		if err != nil {
			return NewError(err)
		}
	}
	return nil
}

func (s *serviceImpl) DoHttpRequestWithResponse(req *http.Request, requestCallback RequestCallback) (*http.Response, error) {
	req.Header.Set("Authorization", s.authorization)
	req.Header.Set("User-Agent", UserAgent)
	if requestCallback != nil {
		requestCallback(req)
	}
	return s.client.Do(req)
}

func (s *serviceImpl) GetRequest(ctx context.Context, url string, requestCallback RequestCallback, responseCallback ResponseCallback) *Error {
	return s.doHttpRequestWithUrl(ctx, http.MethodGet, url, nil, requestCallback, responseCallback)
}

func (s *serviceImpl) handleHttpError(r *http.Response) *Error {
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
