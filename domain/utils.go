// Copyright 2020 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package domain

import ihttp "github.com/influxdata/influxdb-client-go/internal/http"

func DomainErrorToError(error *Error, statusCode int) *ihttp.Error {
	return &ihttp.Error{
		StatusCode: statusCode,
		Code:       string(error.Code),
		Message:    error.Message,
	}
}
