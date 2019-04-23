package influxdb

import (
	"errors"
	"fmt"
)

// ErrUnimplemented is an error for when pieces of the client's functionality is unimplemented.
var ErrUnimplemented = errors.New("unimplemented")

type maxRetriesExceededError struct {
	tries int
	err   error
}

func (err maxRetriesExceededError) Error() string {
	return fmt.Sprintf("max retries of %d reached, and we recieved an error of %v", err.tries, err.err)
}

func (err maxRetriesExceededError) Unwrap() error {
	return err.err
}

type genericRespError struct {
	Code      string
	Message   string
	Line      *int32
	MaxLength *int32
}

func (g genericRespError) Error() string {
	errString := g.Code + ": " + g.Message
	if g.Line != nil {
		return fmt.Sprintf("%s - line[%d]", errString, g.Line)
	} else if g.MaxLength != nil {
		return fmt.Sprintf("%s - maxlen[%d]", errString, g.MaxLength)
	}
	return errString
}
