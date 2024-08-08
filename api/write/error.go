// Copyright 2020-2024 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package write

import (
	"errors"
	"fmt"
	"github.com/influxdata/influxdb-client-go/v2/api/http"
	iHttp "net/http"
	"net/textproto"
)

// Error wraps an error that may have occurred during a write call.  Most often this will be an http.Error.
type Error struct {
	origin  error
	message string
}

// NewError returns a new created Error instance wrapping the original error.
func NewError(origin error, message string) *Error {
	return &Error{
		origin:  origin,
		message: message,
	}
}

// Error fulfills the error interface
func (e *Error) Error() string {
	return fmt.Sprintf("%s:\n   %s", e.message, e.origin)
}

func (e *Error) Unwrap() error {
	return e.origin
}

// HTTPHeader returns the Header of a wrapped http.Error.  If the original error is not http.Error returns standard error.
func (e *Error) HTTPHeader() (iHttp.Header, error) {
	var err *http.Error
	ok := errors.As(e.origin, &err)
	if ok {
		return err.Header, nil
	}
	return nil, fmt.Errorf(fmt.Sprintf("Origin error: (%s) is not of type *http.Error.\n", e.origin.Error()))
}

// GetHeaderValues returns the values from a Header key.  If original error is not http.Error return standard error.
func (e *Error) GetHeaderValues(key string) ([]string, error) {
	var err *http.Error
	if errors.As(e.origin, &err) {
		return err.Header.Values(textproto.CanonicalMIMEHeaderKey(key)), nil
	}
	return nil, fmt.Errorf(fmt.Sprintf("Origin error: (%s) is not of type http.Header.\n", e.origin.Error()))
}

// GetHeader returns the first value from a header key.  If origin is not http.Error or if no match is found returns "".
func (e *Error) GetHeader(key string) string {
	var err *http.Error
	if errors.As(e.origin, &err) {
		return err.Header.Get(textproto.CanonicalMIMEHeaderKey(key))
	}
	return ""
}
