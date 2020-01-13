package influxdb

import (
	"errors"
	"fmt"
)

// Error code constants copied from influxdb
const (
	EInternal            = "internal error"
	ENotFound            = "not found"
	EConflict            = "conflict"             // action cannot be performed
	EInvalid             = "invalid"              // validation failed
	EUnprocessableEntity = "unprocessable entity" // data type is correct, but out of range
	EEmptyValue          = "empty value"
	EUnavailable         = "unavailable"
	EForbidden           = "forbidden"
	ETooManyRequests     = "too many requests"
	EUnauthorized        = "unauthorized"
	EMethodNotAllowed    = "method not allowed"
	ETooLarge            = "request too large"
)

// ErrUnimplemented is an error for when pieces of the client's functionality is unimplemented.
var ErrUnimplemented = errors.New("unimplemented")

// Error is an error returned by a client operation
// It contains a number of contextual fields which describe the nature
// and cause of the error
type Error struct {
	StatusCode int
	Code       string
	Message    string
	Err        string
	Op         string
	Line       *int32
	MaxLength  *int32
	RetryAfter *int32
}

// Error returns the string representation of the Error struct
func (e *Error) Error() string {
	errString := fmt.Sprintf("%s (%s): %s", e.Code, e.Op, e.Message)
	if e.Line != nil {
		return fmt.Sprintf("%s - line[%d]", errString, e.Line)
	}

	if e.MaxLength != nil {
		return fmt.Sprintf("%s - maxlen[%d]", errString, e.MaxLength)
	}

	return errString
}
