package domain

import ihttp "github.com/influxdata/influxdb-client-go/internal/http"

func DomainErrorToError(error *Error, statusCode int) *ihttp.Error {
	return &ihttp.Error{
		StatusCode: statusCode,
		Code:       string(error.Code),
		Message:    error.Message,
	}
}
