package http

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"io"
	"io/ioutil"
	"mime"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type Service interface {
	PostRequest(ctx context.Context, url string, body io.Reader, requestCallback RequestCallback, responseCallback ResponseCallback) *Error
	GetRequest(ctx context.Context, url string, requestCallback RequestCallback, responseCallback ResponseCallback) *Error
	DoHttpRequest(req *http.Request, requestCallback RequestCallback, responseCallback ResponseCallback) *Error
	SetAuthorization(authorization string)
	Authorization() string
	HttpClient() *http.Client
	ServerApiUrl() string
}

type serviceImpl struct {
	serverApiUrl  string
	authorization string
	client        *http.Client
}

func (s *serviceImpl) ServerApiUrl() string {
	return s.serverApiUrl
}

// Http operation callbacks
type RequestCallback func(req *http.Request)
type ResponseCallback func(resp *http.Response) error

func NewService(serverUrl, authorization string, tlsConfig *tls.Config, httpRequestTimeout uint) Service {
	apiUrl, err := url.Parse(serverUrl)
	if err == nil {
		//apiUrl.Path = path.Join(apiUrl.Path, "/api/v2/")
		apiUrl, err = apiUrl.Parse("/api/v2/")
		if err == nil {
			serverUrl = apiUrl.String()
		}
	}
	return &serviceImpl{
		serverApiUrl:  serverUrl,
		authorization: authorization,
		client: &http.Client{
			Timeout: time.Second * time.Duration(httpRequestTimeout),
			Transport: &http.Transport{
				DialContext: (&net.Dialer{
					Timeout: 5 * time.Second,
				}).DialContext,
				TLSHandshakeTimeout: 5 * time.Second,
				TLSClientConfig:     tlsConfig,
			},
		},
	}
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
	req.Header.Set("Authorization", s.authorization)
	req.Header.Set("User-Agent", UserAgent)
	if requestCallback != nil {
		requestCallback(req)
	}
	resp, err := s.client.Do(req)
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
