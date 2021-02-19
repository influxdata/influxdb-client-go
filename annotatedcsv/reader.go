// Copyright 2021 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

// Package annotatedcsv provides a reader for annotated CSV sources.
// Annotated CSV must contain at least an annotation with data types definition.
// The first row after annotations must be a line with column names.
//
// Annotated CSV example:
//		#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
//		#group,false,false,true,true,false,false,true,true,true,true
//		#default,_result,,,,,,,,,
//		,result,table,_start,_stop,_time,_value,_field,_measurement,location,sensor
//		,,0,2020-02-17T22:19:49.747562847Z,2020-02-18T22:19:49.747562847Z,2020-02-18T10:34:08.135814545Z,23.4,tmp,air,livingroom,SHT31
//		,,0,2020-02-17T22:19:49.747562847Z,2020-02-18T22:19:49.747562847Z,2020-02-18T22:08:44.850214724Z,21.6,tmp,air,livingroom,SHT31
//
// Each set of rows started with annotation is called "section".
// CSV source can contain multiple sections, each section can have different columns (schema).
package annotatedcsv

import (
	"encoding/base64"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/influxdb-client-go/internal/csv"
)

type Column struct {
	Name    string
	Group   bool
	Default interface{}
	Type    string
}

// NewReader returns new Reader for parsing annotated csv stream.
func NewReader(r io.ReadCloser) *Reader {
	r1 := &Reader{
		r:      csv.NewReader(r),
		closer: r,
	}
	r1.r.FieldsPerRecord = -1
	return r1
}

// Reader parses annotated csv stream with single or multiple sections.
// Walking though the csv stream is done by repeatedly calling NextSection() and NextRow() until return false.
// NextRow() can be also called initially, to advance straight to the first row of the first table.
// Actual table schema (columns with names, data types, etc) is returned by Columns() method.
// Data are acquired by Row() or by ValueByName() functions.
// Preliminary end can be caused by an error, so when NextSection() or NextRow() return false, check Err() for an error.
// Reader is automatically closed at the end reading or in case of an error.
// Close() must be called manually in case of not reading stream till the end.
type Reader struct {
	cols []Column
	row  []interface{}
	err  error

	hasPeeked bool
	peekRow   []string
	peekErr   error
	closer    io.Closer
	r         *csv.Reader
	// columnIndexes maps column names to their indexes.
	columnIndexes map[string]int
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
	if r.cols != nil {
		for r.NextRow() {
		}
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
// before NextSection has been called, or after NextSection returns false).
func (r *Reader) Columns() []Column {
	return r.cols
}

// NextRow advances to the next row in the current section.
// If called in the beginning it also advances to the first section.
// When there are no more rows in the current section, it returns false.
func (r *Reader) NextRow() bool {
	if r.err != nil {
		return false
	}
	// Support for navigating straight to the first data row
	if r.cols == nil && !r.NextSection() {
		return false
	}
	row, err := r.readRow()
	r.row = row
	if row == nil {
		r.err = err
		r.cols = nil
		return false
	}
	return true
}

// Row returns the values in the current row of the current section.
// It returns nil if there is no current row.
// All rows in a section have the same number of values.
func (r *Reader) Row() []interface{} {
	return r.row
}

// ValueByName returns value for given column name.
// It returns nil if section has no value for such column.
func (r *Reader) ValueByName(name string) interface{} {
	if r.columnIndexes != nil {
		if i, ok := r.columnIndexes[name]; ok {
			return r.row[i]
		}
	}
	return nil
}

// Close closes underlying reader.
func (r *Reader) Close() error {
	r.cols = nil
	r.row = nil
	r.columnIndexes = nil
	return r.closer.Close()
}

func (r *Reader) safeClose() {
	if err := r.Close(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error closing csv reader: %v\n", err)
	}
}

func (r *Reader) readRow() ([]interface{}, error) {
	closer := func() {
		r.safeClose()
	}
	defer func() { closer() }()
	row, err := r.peek()
	if err != nil {
		return nil, err
	}
	if len(row) > 0 && strings.HasPrefix(row[0], "#") {
		// Start of next table.
		return nil, nil
	}
	_, _ = r.read()
	if len(row)-1 != len(r.cols) {
		return nil, fmt.Errorf("inconsistent number of columns at line %d (got %d items want %d)", r.r.Line(), len(row)-1, len(r.cols))
	}
	rowVals := make([]interface{}, len(row)-1)
	for i, val := range row[1:] {
		col := r.cols[i]
		if col.Default != nil && val == "" {
			rowVals[i] = col.Default
			continue
		}
		if val == "" && col.Name == "" {
			continue
		}
		x, err := convertToType(val, col.Type)
		if err != nil {
			return nil, fmt.Errorf("cannot convert value %q to type %q at line %d: %v", val, col.Type, r.r.Line(), err)
		}
		rowVals[i] = x
	}
	closer = func() {}
	return rowVals, nil
}

func (r *Reader) readHeader() ([]Column, error) {
	closer := func() {
		r.safeClose()
	}
	defer func() { closer() }()
	var cols []Column
	var defaults []string
	for {
		row, err := r.peek()
		if err != nil {
			return cols, err
		}
		_, _ = r.read()
		if cols == nil {
			cols = make([]Column, len(row)-1)
		} else if len(row)-1 != len(cols) {
			return nil, fmt.Errorf("inconsistent table header (got %d items want %d)", len(row)-1, len(cols))
		}
		r.columnIndexes = map[string]int{}
		if !strings.HasPrefix(row[0], "#") {
			if row[1] == "error" {
				// next row is error definition
				row, err = r.read()
				if err != nil {
					return cols, err
				}
				var message string
				if len(row) > 1 && len(row[1]) > 0 {
					message = row[1]
				} else {
					message = "unknown query error"
				}
				reference := ""
				if len(row) > 2 && len(row[2]) > 0 {
					reference = fmt.Sprintf(",%s", row[2])
				}
				return nil, fmt.Errorf("%s%s", message, reference)
			}
			if cols[0].Type == "" {
				return nil, fmt.Errorf("datatype annotation not found")
			}
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
			defaults = row[1:]
		default:
			_, _ = fmt.Fprintf(os.Stderr, "unknown column annotation %q\n", row[0])
		}
	}
	for i, d := range defaults {
		if d == "" {
			continue
		}
		x, err := convertToType(d, cols[i].Type)
		if err != nil {
			return nil, fmt.Errorf("cannot convert default value %q to type %q: %v", d, cols[i].Type, err)
		}
		cols[i].Default = x
	}
	closer = func() {}
	return cols, nil
}

func convertToType(s string, typ string) (interface{}, error) {
	switch typ {
	case "boolean":
		return strconv.ParseBool(s)
	case "long":
		return strconv.ParseInt(s, 10, 64)
	case "unsignedLong":
		return strconv.ParseUint(s, 10, 64)
	case "double":
		x, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return nil, err
		}
		if math.IsInf(x, 0) || math.IsNaN(x) {
			return s, nil
		}
		return x, nil
	case "string", "tag", "":
		return s, nil
	case "duration":
		return time.ParseDuration(s)
	case "base64Binary":
		return base64.StdEncoding.DecodeString(s)
	}
	if datetimeFormat := strings.TrimPrefix(typ, "dateTime"); len(datetimeFormat) != len(typ) {
		layout := timeFormats["RFC3339"]
		if strings.HasPrefix(datetimeFormat, ":") {
			layout = timeFormats[datetimeFormat[1:]]
		}
		if layout == "" {
			return nil, fmt.Errorf("unknown time format %q", typ)
		}
		return time.Parse(layout, s)
	}
	_, _ = fmt.Fprintf(os.Stderr, "unknown datatype %q\n", typ)
	return s, nil
}

var timeFormats = map[string]string{
	"RFC3339":     time.RFC3339,
	"RFC3339Nano": time.RFC3339Nano,
}

func (r *Reader) read() (_r []string, _err error) {
	if r.hasPeeked {
		row, err := r.peekRow, r.peekErr
		r.peekRow, r.peekErr, r.hasPeeked = nil, nil, false
		return row, err
	}
	return r.r.Read()
}

func (r *Reader) peek() (_r []string, _err error) {
	if r.hasPeeked {
		return r.peekRow, r.peekErr
	}
	row, err := r.r.Read()
	r.peekRow, r.peekErr, r.hasPeeked = row, err, true
	return row, err
}
