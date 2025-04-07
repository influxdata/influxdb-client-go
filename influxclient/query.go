// Copyright 2021 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package influxclient

import (
	"errors"
	"fmt"
	"io"

	"github.com/influxdata/influxdb-client-go/v3/annotatedcsv"
)

// QueryError defines the information of Flux query error
type QueryError struct {
	// Message is a Flux query error message
	Message string `csv:"error"`
	// Code is an Flux query error code
	Code int64 `csv:"reference"`
}

func (e *QueryError) Error() string {
	return fmt.Sprintf("Flux query error (code %d): %s", e.Code, e.Message)
}

// NewQueryResultReader returns new QueryResultReader for parsing Flux query result stream.
func NewQueryResultReader(r io.ReadCloser) *QueryResultReader {
	return &QueryResultReader{
		Reader:  annotatedcsv.NewReader(r),
		closer:  r,
		initial: true,
	}
}

// QueryResultReader enhances annotatedcsv.Reader
// by allowing NextRow to go straight to the first data line
// and by treating error section as an error.
// QueryResultReader must be closed by calling Close at the end of reading.
type QueryResultReader struct {
	*annotatedcsv.Reader
	err     error
	initial bool
	closer  io.Closer
}

// NextSection is like annotatedcsv.Reader.NextSection
// except that it treats error sections as a terminating section.
// When an error section is encountered, NextSection will
// return false and Err will return the error.
func (r *QueryResultReader) NextSection() bool {
	r.initial = false
	if r.err != nil {
		return false
	}
	if !r.Reader.NextSection() {
		return false
	}
	if err := r.errorSection(); err != nil {
		r.err = err
		return false
	}
	return true
}

// NextRow is like annotatedcsv.Reader.NextRow
// except if it is called in the beginning it advances to the first section.
// When an error section is encountered, NextRow will
// return false and Err will return the error.
func (r *QueryResultReader) NextRow() bool {
	if r.err != nil {
		return false
	}
	if r.initial {
		// Invoke NextSection explicitly so that we'll immediately
		// parse an error result.
		if !r.NextSection() {
			return false
		}
	}
	return r.Reader.NextRow()
}

// errorSection checks if current section annotation defines
// a query error statement and returns error containing the error message.
func (r *QueryResultReader) errorSection() error {
	if r.Columns()[0].Name == "error" {
		cols := r.Columns()
		if len(cols) != 2 || cols[0].Name != "error" || cols[1].Name != "reference" {
			return nil
		}
		// next row is error definition
		if !r.NextRow() {
			if r.err != nil {
				return r.err
			}
			return errors.New("no row found in error section")
		}
		row := r.Row()
		if row[0] == "" {
			return errors.New("no row found in error section")
		}
		var e QueryError
		if err := r.Decode(&e); err != nil {
			return fmt.Errorf("cannot decode row in error section: %v", err)
		}
		return &e
	}
	return nil
}

// Err returns any error encountered when parsing.
func (r *QueryResultReader) Err() error {
	if r.err != nil {
		return r.err
	}
	return r.Reader.Err()
}

// Close closes the underlying reader
func (r *QueryResultReader) Close() error {
	return r.closer.Close()
}
