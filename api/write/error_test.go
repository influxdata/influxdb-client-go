// Copyright 2020-2024 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package write

import (
	"errors"
	"fmt"
	"github.com/influxdata/influxdb-client-go/v2/api/http"
	"github.com/stretchr/testify/assert"
	ihttp "net/http"
	"testing"
)

func TestNewErrorNotHttpError(t *testing.T) {
	err := NewError(fmt.Errorf("origin error"), "error message")
	var errTest *http.Error
	assert.False(t, errors.As(err, &errTest))
	header, okh := err.HTTPHeader()
	assert.Nil(t, header)
	assert.Error(t, okh)
	values, okv := err.GetHeaderValues("Date")
	assert.Nil(t, values)
	assert.Error(t, okv)
	assert.Equal(t, "", err.GetHeader("Date"))
	assert.Equal(t, "error message:\n   origin error", err.Error())
}

func TestNewErrorHttpError(t *testing.T) {
	header := ihttp.Header{
		"Date":           []string{"2024-08-07T12:00:00.009"},
		"Content-Length": []string{"12"},
		"Content-Type":   []string{"application/json", "encoding UTF-8"},
		"X-Test-Value1":  []string{"SaturnV"},
		"X-Test-Value2":  []string{"Apollo11"},
		"Retry-After":    []string{"2044"},
		"Trace-Id":       []string{"123456789ABCDEF0"},
	}

	err := NewError(&http.Error{
		StatusCode: ihttp.StatusBadRequest,
		Code:       "bad request",
		Message:    "this is just a test",
		Err:        nil,
		RetryAfter: 2044,
		Header:     header,
	}, "should be httpError")

	var errTest *http.Error
	assert.True(t, errors.As(err.Unwrap(), &errTest))
	header, okh := err.HTTPHeader()
	assert.NotNil(t, header)
	assert.Nil(t, okh)
	date, okd := err.GetHeaderValues("Date")
	assert.Equal(t, []string{"2024-08-07T12:00:00.009"}, date)
	assert.Nil(t, okd)
	cType, okc := err.GetHeaderValues("Content-Type")
	assert.Equal(t, []string{"application/json", "encoding UTF-8"}, cType)
	assert.Nil(t, okc)
	assert.Equal(t, "2024-08-07T12:00:00.009", err.GetHeader("Date"))
	assert.Equal(t, "SaturnV", err.GetHeader("X-Test-Value1"))
	assert.Equal(t, "Apollo11", err.GetHeader("X-Test-Value2"))
	assert.Equal(t, "123456789ABCDEF0", err.GetHeader("Trace-Id"))
	assert.Equal(t, "should be httpError:\n   bad request: this is just a test", err.Error())
}
