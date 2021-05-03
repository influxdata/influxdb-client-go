// Copyright 2021 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package annotatedcsv

import (
	"encoding/base64"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	ireflect "github.com/influxdata/influxdb-client-go/internal/reflect"
)

// Decode decodes the current row into x, which should be
// a pointer to a struct or a pointer to a slice.
//
// When it's a pointer to a struct, columns in the row are decoded into
// appropriate fields in the struct, using the similar tag conventions
// described by encoding/json to determine how to map
// column names to struct fields. Tag prefix must start with "csv:":
//
//  type Point struct {
//      Timestamp  time.Time `csv:"_time"`
//      Value      float32   `csv:"_value"`
//      Location   string    `csv:"location"`
//      Sensor     string    `csv:"type"`
//  }
//
//  var p Point
//  err := r.Decode(&p)
//
// When it's a pointer to a slice, the slice is changed
// to have one element for each column (reusing
// space in the slice if possible), and each element is
// set to the value in the corresponding column.
//
// When decoding into an empty interface value, the resulting
// type depends on the column type:
//
// - string, tag or unrecognized: string
// - double: float64
// - unsignedLong: uint64
// - long: int64
// - boolean: bool
// - duration: time.Duration
// - dateTime: time.Time
//
// Any value can be decoded into a string without
// error - the result is the value in the CSV, so
//
//     var row []string
//     r.Decode(&row)
//
// will always succeed and provide all the values in the column as strings.
//
// Similarly, any row can be decoded into []interface{}, which will convert
// all values into elements of the slice according to the rules for empty interfaces above.
//
// Slice capacity will be reused when possible.
func (r *Reader) Decode(x interface{}) error {
	t := reflect.TypeOf(x)
	if err := r.initColumns(t, r.cols); err != nil {
		return err
	}
	return r.decodeRow(reflect.ValueOf(x), r.row)
}

// initColumns initializes r.colSetters for decoding the given columns
// into a value of the given type.
func (r *Reader) initColumns(t reflect.Type, columns []Column) error {
	if t == r.decodeType {
		return nil
	}
	if t.Kind() != reflect.Ptr {
		return fmt.Errorf("cannot decode into non-pointer type %v", t)
	}
	et := t.Elem()
	if et.Kind() != reflect.Struct && et.Kind() != reflect.Slice {
		return fmt.Errorf("cannot decode into %v", t)
	}
	setters := make([]fieldSetter, len(columns))
	switch et.Kind() {
	case reflect.Struct:
		fieldsMap := make(map[string]reflect.StructField)
		fields := ireflect.VisibleFields(et)
		for _, f := range fields {
			name := f.Name
			if tag, ok := f.Tag.Lookup("csv"); ok {
				if tag == "-" {
					continue
				}
				if i := strings.Index(tag, ","); i >= 0 {
					if i != len(tag)-1 {
						return fmt.Errorf("tag attributes are not supported")
					}
					tag = tag[:i]
				}
				name = tag
			}
			fieldsMap[name] = f
		}

		for i, col := range columns {
			f, ok := fieldsMap[col.Name]
			if !ok {
				// The column isn't mentioned in the struct.
				continue
			}
			ftype := f.Type
			colType, ok := columnTypes[col.Type]
			if !ok {
				// ignore invalid type and use string
				colType = stringCol
			}
			convert, ok := conversions[conv{colType, fieldKindOf(ftype)}]
			if !ok {
				return fmt.Errorf("cannot convert from column type %s to %v", col.Type, ftype)
			}
			setters[i] = func(v reflect.Value, colIndex int) error {
				return r.convertColumnValue(v.FieldByIndex(f.Index), colIndex, convert)
			}
		}
		r.decodeRow = func(v reflect.Value, row []string) error {
			if v.IsNil() {
				return fmt.Errorf("decode into nil %s", v.Type())
			}
			v = v.Elem()
			for i, c := range r.colSetters {
				if c == nil {
					continue
				}
				if err := c(v, i); err != nil {
					return err
				}
			}
			return nil
		}
	case reflect.Slice:
		s := et.Elem()
		if s != stringType && s != emptyInterfaceType {
			return fmt.Errorf("cannot decode into *[]%v", s)
		}
		fkind := fieldKindOf(s)
		for i, col := range columns {
			colType, ok := columnTypes[col.Type]
			if !ok {
				// ignore invalid type and use string
				colType = stringCol
			}

			convert, ok := conversions[conv{colType, fkind}]
			if !ok {
				return fmt.Errorf("cannot convert from column type %s to %v", col.Type, s)
			}
			setters[i] = func(v reflect.Value, colIndex int) error {
				if err := r.convertColumnValue(v.Index(colIndex), colIndex, convert); err != nil {
					return err
				}
				return nil
			}
		}
		r.decodeRow = func(v reflect.Value, row []string) error {
			if v.IsNil() {
				return fmt.Errorf("decode into nil %s", v.Type())
			}
			e := v.Elem()
			c := len(r.colSetters)
			if e.Cap() < c {
				e = reflect.MakeSlice(e.Type(), c, c)
			} else {
				e = e.Slice(0, c)
			}
			for i, setf := range r.colSetters {
				if err := setf(e, i); err != nil {
					return err
				}
			}
			v.Elem().Set(e)
			return nil
		}
	}

	r.colSetters = setters
	r.decodeType = t

	return nil
}

// convertColumnValue set a value from current row of given column index
// to a struct or a slice field value
func (r *Reader) convertColumnValue(v reflect.Value, colIndex int, convert valueSetter) error {
	if err := convert(v, r.row[colIndex]); err != nil {
		return fmt.Errorf(`cannot convert value %q in column of type %q to Go type %v at line %d: %w`, r.row[colIndex], r.cols[colIndex].Type, v.Type(), r.r.Line(), err)
	}
	return nil
}

// fieldKind represents one of the possible kinds of struct field.
// It's similar to reflect.Kind (every reflect.Kind constant i
// is represented as fieldKind(i)) except that it also has
// defined constants for field types that have special treatment,
// such as time.Duration.
type fieldKind uint

// conv represents a possible conversion source and destination.
type conv struct {
	from colType
	to   fieldKind
}

// intKinds holds all supported signed integer kinds
var intKinds = []fieldKind{
	fieldKind(reflect.Int),
	fieldKind(reflect.Int8),
	fieldKind(reflect.Int16),
	fieldKind(reflect.Int32),
	fieldKind(reflect.Int64),
}

// uintKinds holds all supported unsigned integer kinds
var uintKinds = []fieldKind{
	fieldKind(reflect.Uint),
	fieldKind(reflect.Uint8),
	fieldKind(reflect.Uint16),
	fieldKind(reflect.Uint32),
	fieldKind(reflect.Uint64),
}

// floatKinds holds all supported floating point number kinds
var floatKinds = []fieldKind{
	fieldKind(reflect.Float32),
	fieldKind(reflect.Float64),
}

// rest of supported type kinds
const (
	durationKind = fieldKind(255 + iota)
	timeKind
	bytesKind
)

// colType represents the type of an annotated CSV column.
type colType uint

const (
	stringCol = colType(iota)
	boolCol
	durationCol
	longCol
	unsignedLongCol
	doubleCol
	base64BinaryCol
	rfc3339Col
)

// columnTypes maps annotated CSV types
// to integer column type
var columnTypes = map[string]colType{
	"string":               stringCol,
	"boolean":              boolCol,
	"duration":             durationCol,
	"long":                 longCol,
	"unsignedLong":         unsignedLongCol,
	"double":               doubleCol,
	"base64Binary":         base64BinaryCol,
	"dateTime:RFC3339":     rfc3339Col,
	"dateTime:RFC3339Nano": rfc3339Col,
	"dateTime":             rfc3339Col,
}

// conversions maps all possible conversions from column types to field kinds
var conversions map[conv]valueSetter

// stringType is the exact type for the empty string
var stringType = reflect.TypeOf("")

// emptyInterfaceType is the exact type for the empty interface{}
var emptyInterfaceType = reflect.TypeOf(new(interface{})).Elem()

// canonicalTypes holds the Go type that best represents
// all the annotated CSV column kinds (keyed  by column type).
var canonicalTypes = []reflect.Type{
	stringCol:       reflect.TypeOf(""),
	boolCol:         reflect.TypeOf(true),
	durationCol:     reflect.TypeOf(time.Duration(1)),
	longCol:         reflect.TypeOf(int64(0)),
	unsignedLongCol: reflect.TypeOf(uint64(0)),
	doubleCol:       reflect.TypeOf(0.0),
	base64BinaryCol: reflect.TypeOf([]byte{}),
	rfc3339Col:      reflect.TypeOf(time.Time{}),
}

// fieldKindOf determines a fieldKind for given reflect.Type
func fieldKindOf(t reflect.Type) fieldKind {
	switch t {
	case canonicalTypes[durationCol]:
		return durationKind
	case canonicalTypes[rfc3339Col]:
		return timeKind
	case canonicalTypes[base64BinaryCol]:
		return bytesKind
	}
	if t.Kind() == reflect.Interface && t != emptyInterfaceType {
		// Only the empty interface type is allowed.
		return fieldKind(reflect.Invalid)
	}
	return fieldKind(t.Kind())
}

// fieldSetter sets value of column from colIndex
// to the corresponding field of v
type fieldSetter func(v reflect.Value, colIndex int) error

// valueSetter converts a string value to the value type of v
// and sets it to v
type valueSetter func(v reflect.Value, s string) error

// rowDecoder decodes the row with the given values into v,
// which must have type decodeType.
type rowDecoder func(v reflect.Value, row []string) error

// toInterface returns a function for setting value of given column type
// to an interface{} field
func toInterface(col colType) valueSetter {
	t := canonicalTypes[col]
	convert, ok := conversions[conv{col, fieldKindOf(t)}]
	if !ok {
		panic("conversion not found (should not happen)")
	}
	return func(v reflect.Value, s string) error {
		e := reflect.New(t).Elem()
		if err := convert(e, s); err != nil {
			return err
		}
		v.Set(e)
		return nil
	}
}

// toString sets a string to a value of type string
func toString(v reflect.Value, s string) error {
	v.SetString(s)
	return nil
}

// toFloat converts a string to a value of type float
func toFloat(v reflect.Value, s string) error {
	x, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return err
	}
	if v.OverflowFloat(x) {
		return fmt.Errorf("value %s overflows type %s", s, v.Type())
	}
	v.SetFloat(x)
	return nil
}

// toBool converts a string to a value of type bool
func toBool(v reflect.Value, s string) error {
	var b bool
	switch s {
	case "true":
		b = true
	case "false":
		b = false
	default:
		return fmt.Errorf("invalid bool value: %q", s)
	}
	v.SetBool(b)
	return nil
}

// toInt converts a string to a value of type signed int
func toInt(v reflect.Value, s string) error {
	x, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return err
	}
	if v.OverflowInt(x) {
		return fmt.Errorf("value %s overflows type %s", s, v.Type())
	}
	v.SetInt(x)
	return nil
}

// toUint converts a string to a value of type unsigned int
func toUint(v reflect.Value, s string) error {
	x, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return err
	}
	if v.OverflowUint(x) {
		return fmt.Errorf("value %s overflows type %s", s, v.Type())
	}
	v.SetUint(x)
	return nil
}

// toTime converts a string to a time value
func toTime(v reflect.Value, s string) error {
	x, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		return err
	}
	v.Set(reflect.ValueOf(x))
	return nil
}

// toDuration converts a string to a duration value
func toDuration(v reflect.Value, s string) error {
	x, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	v.Set(reflect.ValueOf(x))
	return nil
}

// toBytes decodes a base64 encoded string to a slice of bytes
func toBytes(v reflect.Value, s string) error {
	x, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return err
	}
	v.Set(reflect.ValueOf(x))
	return nil
}

func init() {
	// initializing conversions in init() because of cycle dependency in toInterface()
	conversions = make(map[conv]valueSetter)
	for _, k := range intKinds {
		conversions[conv{longCol, k}] = toInt
		conversions[conv{unsignedLongCol, k}] = toInt
	}
	for _, k := range uintKinds {
		conversions[conv{unsignedLongCol, k}] = toUint
	}
	for _, k := range floatKinds {
		conversions[conv{longCol, k}] = toFloat
		conversions[conv{unsignedLongCol, k}] = toFloat
		conversions[conv{doubleCol, k}] = toFloat
	}
	conversions[conv{boolCol, fieldKind(reflect.Bool)}] = toBool
	conversions[conv{stringCol, fieldKind(reflect.String)}] = toString
	conversions[conv{durationCol, durationKind}] = toDuration
	conversions[conv{base64BinaryCol, bytesKind}] = toBytes
	conversions[conv{rfc3339Col, timeKind}] = toTime

	for t := range canonicalTypes {
		conversions[conv{colType(t), fieldKind(reflect.Interface)}] = toInterface(colType(t))
		conversions[conv{colType(t), fieldKind(reflect.String)}] = toString
	}
}
