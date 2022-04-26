// Copyright 2021 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package influxclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"reflect"
	"time"
)

// dialect defines attributes of Flux query response header.
type dialect struct {
	Annotations []string `json:"annotations"`
	Delimiter   string   `json:"delimiter"`
	Header      bool     `json:"header"`
}

//  queryBody holds the body for an HTTP query request.
type queryBody struct {
	Dialect dialect     `json:"dialect"`
	Query   string      `json:"query"`
	Type    string      `json:"type"`
	Params  interface{} `json:"params,omitempty"`
}

// defaultDialect is queryBody dialect value with all annotations, name header and comma as delimiter
var defaultDialect = dialect{
	Annotations: []string{"datatype", "default", "group"},
	Delimiter:   ",",
	Header:      true,
}

// Query sends the given Flux query to server and returns QueryResultReader for further parsing result.
// The result must be closed after use.
//
// Flux query can contain reference to parameters, which must be passed via queryParams.
// it can be a struct or map. Param values can be only simple types or time.Time.
// Name of a struct field or a map key (must be a string) will be a param name.
//
// Fields of a struct can be more specified by json annotations:
//
//   type Condition struct {
//      Start  time.Time  `json:"start"`
//      Field  string     `json:"field"`
//      Value  float64    `json:"value"`
//   }
//
//   cond  := Condition {
//	  "iot_center",
//	  "Temperature",
//	  "-10m",
//	  23.0,
//  }
//
// Parameters are then accessed via the params object:
//
//  query:= `from(bucket: "environment")
//	 |> range(start: time(v: params.start))
//	 |> filter(fn: (r) => r._measurement == "air")
//	 |> filter(fn: (r) => r._field == params.field)
//	 |> filter(fn: (r) => r._value > params.value)`
//
// And used in the call to Query:
//
//  result, err := client.Query(ctx, query, cond);
//
// Use QueryResultReader.NextSection() for navigation to the sections in the query result set.
//
// Use QueryResultReader.NextRow() for iterating over rows in the section.
//
// Read the row raw data using QueryResultReader.Row()
// or deserialize data into a struct or a slice via QueryResultReader.Decode:
//
//  val := &struct {
//	 Time  time.Time	`csv:"_time"`
//	 Temp  float64		`csv:"_value"`
//	 Sensor string		`csv:"sensor"`
//  }{}
//  err = result.Decode(val)
//
func (c *Client) Query(ctx context.Context, query string, queryParams interface{}) (*QueryResultReader, error) {
	if err := checkParamsType(queryParams); err != nil {
		return nil, err
	}
	queryURL, _ := c.apiURL.Parse("query")

	q := queryBody{
		Dialect: defaultDialect,
		Query:   query,
		Type:    "flux",
		Params:  queryParams,
	}
	qrJSON, err := json.Marshal(q)
	if err != nil {
		return nil, err
	}
	resp, err := c.makeAPICall(ctx, httpParams{
		endpointURL: queryURL,
		httpMethod:  "POST",
		headers:     map[string]string{"Content-Type": "application/json"},
		queryParams: url.Values{"org": []string{c.params.Organization}},
		body:        bytes.NewReader(qrJSON),
	})
	if err != nil {
		return nil, err
	}

	return NewQueryResultReader(resp.Body), nil
}

// checkParamsType validates the value is struct with simple type fields
// or a map with key as string and value as a simple type
func checkParamsType(p interface{}) error {
	if p == nil {
		return nil
	}
	t := reflect.TypeOf(p)
	v := reflect.ValueOf(p)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		v = v.Elem()
	}
	if t.Kind() != reflect.Struct && t.Kind() != reflect.Map {
		return fmt.Errorf("cannot use %v as query params", t)
	}
	switch t.Kind() {
	case reflect.Struct:
		fields := reflect.VisibleFields(t)
		for _, f := range fields {
			fv := v.FieldByIndex(f.Index)
			t := getFieldType(fv)
			if !validParamType(t) {
				return fmt.Errorf("cannot use field '%s' of type '%v' as a query param", f.Name, t)
			}

		}
	case reflect.Map:
		key := t.Key()
		if key.Kind() != reflect.String {
			return fmt.Errorf("cannot use map key of type '%v' for query param name", key)
		}
		for _, k := range v.MapKeys() {
			f := v.MapIndex(k)
			t := getFieldType(f)
			if !validParamType(t) {
				return fmt.Errorf("cannot use map value type '%v' as a query param", t)
			}
		}
	}
	return nil
}

// getFieldType extracts type of value
func getFieldType(v reflect.Value) reflect.Type {
	t := v.Type()
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		v = v.Elem()
	}
	if t.Kind() == reflect.Interface && !v.IsNil() {
		t = reflect.ValueOf(v.Interface()).Type()
	}
	return t
}

// timeType is the exact type for the Time
var timeType = reflect.TypeOf(time.Time{})

// validParamType validates that t is primitive type or string or interface
func validParamType(t reflect.Type) bool {
	return (t.Kind() > reflect.Invalid && t.Kind() < reflect.Complex64) ||
		t.Kind() == reflect.String ||
		t == timeType
}
