// Copyright 2021 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package influxclient_test

import (
	"fmt"
	"io/ioutil"
	"strings"
	"testing"
	"time"

	"github.com/influxdata/influxdb-client-go/annotatedcsv"
	influxclient "github.com/influxdata/influxdb-client-go/inluxclient"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMultiSections(t *testing.T) {
	csvTableMultiStructure := `#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#group,false,false,true,true,false,false,true,true,true,true
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,a,b
,,0,2020-02-17T22:19:49.747562847Z,2020-02-18T22:19:49.747562847Z,2020-02-18T10:34:08.135814545Z,1.4,f,test,1,adsfasdf
,,0,2020-02-17T22:19:49.747562847Z,2020-02-18T22:19:49.747562847Z,2020-02-18T22:08:44.850214724Z,6.6,f,test,1,adsfasdf

#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#group,false,false,true,true,false,false,true,true,true,true
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,a,b
,,1,2020-02-17T22:19:49.747562847Z,2020-02-18T22:19:49.747562847Z,2020-02-18T10:34:08.135814545Z,4,i,test,1,adsfasdf
,,1,2020-02-17T22:19:49.747562847Z,2020-02-18T22:19:49.747562847Z,2020-02-18T22:08:44.850214724Z,-1,i,test,1,adsfasdf

#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,boolean,string,string,string,string
#group,false,false,true,true,false,false,true,true,true,true
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,a,b
,,2,2020-02-17T22:19:49.747562847Z,2020-02-18T22:19:49.747562847Z,2020-02-18T22:08:44.62797864Z,false,f,test,0,adsfasdf
,,2,2020-02-17T22:19:49.747562847Z,2020-02-18T22:19:49.747562847Z,2020-02-18T22:08:44.969100374Z,true,f,test,0,adsfasdf

#datatype,string,long,dateTime:RFC3339Nano,dateTime:RFC3339Nano,dateTime:RFC3339Nano,unsignedLong,string,string,string,string
#group,false,false,true,true,false,false,true,true,true,true
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,a,b
,,3,2020-02-17T22:19:49.747562847Z,2020-02-18T22:19:49.747562847Z,2020-02-18T22:08:44.62797864Z,0,i,test,0,adsfasdf
,,3,2020-02-17T22:19:49.747562847Z,2020-02-18T22:19:49.747562847Z,2020-02-18T22:08:44.969100374Z,2,i,test,0,adsfasdf

`
	tables := []expectedTable{
		{ // Table 1
			columns: []annotatedcsv.Column{
				{Type: "string", Default: "_result", Name: "result", Group: false},
				{Type: "long", Default: "", Name: "table", Group: false},
				{Type: "dateTime:RFC3339", Default: "", Name: "_start", Group: true},
				{Type: "dateTime:RFC3339", Default: "", Name: "_stop", Group: true},
				{Type: "dateTime:RFC3339", Default: "", Name: "_time", Group: false},
				{Type: "double", Default: "", Name: "_value", Group: false},
				{Type: "string", Default: "", Name: "_field", Group: true},
				{Type: "string", Default: "", Name: "_measurement", Group: true},
				{Type: "string", Default: "", Name: "a", Group: true},
				{Type: "string", Default: "", Name: "b", Group: true},
			},
			rows: [][]interface{}{
				{"_result",
					int64(0),
					mustParseTime("2020-02-17T22:19:49.747562847Z"),
					mustParseTime("2020-02-18T22:19:49.747562847Z"),
					mustParseTime("2020-02-18T10:34:08.135814545Z"),
					1.4,
					"f",
					"test",
					"1",
					"adsfasdf",
				},
				{
					"_result",
					int64(0),
					mustParseTime("2020-02-17T22:19:49.747562847Z"),
					mustParseTime("2020-02-18T22:19:49.747562847Z"),
					mustParseTime("2020-02-18T22:08:44.850214724Z"),
					6.6,
					"f",
					"test",
					"1",
					"adsfasdf",
				},
			},
		},
		{ //Table 2
			columns: []annotatedcsv.Column{
				{Type: "string", Default: "_result", Name: "result", Group: false},
				{Type: "long", Default: "", Name: "table", Group: false},
				{Type: "dateTime:RFC3339", Default: "", Name: "_start", Group: true},
				{Type: "dateTime:RFC3339", Default: "", Name: "_stop", Group: true},
				{Type: "dateTime:RFC3339", Default: "", Name: "_time", Group: false},
				{Type: "long", Default: "", Name: "_value", Group: false},
				{Type: "string", Default: "", Name: "_field", Group: true},
				{Type: "string", Default: "", Name: "_measurement", Group: true},
				{Type: "string", Default: "", Name: "a", Group: true},
				{Type: "string", Default: "", Name: "b", Group: true},
			},
			rows: [][]interface{}{
				{"_result",
					int64(1),
					mustParseTime("2020-02-17T22:19:49.747562847Z"),
					mustParseTime("2020-02-18T22:19:49.747562847Z"),
					mustParseTime("2020-02-18T10:34:08.135814545Z"),
					int64(4),
					"i",
					"test",
					"1",
					"adsfasdf",
				},
				{
					"_result",
					int64(1),
					mustParseTime("2020-02-17T22:19:49.747562847Z"),
					mustParseTime("2020-02-18T22:19:49.747562847Z"),
					mustParseTime("2020-02-18T22:08:44.850214724Z"),
					int64(-1),
					"i",
					"test",
					"1",
					"adsfasdf",
				},
			},
		},
		{ // Table 3
			columns: []annotatedcsv.Column{
				{Type: "string", Default: "_result", Name: "result", Group: false},
				{Type: "long", Default: "", Name: "table", Group: false},
				{Type: "dateTime:RFC3339", Default: "", Name: "_start", Group: true},
				{Type: "dateTime:RFC3339", Default: "", Name: "_stop", Group: true},
				{Type: "dateTime:RFC3339", Default: "", Name: "_time", Group: false},
				{Type: "boolean", Default: "", Name: "_value", Group: false},
				{Type: "string", Default: "", Name: "_field", Group: true},
				{Type: "string", Default: "", Name: "_measurement", Group: true},
				{Type: "string", Default: "", Name: "a", Group: true},
				{Type: "string", Default: "", Name: "b", Group: true},
			},
			rows: [][]interface{}{
				{"_result",
					int64(2),
					mustParseTime("2020-02-17T22:19:49.747562847Z"),
					mustParseTime("2020-02-18T22:19:49.747562847Z"),
					mustParseTime("2020-02-18T22:08:44.62797864Z"),
					false,
					"f",
					"test",
					"0",
					"adsfasdf",
				},
				{
					"_result",
					int64(2),
					mustParseTime("2020-02-17T22:19:49.747562847Z"),
					mustParseTime("2020-02-18T22:19:49.747562847Z"),
					mustParseTime("2020-02-18T22:08:44.969100374Z"),
					true,
					"f",
					"test",
					"0",
					"adsfasdf",
				},
			},
		},
		{ //Table 4
			columns: []annotatedcsv.Column{
				{Type: "string", Default: "_result", Name: "result", Group: false},
				{Type: "long", Default: "", Name: "table", Group: false},
				{Type: "dateTime:RFC3339Nano", Default: "", Name: "_start", Group: true},
				{Type: "dateTime:RFC3339Nano", Default: "", Name: "_stop", Group: true},
				{Type: "dateTime:RFC3339Nano", Default: "", Name: "_time", Group: false},
				{Type: "unsignedLong", Default: "", Name: "_value", Group: false},
				{Type: "string", Default: "", Name: "_field", Group: true},
				{Type: "string", Default: "", Name: "_measurement", Group: true},
				{Type: "string", Default: "", Name: "a", Group: true},
				{Type: "string", Default: "", Name: "b", Group: true},
			},
			rows: [][]interface{}{
				{"_result",
					int64(3),
					mustParseTime("2020-02-17T22:19:49.747562847Z"),
					mustParseTime("2020-02-18T22:19:49.747562847Z"),
					mustParseTime("2020-02-18T22:08:44.62797864Z"),
					uint64(0),
					"i",
					"test",
					"0",
					"adsfasdf",
				},
				{
					"_result",
					int64(3),
					mustParseTime("2020-02-17T22:19:49.747562847Z"),
					mustParseTime("2020-02-18T22:19:49.747562847Z"),
					mustParseTime("2020-02-18T22:08:44.969100374Z"),
					uint64(2),
					"i",
					"test",
					"0",
					"adsfasdf",
				},
			},
		},
	}

	//verify full content
	verifyTables(t, csvTableMultiStructure, tables)

	// test advancing
	reader := strings.NewReader(csvTableMultiStructure)
	res := influxclient.NewQueryResultReader(ioutil.NopCloser(reader))

	//test skip first table header
	require.True(t, res.NextRow())
	require.NoError(t, res.Err())
	var row []interface{}
	require.NoError(t, res.Decode(&row))
	require.Equal(t, tables[0].rows[0], row)
	require.Nil(t, res.Close())

	reader = strings.NewReader(csvTableMultiStructure)
	res = influxclient.NewQueryResultReader(ioutil.NopCloser(reader))

	//test skip tables
	require.True(t, res.NextSection())
	require.NoError(t, res.Err())
	require.True(t, res.NextSection())
	require.NoError(t, res.Err())
	require.True(t, res.NextRow())
	require.NoError(t, res.Err())
	require.True(t, res.NextSection())
	require.NoError(t, res.Err())
	require.True(t, res.NextRow())
	require.NoError(t, res.Err())
	require.NoError(t, res.Decode(&row))
	require.Equal(t, tables[2].rows[0], row)

	require.Nil(t, res.Close())

	csvTableMultiTables := `#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#group,false,false,true,true,false,false,true,true,true,true
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,a,b
,,0,2020-02-17T22:19:49.747562847Z,2020-02-18T22:19:49.747562847Z,2020-02-18T10:34:08.135814545Z,1.4,f,test,1,adsfasdf
,,0,2020-02-17T22:19:49.747562847Z,2020-02-18T22:19:49.747562847Z,2020-02-18T22:08:44.850214724Z,6.6,f,test,1,adsfasdf
,,1,2020-02-17T22:19:49.747562847Z,2020-02-18T22:19:49.747562847Z,2020-02-18T10:34:08.135814545Z,4.3,i,test,1,xyxyxyxy
,,1,2020-02-17T22:19:49.747562847Z,2020-02-18T22:19:49.747562847Z,2020-02-18T22:08:44.850214724Z,-1.2,i,test,1,xyxyxyxy
,,2,2020-02-17T22:19:49.747562847Z,2020-02-18T22:19:49.747562847Z,2020-02-18T22:08:44.62797864Z,0.1,f,test,0,adsfasdf
,,2,2020-02-17T22:19:49.747562847Z,2020-02-18T22:19:49.747562847Z,2020-02-18T22:08:44.969100374Z,0.3,f,test,0,adsfasdf
,,3,2020-02-17T22:19:49.747562847Z,2020-02-18T22:19:49.747562847Z,2020-02-18T22:08:44.62797864Z,10,i,test,0,xyxyxyxy
,,3,2020-02-17T22:19:49.747562847Z,2020-02-18T22:19:49.747562847Z,2020-02-18T22:08:44.969100374Z,2,i,test,0,xyxyxyxy

`
	reader = strings.NewReader(csvTableMultiTables)
	res = influxclient.NewQueryResultReader(ioutil.NopCloser(reader))

	//test skip first table header
	require.True(t, res.NextRow())
	require.True(t, res.NextRow())
	require.NoError(t, res.Err())
	row = row[:0]
	require.NoError(t, res.Decode(&row))
	require.Equal(t, tables[0].rows[1], row)
	require.Nil(t, res.Close())

	// skip sections
	reader = strings.NewReader(csvTableMultiTables)
	res = influxclient.NewQueryResultReader(ioutil.NopCloser(reader))
	require.True(t, res.NextSection())
	require.NoError(t, res.Err())
	require.False(t, res.NextSection())
	require.NoError(t, res.Err())
	require.Nil(t, res.Close())
}

func TestErrorInRow(t *testing.T) {
	csvTableError := `#datatype,string,long
#group,true,true
#default,,
,error,reference
,failed to create physical plan: invalid time bounds from procedure from: bounds contain zero time,897

`
	verifyParsingError(t, csvTableError, "Flux query error (code 897): failed to create physical plan: invalid time bounds from procedure from: bounds contain zero time", true)

	csvTableErrorNoReference := `#datatype,string,long
#group,true,true
#default,,
,error,reference
,failed to create physical plan: invalid time bounds from procedure from: bounds contain zero time,

`
	verifyParsingError(t, csvTableErrorNoReference, `cannot decode row in error section: cannot convert value "" in column of type "long" to Go type int64 at line 5:2: strconv.ParseInt: parsing "": invalid syntax`, true)

	csvTableErrorNoMessage := `#datatype,string,long
#group,true,true
#default,,
,error,reference
,,

`
	verifyParsingError(t, csvTableErrorNoMessage, "no row found in error section", true)

	csvTableErrorNoRows := `#datatype,string,string
#group,true,true
#default,,
,error,reference

`
	verifyParsingError(t, csvTableErrorNoRows, "no row found in error section", true)

	csvTableErrorInvalidReferenceType := `#datatype,string,string
#group,true,true
#default,,
,error,reference
,failed to create physical plan: invalid time bounds from procedure from: bounds contain zero time,897

`
	verifyParsingError(t, csvTableErrorInvalidReferenceType, "cannot decode row in error section: cannot convert from column type string to int64", true)

	csvTableErrorNoColumnReference := `#datatype,string
#group,true
#default,
,error
,failed to create physical plan: invalid time bounds from procedure from: bounds contain zero time

`
	tables := []expectedTable{
		{ // Table 1
			[]annotatedcsv.Column{
				{Type: "string", Default: "", Name: "error", Group: true},
			},
			[][]interface{}{
				{
					"failed to create physical plan: invalid time bounds from procedure from: bounds contain zero time",
				},
			},
		},
	}
	verifyTables(t, csvTableErrorNoColumnReference, tables)
}

func TestCSVError(t *testing.T) {
	csvErrInHeader := `#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
,result,"table,_start,_stop,_time,_value,_field,_measurement,a,b
,,0,2020-02-17T22:19:49.747562847Z,2020-02-18T22:19:49.747562847Z,2020-02-18T10:34:08.135814545Z,1.4,f,test,1,adsfasdf
,,0,2020-02-17T22:19:49.747562847Z,2020-02-18T22:19:49.747562847Z,2020-02-18T22:08:44.850214724Z,6.6,f,test,1,adsfasdf

`
	reader := strings.NewReader(csvErrInHeader)
	res := influxclient.NewQueryResultReader(ioutil.NopCloser(reader))

	require.False(t, res.NextSection())
	require.Error(t, res.Err())
	require.Nil(t, res.Close())

	reader = strings.NewReader(csvErrInHeader)
	res = influxclient.NewQueryResultReader(ioutil.NopCloser(reader))
	//straight to error
	require.False(t, res.NextRow())
	require.Error(t, res.Err())
	require.Nil(t, res.Close())

	csvErrInRow := `#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
,result,table,_start,_stop,_time,_value,_field,_measurement,a,b
,",0,2020-02-17T22:19:49.747562847Z,2020-02-18T22:19:49.747562847Z,2020-02-18T10:34:08.135814545Z,1.4,f,test,1,adsfasdf
,,0,2020-02-17T22:19:49.747562847Z,2020-02-18T22:19:49.747562847Z,2020-02-18T22:08:44.850214724Z,6.6,f,test,1,adsfasdf

`
	reader = strings.NewReader(csvErrInRow)
	res = influxclient.NewQueryResultReader(ioutil.NopCloser(reader))

	require.True(t, res.NextSection())
	require.False(t, res.NextRow())
	require.Error(t, res.Err())
	require.Nil(t, res.Close())
}
func verifyParsingError(t *testing.T, csvTable, error string, inHeader bool) {
	reader := strings.NewReader(csvTable)
	res := influxclient.NewQueryResultReader(ioutil.NopCloser(reader))

	if inHeader {
		require.False(t, res.NextSection())
		//repeatedly
		require.False(t, res.NextSection())
		require.False(t, res.NextRow())
		require.Error(t, res.Err())
		assert.Equal(t, error, res.Err().Error())
	} else {
		require.True(t, res.NextSection())
		if res.NextRow() {
			require.NoError(t, res.Err())
			var row []interface{}
			err := res.Decode(&row)
			require.Error(t, err)
			assert.Equal(t, error, err.Error())
		} else {
			require.Error(t, res.Err())
			assert.Equal(t, error, res.Err().Error())
		}
	}
	require.NoError(t, res.Close())
}

func verifyTables(t *testing.T, csvTable string, expectedTables []expectedTable) {
	reader := strings.NewReader(csvTable)
	res := influxclient.NewQueryResultReader(ioutil.NopCloser(reader))

	for _, table := range expectedTables {
		require.True(t, res.NextSection(), res.Err())
		require.NoError(t, res.Err())
		require.Equal(t, table.columns, res.Columns())
		for _, row := range table.rows {
			require.True(t, res.NextRow(), res.Err())
			require.NoError(t, res.Err())
			var r []interface{}
			err := res.Decode(&r)
			require.NoError(t, err)
			for i, v := range r {
				if table.columns[i].Type == "base64Binary" {
					require.Equal(t, row[i], v)
				} else {
					require.True(t, row[i] == v, fmt.Sprintf("%v vs %v", row[i], v))
				}
			}
		}
		require.False(t, res.NextRow(), res.Err())
		require.NoError(t, res.Err())
	}

	require.False(t, res.NextSection(), res.Err())
	require.NoError(t, res.Err())

	require.Nil(t, res.Columns())
	require.Nil(t, res.Row())
	require.NoError(t, res.Close())
}

// ExpectedTable represents table for comparison with parsed tables
type expectedTable struct {
	columns []annotatedcsv.Column
	rows    [][]interface{}
}

// MustParseTime returns  parsed dateTime in RFC3339 and panics if it fails
func mustParseTime(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return t
}
