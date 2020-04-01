// Copyright 2020 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package influxdb2

import "fmt"

// Error represent error response from InfluxDBServer or http error
type Error struct {
	StatusCode int
	Code       string
	Message    string
	Err        error
	RetryAfter uint
}

// Error fulfils error interface
func (e *Error) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// NewError returns newly created Error initialised with nested error and default values
func NewError(err error) *Error {
	return &Error{
		StatusCode: 0,
		Code:       "",
		Message:    "",
		Err:        err,
		RetryAfter: 0,
	}
}
