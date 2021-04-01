// Copyright 2021 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

// Package annotatedcsv provides a reader for annotated CSV sources.
// Annotated CSV must contain at least an annotation with data types definition.
// The first row after annotations must be a line with column names.
//
// Annotated CSV with multiple sections example:
//		#datatype,string,long,dateTime:RFC3339,double,string,string
//		#group,false,false,false,false,true,true
//		#default,_result,,,,,
//		,result,table,_time,_value,_field,location
//		,,0,2020-02-18T10:34:08.135814545Z,23.4,temp,livingroom
//		,,0,2020-02-18T22:08:44.850214724Z,21.6,temp,livingroom
//
//		#datatype,string,long,dateTime:RFC3339,double,string,string,string
//		#group,false,false,false,false,true,true,true
//		#default,_result,,,,,,
//		,result,table,_time,_value,_field,location,sensor
//		,,0,2020-02-18T10:34:08.135814545Z,20.5,temp,bedroom,SHT31
//		,,0,2020-02-18T22:08:44.850214724Z,19.7,temp,bedroom,SHT31
//
// The set of rows following the annotation header is referred to in this documentation
// as a "section". A single CSV stream can contain multiple sections, each with its own
// column metadata.
// CSV source can contain multiple sections, each section can have different columns (schema).
//
// Read more info about annotated CSV at https://docs.influxdata.com/influxdb/v2.0/reference/syntax/annotated-csv/.
package annotatedcsv

import (
	"fmt"
	"github.com/influxdata/influxdb-client-go/internal/csv"
	"io"
	"reflect"
	"strings"
)

// Column represents a column in one section of CSV.
type Column struct {
	// Name holds the name of the column.
	Name string
	// Group specifies whether this column is a group key.
	Group bool
	// Default holds the default value for items in this column.
	Default string
	// Type holds the type of the column.
	Type string
}

// NewReader returns new Reader for parsing a stream in annotated CSV format.
func NewReader(r io.Reader) *Reader {
	r1 := &Reader{
		r: csv.NewReader(r),
	}
	r1.r.FieldsPerRecord = -1
	return r1
}

// Reader parses an annotated CSV stream.
// Successive calls to NextSection will start reading each section of the CSV in turn.
// The Columns method returns information on the columns (schema) for the section.
//
// Within a section, successive calls to NextRow will advance to each row.
// If NextRow is called without first calling NextSection, the rows from the first section will be returned.
//
// The Row method returns all the data for a given row in raw (string) format.
// Use Decode to convert values to fields of a struct or slice.
type Reader struct {
	cols []Column
	row  []string
	err  error

	hasPeeked bool
	peekRow   []string
	peekErr   error
	r         *csv.Reader
	// columnIndexes maps column names to their indexes.
	columnIndexes map[string]int
	// decodeType holds the type that was last decoded into.
	// This is reset at the start of each section.
	decodeType reflect.Type
	// colSetters holds an element for each column in the current
	// section. Given an instance of decodeType, it decodes
	// the column string value into the appropriate field.
	colSetters []fieldSetter
	// decodeRow is the function for decoding a CSV row
	// into an instance of decodeType using colSetters
	decodeRow rowDecoder
}

// NextSection advances to the next section in the csv stream and reports whether
// there is one.
// Any remaining data in the current section is discarded.
// When there are no more sections, it returns false.
func (r *Reader) NextSection() bool {
	if r.err != nil {
		return false
	}
	// Read all the current rows.
	for r.NextRow() {
	}
	_, err := r.peek()
	if err != nil {
		r.err = err
		return false
	}
	cols, err := r.readHeader()
	if err != nil {
		r.err = err
		return false
	}
	r.cols = cols
	// Zero the cached decode type because it probably won't be
	// valid for the next section.
	r.decodeType = nil
	return true
}

// Err returns any error encountered when parsing.
func (r *Reader) Err() error {
	if r.err == io.EOF {
		return nil
	}
	return r.err
}

// Columns returns information on the columns in the current
// section. It returns nil if there is no current section (for example
// before NextSection or NextRow has been called, or after NextSection returns false).
func (r *Reader) Columns() []Column {
	return r.cols
}

// NextRow advances to the next row in the current section.
// When there are no more rows in the current section, it returns false.
func (r *Reader) NextRow() bool {
	if r.cols == nil || r.err != nil {
		return false
	}
	row, err := r.readRow()
	r.row = row
	if row == nil {
		r.err = err
		r.cols = nil
		r.columnIndexes = nil
		return false
	}
	return true
}

// Row returns the raw values in the current row of the current section.
// It returns nil if there is no current row.
// All rows in a section have the same number of values.
func (r *Reader) Row() []string {
	return r.row
}

func (r *Reader) readRow() ([]string, error) {
	row, err := r.peek()
	if err != nil {
		return nil, err
	}
	if len(row) > 0 && strings.HasPrefix(row[0], "#") {
		// Start of next table.
		return nil, nil
	}
	_, _ = r.read()
	colsCount := len(row) - 1
	if colsCount != len(r.cols) {
		return nil, fmt.Errorf("inconsistent number of columns at line %d (got %d items want %d)", r.r.Line(), colsCount, len(r.cols))
	}
	for i, v := range row[1:] {
		if v == "" {
			row[i+1] = r.cols[i].Default
		}
	}
	return row[1:], nil
}

func (r *Reader) readHeader() ([]Column, error) {
	var cols []Column
	r.columnIndexes = map[string]int{}
	for {
		row, err := r.read()
		if err != nil {
			return cols, err
		}
		colsCount := len(row) - 1
		if cols == nil {
			cols = make([]Column, colsCount)
		} else if colsCount != len(cols) {
			return nil, fmt.Errorf("inconsistent table header (got %d items want %d)", colsCount, len(cols))
		}
		if !strings.HasPrefix(row[0], "#") {
			for i, col := range row[1:] {
				cols[i].Name = col
				r.columnIndexes[col] = i
			}
			break
		}
		switch row[0] {
		case "#datatype":
			for i, t := range row[1:] {
				cols[i].Type = t
			}
		case "#group":
			for i, c := range row[1:] {
				cols[i].Group = c == "true"
			}
		case "#default":
			for i, c := range row[1:] {
				cols[i].Default = c
			}
		}
	}
	return cols, nil
}

// read consumes the next line from the CSV,
// discarding any peeked data.
func (r *Reader) read() ([]string, error) {
	if r.hasPeeked {
		row, err := r.peekRow, r.peekErr
		r.peekRow, r.peekErr, r.hasPeeked = nil, nil, false
		return row, err
	}
	return r.r.Read()
}

// peek returns the next line without consuming it.
// Successive calls to peek will return the same line.
func (r *Reader) peek() ([]string, error) {
	if r.hasPeeked {
		return r.peekRow, r.peekErr
	}
	row, err := r.r.Read()
	r.peekRow, r.peekErr, r.hasPeeked = row, err, true
	return row, err
}
