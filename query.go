package influxdb

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/influxdb-client-go/internal/ast"
)

const (
	stringDatatype = "string"
	timeDatatype   = "dateTime"
	floatDatatype  = "double"
	boolDatatype   = "boolean"
	intDatatype    = "long"
	uintDatatype   = "unsignedLong"

	timeDataTypeWithFmt = "dateTime:RFC3339"
)

type queryPost struct {
	Query string `json:"query"`
	//Type    string      `json:"type"`
	Dialect dialect     `json:"dialect"`
	Extern  interface{} `json:"extern,omitempty"`
}

type dialect struct {
	Annotations    []string `json:"annotations,omitempty"`
	CommentPrefix  string   `json:"commentPrefix,omitempty"`
	DateTimeFormat string   `json:"dateTimeFormat,omitempty"`
	Delimiter      string   `json:"delimiter,omitempty"`
	Header         bool     `json:"header"`
}

// QueryCSV returns the result of a flux query.
// TODO: annotations
func (c *Client) QueryCSV(ctx context.Context, flux string, org string, extern ...interface{}) (*QueryCSVResult, error) {
	qURL, err := c.makeQueryURL(org)
	if err != nil {
		return nil, err
	}
	qp := queryPost{Query: flux /*Type:"flux",*/, Dialect: dialect{
		Annotations: []string{"datatype", "group", "default"},
		Delimiter:   ",",
		Header:      true,
	}}
	if len(extern) > 0 {
		qp.Extern, err = ast.FluxExtern(extern...)
		if err != nil {
			return nil, err
		}
	}
	data, err := json.Marshal(qp)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", qURL, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	req = req.WithContext(ctx)
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Authorization", c.authorization)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept-Encoding", "gzip")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	// this is so we can unset the defer later if we don't error.
	cleanup := func() {
		resp.Body.Close()
	}
	defer func() { cleanup() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		r := io.LimitReader(resp.Body, 1<<14) // only support errors that are 16kB long, more than that and something is probably wrong.
		gerr := &genericRespError{Code: resp.Status}
		if resp.ContentLength != 0 {
			if err := json.NewDecoder(r).Decode(gerr); err != nil {
				gerr.Code = resp.Status
				message, err := ioutil.ReadAll(r)
				if err != nil {
					return nil, err
				}
				gerr.Message = string(message)
			}
		}
		return nil, gerr
	}
	cleanup = func() {} // we don't want to close the body if we got a status code in the 2xx range.
	return &QueryCSVResult{ReadCloser: resp.Body, csvReader: csv.NewReader(resp.Body)}, nil
}

func (c *Client) makeQueryURL(org string) (string, error) {
	qu, err := url.Parse(c.url.String())
	if err != nil {
		return "", err
	}
	qu.Path = path.Join(qu.Path, "query")

	params := qu.Query()
	params.Set("org", org)
	qu.RawQuery = params.Encode()
	return qu.String(), nil
}

// QueryCSVResult is the result of a flux query in CSV format
type QueryCSVResult struct {
	io.ReadCloser
	csvReader   *csv.Reader
	Row         []string
	ColNames    []string
	colNamesMap map[string]int
	dataTypes   []string
	group       []bool
	defaultVals []string
	Err         error
}

// Next iterates to the next row in the data set.  Typically this is called like so:
//
// 	for q.Next(){
// 		... // do thing here
// 	}
//
// It will call Close() on the result when it encounters EOF.
func (q *QueryCSVResult) Next() bool {
	inNameRow := false
readRow:
	q.Row, q.Err = q.csvReader.Read()
	if q.Err == io.EOF {
		q.Err = q.Close()
		return false
	}
	if q.Err != nil {
		return false
	}
	if len(q.Row) == 0 {
		goto readRow
	}
	if len(q.Row) == 1 {
		goto readRow
		//this shouldn't ever happen, we should error here
	}
	switch q.Row[0] {
	case "#datatype":
		// parse datatypes here

		q.dataTypes = q.Row[1:]
		goto readRow
	case "#group":
		q.group = q.group[:0]
		for _, x := range q.Row[1:] {
			if x == "true" {
				q.group = append(q.group, true)
			} else {
				q.group = append(q.group, false)
			}
		}
		goto readRow
	case "#default":
		q.defaultVals = q.Row[1:]
		inNameRow = true
		goto readRow
	case "":
		if inNameRow {
			q.ColNames = q.Row[1:]
			q.colNamesMap = make(map[string]int, len(q.Row[1:]))
			for i := range q.ColNames {
				q.colNamesMap[q.ColNames[i]] = i
			}
			inNameRow = false
			goto readRow
		}

	}
	return true
}

// Unmarshal alows you to easily unmarshal rows of results into your own types.
func (q *QueryCSVResult) Unmarshal(x interface{}) error {
	// TODO(docmerlin): add things besides map kinds
	typeOf := reflect.TypeOf(x)
	xVal := reflect.ValueOf(x)
	if x == nil {
		return errors.New("cannot marshal into a nil object")
	}
	kind := typeOf.Kind()
	switch {
	case kind == reflect.Map:
		if typeOf.Key().Kind() != reflect.String {
			return errors.New("cannot marshal into a map where the key is not of a string kind")
		}
		//TODO(docmerlin): add in other kinds
		elem := typeOf.Elem()
		// easy case
		if elem.Kind() == reflect.String {
			for i, val := range q.Row[1:] {
				xVal.SetMapIndex(
					reflect.ValueOf(q.ColNames[i]),
					reflect.ValueOf(
						stringTernary(val, q.defaultVals[i])))
			}
			return nil
		}
		for i, val := range q.Row[1:] {
			cell, err := convert(stringTernary(val, q.defaultVals[i]), q.dataTypes[i])
			if err != nil {
				return err
			}
			if !reflect.TypeOf(cell).ConvertibleTo(elem) {
				return fmt.Errorf("cannot marshal type %s into type %s", reflect.TypeOf(cell), elem)
			}
			xVal.SetMapIndex(reflect.ValueOf(q.ColNames[i]), reflect.ValueOf(cell))
		}
	case kind == reflect.Ptr && typeOf.Elem().Kind() == reflect.Struct:
		xVal = xVal.Elem()
		typeOf = typeOf.Elem()
		numFields := typeOf.NumField()
		for i := 0; i < numFields; i++ {
			f := typeOf.Field(i)
			name := f.Name
			usedTag := false
			if tag, ok := f.Tag.Lookup("flux"); ok {
				if tag != "" {
					name = tag
					usedTag = true
				}
			}
			if _, nameOk := q.colNamesMap[name]; !nameOk {
				lowerName := strings.ToLower(name)
				if _, lowerNameok := q.colNamesMap[lowerName]; !lowerNameok && !usedTag {
					continue
				} else if lowerNameok && !usedTag {
					name = lowerName
				}
			}
			fVal := xVal.Field(i)
			// ignore fields that are private or otherwise unsetable
			if !fVal.CanSet() {
				continue
			}
			// grab the row by name
			s := q.Row[q.colNamesMap[name]+1]

			fType := f.Type
			if fType.Kind() == reflect.Ptr {
				if s == "" { // handle nil case
					fVal.Set(reflect.New(fType).Elem())
					continue
				}
				fVal.Set(reflect.New(f.Type.Elem()))
				fType = fVal.Elem().Type()
				for fType.Kind() == reflect.Ptr {
					fVal.Elem().Set(reflect.New(fType.Elem()))
					fVal = fVal.Elem()
					fType = fType.Elem()
				}
				fVal = fVal.Elem()
			}
			switch fType.Kind() {
			case reflect.String:
				fVal.SetString(s)
			case reflect.Bool:
				if q.dataTypes[q.colNamesMap[name]] != boolDatatype {
					return errors.New("cannot marshal column into a bool type")
				}
				val := false
				if s == "true" {
					val = true
				}
				fVal.SetBool(val)
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				if s == "" {
					fVal.SetInt(0)
					continue
				}
				val, err := strconv.ParseInt(s, 10, 64)
				if err != nil {
					return fmt.Errorf("cannot marshal column value %s into an int type", s)
				}
				fVal.SetInt(val)
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
				if s == "" {
					fVal.SetUint(0)
					continue
				}
				val, err := strconv.ParseUint(s, 10, 64)
				if err != nil {
					return errors.New("cannot marshal column into an uint type")
				}
				fVal.SetUint(val)
			case reflect.Float32:
				if s == "" {
					fVal.SetFloat(0)
					continue
				}
				val, err := strconv.ParseFloat(s, 32)
				if err != nil {
					return errors.New("cannot marshal column into an float32 type")
				}
				fVal.SetFloat(val)

			case reflect.Float64:
				if s == "" {
					fVal.SetFloat(0)
					continue
				}
				val, err := strconv.ParseFloat(s, 64)
				if err != nil {
					return errors.New("cannot marshal column into an float64 type")
				}
				xVal.Field(i).SetFloat(val)
			case reflect.Struct:
				if fType != reflect.TypeOf(time.Time{}) {
					return errors.New("the only struct supported is a time.Time")
				}
				ts, err := time.Parse(time.RFC3339, s)
				if err != nil {
					return errors.New("cannot marshal column into a time")
				}
				fVal.Set(reflect.ValueOf(ts))
			case reflect.Interface:
				if s == "" {
					fVal.Set(reflect.Zero(fType))
					continue
				}
				x, err := convert(s, q.dataTypes[q.colNamesMap[name]])
				if err != nil {
					return errors.New("badly encoded column")
				}
				if !reflect.TypeOf(x).ConvertibleTo(f.Type) {
					return fmt.Errorf("cannot convert type column to type %s", f.Type)
				}
				fVal.Set(reflect.ValueOf(x))
			}
		}
	case kind == reflect.Struct:
		return errors.New("struct argument must be a pointer")
	default:
		return fmt.Errorf("cannot marshal into a type of %s", typeOf)
	}
	return nil
}

func stringTernary(a, b string) string {
	if a == "" {
		return b
	}
	return a
}

func convert(s, t string) (interface{}, error) {
	switch t {
	case stringDatatype:
		return s, nil
	case timeDatatype, timeDataTypeWithFmt:
		return time.Parse(time.RFC3339, s)
	case floatDatatype:
		return strconv.ParseFloat(s, 64)
	case boolDatatype:
		if s == "false" {
			return false, nil
		}
		return true, nil
	case intDatatype:
		return strconv.ParseInt(s, 10, 64)
	case uintDatatype:
		return strconv.ParseUint(s, 10, 64)
	default:
		return nil, fmt.Errorf("%s has unknown data type %s", s, t)
	}
}
