// Copyright 2020 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

// Package test provides shared test utils
package test

import (
	"compress/gzip"
	"context"
	"fmt"
	ihttp "github.com/influxdata/influxdb-client-go/internal/http"
	"github.com/stretchr/testify/assert"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"testing"
)

type HTTPService struct {
	serverURL      string
	authorization  string
	lines          []string
	t              *testing.T
	wasGzip        bool
	requestHandler func(c *HTTPService, url string, body io.Reader) error
	replyError     *ihttp.Error
	lock           sync.Mutex
}

func (t *HTTPService) WasGzip() bool {
	return t.wasGzip
}

func (t *HTTPService) SetWasGzip(wasGzip bool) {
	t.wasGzip = wasGzip
}

func (t *HTTPService) ServerURL() string {
	return t.serverURL
}

func (t *HTTPService) ServerAPIURL() string {
	return t.serverURL
}

func (t *HTTPService) Authorization() string {
	return t.authorization
}

func (t *HTTPService) HTTPClient() *http.Client {
	return nil
}

func (t *HTTPService) Close() {
	t.lock.Lock()
	if len(t.lines) > 0 {
		t.lines = t.lines[:0]
	}
	t.wasGzip = false
	t.replyError = nil
	t.requestHandler = nil
	t.lock.Unlock()
}

func (t *HTTPService) SetReplyError(replyError *ihttp.Error) {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.replyError = replyError
}

func (t *HTTPService) ReplyError() *ihttp.Error {
	t.lock.Lock()
	defer t.lock.Unlock()
	return t.replyError
}

func (t *HTTPService) SetAuthorization(_ string) {

}
func (t *HTTPService) GetRequest(_ context.Context, _ string, _ ihttp.RequestCallback, _ ihttp.ResponseCallback) *ihttp.Error {
	return nil
}
func (t *HTTPService) DoHTTPRequest(_ *http.Request, _ ihttp.RequestCallback, _ ihttp.ResponseCallback) *ihttp.Error {
	return nil
}

func (t *HTTPService) DoHTTPRequestWithResponse(_ *http.Request, _ ihttp.RequestCallback) (*http.Response, error) {
	return nil, nil
}

func (t *HTTPService) PostRequest(_ context.Context, url string, body io.Reader, requestCallback ihttp.RequestCallback, _ ihttp.ResponseCallback) *ihttp.Error {
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return ihttp.NewError(err)
	}
	if requestCallback != nil {
		requestCallback(req)
	}
	if req.Header.Get("Content-Encoding") == "gzip" {
		body, _ = gzip.NewReader(body)
		t.wasGzip = true
	}
	assert.Equal(t.t, fmt.Sprintf("%swrite?bucket=my-bucket&org=my-org&precision=ns", t.serverURL), url)

	if t.ReplyError() != nil {
		return t.ReplyError()
	}
	if t.requestHandler != nil {
		err = t.requestHandler(t, url, body)
	} else {
		err = t.decodeLines(body)
	}

	if err != nil {
		return ihttp.NewError(err)
	} else {
		return nil
	}
}

func (t *HTTPService) decodeLines(body io.Reader) error {
	bytes, err := ioutil.ReadAll(body)
	if err != nil {
		return err
	}
	lines := strings.Split(string(bytes), "\n")
	lines = lines[:len(lines)-1]
	t.lock.Lock()
	t.lines = append(t.lines, lines...)
	t.lock.Unlock()
	return nil
}

func (t *HTTPService) Lines() []string {
	t.lock.Lock()
	defer t.lock.Unlock()
	return t.lines
}

func NewTestService(t *testing.T, serverURL string) *HTTPService {
	return &HTTPService{
		t:         t,
		serverURL: serverURL + "/api/v2/",
	}
}
