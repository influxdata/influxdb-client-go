package client

import (
	"errors"
	"fmt"
)

// ErrUnimplemented is an error for when pieces of the client's functionality is unimplemented.
var ErrUnimplemented = errors.New("unimplemented")

type apiError struct {
	StatusCode  int
	Title       string
	Description string
}

func (e apiError) Error() string {
	if e.Description != "" {
		return fmt.Sprintf("%s: %s", e.Title, e.Description)
	}
	return e.Title
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
