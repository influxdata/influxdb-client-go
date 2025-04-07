// Copyright 2021 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package influxclient

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/url"
)

// dialect defines attributes of Flux query response header.
type dialect struct {
	Annotations []string `json:"annotations"`
	Delimiter   string   `json:"delimiter"`
	Header      bool     `json:"header"`
}

// queryBody holds the body for an HTTP query request.
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
//	  type Condition struct {
//	     Start  time.Time  `json:"start"`
//	     Field  string     `json:"field"`
//	     Value  float64    `json:"value"`
//	  }
//
//	  cond  := Condition {
//		  "iot_center",
//		  "Temperature",
//		  "-10m",
//		  23.0,
//	 }
//
// Parameters are then accessed via the params object:
//
//	 query:= `from(bucket: "environment")
//		 |> range(start: time(v: params.start))
//		 |> filter(fn: (r) => r._measurement == "air")
//		 |> filter(fn: (r) => r._field == params.field)
//		 |> filter(fn: (r) => r._value > params.value)`
//
// And used in the call to Query:
//
//	result, err := client.Query(ctx, query, cond);
//
// Use QueryResultReader.NextSection() for navigation to the sections in the query result set.
//
// Use QueryResultReader.NextRow() for iterating over rows in the section.
//
// Read the row raw data using QueryResultReader.Row()
// or deserialize data into a struct or a slice via QueryResultReader.Decode:
//
//	 val := &struct {
//		 Time  time.Time	`csv:"_time"`
//		 Temp  float64		`csv:"_value"`
//		 Sensor string		`csv:"sensor"`
//	 }{}
//	 err = result.Decode(val)
func (c *Client) Query(ctx context.Context, query string, queryParams interface{}) (*QueryResultReader, error) {
	if err := checkContainerType(queryParams, true, "query param"); err != nil {
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
		headers:     http.Header{"Content-Type": {"application/json"}},
		queryParams: url.Values{"org": []string{c.params.Organization}},
		body:        bytes.NewReader(qrJSON),
	})
	if err != nil {
		return nil, err
	}

	return NewQueryResultReader(resp.Body), nil
}
