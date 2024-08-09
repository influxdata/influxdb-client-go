// Copyright 2020-2024 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package http

import (
	"fmt"
	ihttp "net/http"

	"github.com/stretchr/testify/assert"

	"testing"
)

func TestWriteErrorHeaderToString(t *testing.T) {
	header := ihttp.Header{
		"Date":           []string{"2024-08-07T12:00:00.009"},
		"Content-Length": []string{"12"},
		"Content-Type":   []string{"application/json", "encoding UTF-8"},
		"X-Test-Value1":  []string{"SaturnV"},
		"X-Test-Value2":  []string{"Apollo11"},
		"Retry-After":    []string{"2044"},
		"Trace-Id":       []string{"123456789ABCDEF0"},
	}

	err := Error{
		StatusCode: ihttp.StatusBadRequest,
		Code:       "bad request",
		Message:    "this is just a test",
		Err:        nil,
		RetryAfter: 2044,
		Header:     header,
	}

	fullString := err.HeaderToString([]string{})

	// write order is not guaranteed
	assert.Contains(t, fullString, "Date: 2024-08-07T12:00:00.009")
	assert.Contains(t, fullString, "Content-Length: 12")
	assert.Contains(t, fullString, "Content-Type: application/json")
	assert.Contains(t, fullString, "X-Test-Value1: SaturnV")
	assert.Contains(t, fullString, "X-Test-Value2: Apollo11")
	assert.Contains(t, fullString, "Retry-After: 2044")
	assert.Contains(t, fullString, "Trace-Id: 123456789ABCDEF0")

	filterString := err.HeaderToString([]string{"date", "trace-id", "x-test-value1", "x-test-value2"})

	// write order will follow filter arguments
	assert.Equal(t, filterString,
		"Date: 2024-08-07T12:00:00.009\nTrace-Id: 123456789ABCDEF0\nX-Test-Value1: SaturnV\nX-Test-Value2: Apollo11\n",
	)
	assert.NotContains(t, filterString, "Content-Type: application/json")
	assert.NotContains(t, filterString, "Retry-After: 2044")
}

func TestErrorIfaceError(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		err        error
		code       string
		message    string
		expected   string
	}{
		{name: "TestNestedErrorNotNilCode0Message0",
			statusCode: 418,
			err:        fmt.Errorf("original test message"),
			code:       "",
			message:    "",
			expected:   "original test message"},
		{name: "TestNestedErrorNotNilCodeXMessageX",
			statusCode: 418,
			err:        fmt.Errorf("original test message"),
			code:       "bad request",
			message:    "is this a teapot?",
			expected:   "original test message"},
		{name: "TestNestedErrorNilCodeXMessageX",
			statusCode: 418,
			err:        nil,
			code:       "bad request",
			message:    "is this a teapot?",
			expected:   "bad request: is this a teapot?"},
		{name: "TestNesterErrorNilCodeXMessage0",
			statusCode: 418,
			err:        nil,
			code:       "I'm a teapot",
			message:    "",
			expected:   "Unexpected status code 418"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := Error{
				StatusCode: test.statusCode,
				Code:       test.code,
				Message:    test.message,
				Err:        test.err,
				RetryAfter: 0,
				Header:     ihttp.Header{},
			}
			assert.Equal(t, test.expected, err.Error())
		})
	}
}
