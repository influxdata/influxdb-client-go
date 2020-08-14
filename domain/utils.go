// Copyright 2020 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package domain

import (
	"github.com/influxdata/influxdb-client-go/api/http"
)

func DomainErrorToError(error *Error, statusCode int) *http.Error {
	return &http.Error{
		StatusCode: statusCode,
		Code:       string(error.Code),
		Message:    error.Message,
	}
}
