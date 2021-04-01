// Copyright 2021 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package annotatedcsv_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/influxdata/influxdb-client-go/annotatedcsv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSingleTable(t *testing.T) {
	csvTable := `#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#group,false,false,true,true,false,false,true,true,true,true
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,a,b
,,0,2020-02-17T22:19:49.747562847Z,2020-02-18T22:19:49.747562847Z,2020-02-18T10:34:08.135814545Z,1.4,f,test,1,adsfasdf
,,0,2020-02-17T22:19:49.747562847Z,2020-02-18T22:19:49.747562847Z,2020-02-18T22:08:44.850214724Z,6.6,f,test,1,adsfasdf

`
	tables := []expectedTable{
		{columns: []annotatedcsv.Column{
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
				{
					"_result",
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
	}

	verifyTables(t, csvTable, tables)
}

func TestMultiTables(t *testing.T) {
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
	tablesMultiStructure := []expectedTable{
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
	verifyTables(t, csvTableMultiStructure, tablesMultiStructure)

	// test advancing
	reader := strings.NewReader(csvTableMultiStructure)
	res := annotatedcsv.NewReader(reader)

	require.False(t, res.NextRow())
	require.NoError(t, res.Err())
	require.True(t, res.NextSection())
	require.NoError(t, res.Err())
	require.True(t, res.NextRow())
	require.NoError(t, res.Err())
	var row []interface{}
	require.NoError(t, res.Decode(&row))
	require.Equal(t, tablesMultiStructure[0].rows[0], row)

	reader = strings.NewReader(csvTableMultiStructure)
	res = annotatedcsv.NewReader(reader)

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
	require.Equal(t, tablesMultiStructure[2].rows[0], row)

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
	tablesMultiTables := []expectedTable{
		{
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
				{"_result",
					int64(1),
					mustParseTime("2020-02-17T22:19:49.747562847Z"),
					mustParseTime("2020-02-18T22:19:49.747562847Z"),
					mustParseTime("2020-02-18T10:34:08.135814545Z"),
					4.3,
					"i",
					"test",
					"1",
					"xyxyxyxy",
				},
				{
					"_result",
					int64(1),
					mustParseTime("2020-02-17T22:19:49.747562847Z"),
					mustParseTime("2020-02-18T22:19:49.747562847Z"),
					mustParseTime("2020-02-18T22:08:44.850214724Z"),
					-1.2,
					"i",
					"test",
					"1",
					"xyxyxyxy",
				},
				{"_result",
					int64(2),
					mustParseTime("2020-02-17T22:19:49.747562847Z"),
					mustParseTime("2020-02-18T22:19:49.747562847Z"),
					mustParseTime("2020-02-18T22:08:44.62797864Z"),
					0.1,
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
					0.3,
					"f",
					"test",
					"0",
					"adsfasdf",
				},
				{"_result",
					int64(3),
					mustParseTime("2020-02-17T22:19:49.747562847Z"),
					mustParseTime("2020-02-18T22:19:49.747562847Z"),
					mustParseTime("2020-02-18T22:08:44.62797864Z"),
					float64(10),
					"i",
					"test",
					"0",
					"xyxyxyxy",
				},
				{
					"_result",
					int64(3),
					mustParseTime("2020-02-17T22:19:49.747562847Z"),
					mustParseTime("2020-02-18T22:19:49.747562847Z"),
					mustParseTime("2020-02-18T22:08:44.969100374Z"),
					float64(2),
					"i",
					"test",
					"0",
					"xyxyxyxy",
				},
			},
		},
	}
	verifyTables(t, csvTableMultiTables, tablesMultiTables)

	//test advancing
	reader = strings.NewReader(csvTableMultiTables)
	res = annotatedcsv.NewReader(reader)
	require.True(t, res.NextSection())
	require.NoError(t, res.Err())
	require.False(t, res.NextSection())
	require.NoError(t, res.Err())
}

func TestErrorInRow(t *testing.T) {
	csvTableError := `#datatype,string,string
#group,true,true
#default,,
,error,reference
,failed to create physical plan: invalid time bounds from procedure from: bounds contain zero time,897

`

	tables := []expectedTable{
		{ // Table 1
			[]annotatedcsv.Column{
				{Type: "string", Default: "", Name: "error", Group: true},
				{Type: "string", Default: "", Name: "reference", Group: true},
			},
			[][]interface{}{
				{
					"failed to create physical plan: invalid time bounds from procedure from: bounds contain zero time",
					"897",
				},
			},
		},
	}
	verifyTables(t, csvTableError, tables)

	csvTableErrorNoReference := `#datatype,string,string
#group,true,true
#default,,
,error,reference
,failed to create physical plan: invalid time bounds from procedure from: bounds contain zero time,

`
	tables = []expectedTable{
		{ // Table 1
			[]annotatedcsv.Column{
				{Type: "string", Default: "", Name: "error", Group: true},
				{Type: "string", Default: "", Name: "reference", Group: true},
			},
			[][]interface{}{
				{
					"failed to create physical plan: invalid time bounds from procedure from: bounds contain zero time",
					"",
				},
			},
		},
	}
	verifyTables(t, csvTableErrorNoReference, tables)

	csvTableErrorNoMessage := `#datatype,string,string
#group,true,true
#default,,
,error,reference
,,

`
	tables = []expectedTable{
		{ // Table 1
			[]annotatedcsv.Column{
				{Type: "string", Default: "", Name: "error", Group: true},
				{Type: "string", Default: "", Name: "reference", Group: true},
			},
			[][]interface{}{
				{
					"",
					"",
				},
			},
		},
	}
	verifyTables(t, csvTableErrorNoMessage, tables)
}

func TestInvalidDataType(t *testing.T) {
	csvTable := `#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,int,string,string,string,string
#group,false,false,true,true,false,false,true,true,true,true
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,a,b
,,0,2020-02-17T22:19:49.747562847Z,2020-02-18T22:19:49.747562847Z,2020-02-18T10:34:08.135814545Z,1,f,test,1,adsfasdf
,,0,2020-02-17T22:19:49.747562847Z,2020-02-18T22:19:49.747562847Z,2020-02-18T22:08:44.850214724Z,6,f,test,1,adsfasdf

`
	tables := []expectedTable{
		{[]annotatedcsv.Column{
			{Type: "string", Default: "_result", Name: "result", Group: false},
			{Type: "long", Default: "", Name: "table", Group: false},
			{Type: "dateTime:RFC3339", Default: "", Name: "_start", Group: true},
			{Type: "dateTime:RFC3339", Default: "", Name: "_stop", Group: true},
			{Type: "dateTime:RFC3339", Default: "", Name: "_time", Group: false},
			{Type: "int", Default: "", Name: "_value", Group: false},
			{Type: "string", Default: "", Name: "_field", Group: true},
			{Type: "string", Default: "", Name: "_measurement", Group: true},
			{Type: "string", Default: "", Name: "a", Group: true},
			{Type: "string", Default: "", Name: "b", Group: true},
		},
			[][]interface{}{
				{
					"_result",
					int64(0),
					mustParseTime("2020-02-17T22:19:49.747562847Z"),
					mustParseTime("2020-02-18T22:19:49.747562847Z"),
					mustParseTime("2020-02-18T10:34:08.135814545Z"),
					"1",
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
					"6",
					"f",
					"test",
					"1",
					"adsfasdf",
				},
			},
		},
	}

	verifyTables(t, csvTable, tables)
}

func TestReorderedAnnotations(t *testing.T) {
	csvTable1 := `#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,a,b
,,0,2020-02-17T22:19:49.747562847Z,2020-02-18T22:19:49.747562847Z,2020-02-18T10:34:08.135814545Z,1.4,f,test,1,adsfasdf
,,0,2020-02-17T22:19:49.747562847Z,2020-02-18T22:19:49.747562847Z,2020-02-18T22:08:44.850214724Z,6.6,f,test,1,adsfasdf

`
	tables := []expectedTable{
		{[]annotatedcsv.Column{
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
			[][]interface{}{
				{
					"_result",
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
	}

	verifyTables(t, csvTable1, tables)

	csvTable2 := `#default,_result,,,,,,,,,
#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
,result,table,_start,_stop,_time,_value,_field,_measurement,a,b
,,0,2020-02-17T22:19:49.747562847Z,2020-02-18T22:19:49.747562847Z,2020-02-18T10:34:08.135814545Z,1.4,f,test,1,adsfasdf
,,0,2020-02-17T22:19:49.747562847Z,2020-02-18T22:19:49.747562847Z,2020-02-18T22:08:44.850214724Z,6.6,f,test,1,adsfasdf

`
	verifyTables(t, csvTable2, tables)
}

func TestDatatypeOnlyAnnotation(t *testing.T) {
	csvTable1 := `#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
,result,table,_start,_stop,_time,_value,_field,_measurement,a,b
,,0,2020-02-17T22:19:49.747562847Z,2020-02-18T22:19:49.747562847Z,2020-02-18T10:34:08.135814545Z,1.4,f,test,1,adsfasdf
,,0,2020-02-17T22:19:49.747562847Z,2020-02-18T22:19:49.747562847Z,2020-02-18T22:08:44.850214724Z,6.6,f,test,1,adsfasdf

`
	tables := []expectedTable{
		{[]annotatedcsv.Column{
			{Type: "string", Default: "", Name: "result", Group: false},
			{Type: "long", Default: "", Name: "table", Group: false},
			{Type: "dateTime:RFC3339", Default: "", Name: "_start", Group: false},
			{Type: "dateTime:RFC3339", Default: "", Name: "_stop", Group: false},
			{Type: "dateTime:RFC3339", Default: "", Name: "_time", Group: false},
			{Type: "double", Default: "", Name: "_value", Group: false},
			{Type: "string", Default: "", Name: "_field", Group: false},
			{Type: "string", Default: "", Name: "_measurement", Group: false},
			{Type: "string", Default: "", Name: "a", Group: false},
			{Type: "string", Default: "", Name: "b", Group: false},
		},
			[][]interface{}{
				{
					"",
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
					"",
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
	}

	verifyTables(t, csvTable1, tables)
}

func TestMissingDatatypeAnnotation(t *testing.T) {
	csvTable1 := `
#group,false,false,true,true,false,true,true,false,false,false
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,deviceId,sensor,elapsed,note,start
,,0,2020-04-28T12:36:50.990018157Z,2020-04-28T12:51:50.990018157Z,2020-04-28T12:38:11.480545389Z,1467463,BME280,1m1s,ZGF0YWluYmFzZTY0,2020-04-27T00:00:00Z
,,0,2020-04-28T12:36:50.990018157Z,2020-04-28T12:51:50.990018157Z,2020-04-28T12:39:36.330153686Z,1467463,BME280,1h20m30.13245s,eHh4eHhjY2NjY2NkZGRkZA==,2020-04-28T00:00:00Z
`
	tables := []expectedTable{
		{ // Table 1
			columns: []annotatedcsv.Column{
				{Type: "", Default: "_result", Name: "result", Group: false},
				{Type: "", Default: "", Name: "table", Group: false},
				{Type: "", Default: "", Name: "_start", Group: true},
				{Type: "", Default: "", Name: "_stop", Group: true},
				{Type: "", Default: "", Name: "_time", Group: false},
				{Type: "", Default: "", Name: "deviceId", Group: true},
				{Type: "", Default: "", Name: "sensor", Group: true},
				{Type: "", Default: "", Name: "elapsed", Group: false},
				{Type: "", Default: "", Name: "note", Group: false},
				{Type: "", Default: "", Name: "start", Group: false},
			},
			rows: [][]interface{}{
				{
					"_result",
					"0",
					"2020-04-28T12:36:50.990018157Z",
					"2020-04-28T12:51:50.990018157Z",
					"2020-04-28T12:38:11.480545389Z",
					"1467463",
					"BME280",
					"1m1s",
					"ZGF0YWluYmFzZTY0",
					"2020-04-27T00:00:00Z",
				},
				{
					"_result",
					"0",
					"2020-04-28T12:36:50.990018157Z",
					"2020-04-28T12:51:50.990018157Z",
					"2020-04-28T12:39:36.330153686Z",
					"1467463",
					"BME280",
					"1h20m30.13245s",
					"eHh4eHhjY2NjY2NkZGRkZA==",
					"2020-04-28T00:00:00Z",
				},
			},
		},
	}
	verifyTables(t, csvTable1, tables)

	csvTable2 := `
#default,_result,,,,,,,,,
#group,false,false,true,true,false,true,true,false,false,false
,result,table,_start,_stop,_time,deviceId,sensor,elapsed,note,start
,,0,2020-04-28T12:36:50.990018157Z,2020-04-28T12:51:50.990018157Z,2020-04-28T12:38:11.480545389Z,1467463,BME280,1m1s,ZGF0YWluYmFzZTY0,2020-04-27T00:00:00Z
,,0,2020-04-28T12:36:50.990018157Z,2020-04-28T12:51:50.990018157Z,2020-04-28T12:39:36.330153686Z,1467463,BME280,1h20m30.13245s,eHh4eHhjY2NjY2NkZGRkZA==,2020-04-28T00:00:00Z
`
	verifyTables(t, csvTable2, tables)
}

func TestMissingAnnotations(t *testing.T) {
	csvTable := `
,result,table,_start,_stop,_time,deviceId,sensor,elapsed,note,start
,,0,2020-04-28T12:36:50.990018157Z,2020-04-28T12:51:50.990018157Z,2020-04-28T12:38:11.480545389Z,1467463,BME280,1m1s,ZGF0YWluYmFzZTY0,2020-04-27T00:00:00Z
,,0,2020-04-28T12:36:50.990018157Z,2020-04-28T12:51:50.990018157Z,2020-04-28T12:39:36.330153686Z,1467463,BME280,1h20m30.13245s,eHh4eHhjY2NjY2NkZGRkZA==,2020-04-28T00:00:00Z

`
	tables := []expectedTable{
		{ // Table 1
			columns: []annotatedcsv.Column{
				{Type: "", Default: "", Name: "result", Group: false},
				{Type: "", Default: "", Name: "table", Group: false},
				{Type: "", Default: "", Name: "_start", Group: false},
				{Type: "", Default: "", Name: "_stop", Group: false},
				{Type: "", Default: "", Name: "_time", Group: false},
				{Type: "", Default: "", Name: "deviceId", Group: false},
				{Type: "", Default: "", Name: "sensor", Group: false},
				{Type: "", Default: "", Name: "elapsed", Group: false},
				{Type: "", Default: "", Name: "note", Group: false},
				{Type: "", Default: "", Name: "start", Group: false},
			},
			rows: [][]interface{}{
				{
					"",
					"0",
					"2020-04-28T12:36:50.990018157Z",
					"2020-04-28T12:51:50.990018157Z",
					"2020-04-28T12:38:11.480545389Z",
					"1467463",
					"BME280",
					"1m1s",
					"ZGF0YWluYmFzZTY0",
					"2020-04-27T00:00:00Z",
				},
				{
					"",
					"0",
					"2020-04-28T12:36:50.990018157Z",
					"2020-04-28T12:51:50.990018157Z",
					"2020-04-28T12:39:36.330153686Z",
					"1467463",
					"BME280",
					"1h20m30.13245s",
					"eHh4eHhjY2NjY2NkZGRkZA==",
					"2020-04-28T00:00:00Z",
				},
			},
		},
	}
	verifyTables(t, csvTable, tables)
}

func TestUnexpectedAnnotation(t *testing.T) {
	csvTable := `#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#group,false,false,true,true,false,false,true,true,true,true
#extra,1,1,1,1,1,1,1,1,1,1
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,a,b
,,0,2020-02-17T22:19:49.747562847Z,2020-02-18T22:19:49.747562847Z,2020-02-18T10:34:08.135814545Z,1.4,f,test,1,adsfasdf
,,0,2020-02-17T22:19:49.747562847Z,2020-02-18T22:19:49.747562847Z,2020-02-18T22:08:44.850214724Z,6.6,f,test,1,adsfasdf

`
	tables := []expectedTable{
		{[]annotatedcsv.Column{
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
			[][]interface{}{
				{
					"_result",
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
	}

	verifyTables(t, csvTable, tables)
}

func TestDatetimeConversion(t *testing.T) {
	csvTable := `#datatype,string,long,dateTime:RFC3339,dateTime,dateTime:RFC3339Nano,double,string,string,string,string
#group,false,false,true,true,false,false,true,true,true,true
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,a,b
,,0,2020-02-17T22:19:49.747562847Z,2020-02-18T22:19:49.747562847Z,2020-02-18T10:34:08.135814545Z,1.4,f,test,1,adsfasdf
,,0,2020-02-17T22:19:49.747562847Z,2020-02-18T22:19:49.747562847Z,2020-02-18T22:08:44.850214724Z,6.6,f,test,1,adsfasdf

`
	tables := []expectedTable{
		{[]annotatedcsv.Column{
			{Type: "string", Default: "_result", Name: "result", Group: false},
			{Type: "long", Default: "", Name: "table", Group: false},
			{Type: "dateTime:RFC3339", Default: "", Name: "_start", Group: true},
			{Type: "dateTime", Default: "", Name: "_stop", Group: true},
			{Type: "dateTime:RFC3339Nano", Default: "", Name: "_time", Group: false},
			{Type: "double", Default: "", Name: "_value", Group: false},
			{Type: "string", Default: "", Name: "_field", Group: true},
			{Type: "string", Default: "", Name: "_measurement", Group: true},
			{Type: "string", Default: "", Name: "a", Group: true},
			{Type: "string", Default: "", Name: "b", Group: true},
		},
			[][]interface{}{
				{
					"_result",
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
	}

	verifyTables(t, csvTable, tables)

	//invalid datetime layout
	csvTable = `#datatype,string,long,dateTime:wrongLayout,dateTime,dateTime:RFC3339Nano,double,string,string,string,string
#group,false,false,true,true,false,false,true,true,true,true
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,a,b
,,0,2020-02-17T22:19:49.747562847Z,2020-02-18T22:19:49.747562847Z,2020-02-18T10:34:08.135814545Z,1.4,f,test,1,adsfasdf
,,0,2020-02-17T22:19:49.747562847Z,2020-02-18T22:19:49.747562847Z,2020-02-18T22:08:44.850214724Z,6.6,f,test,1,adsfasdf

`
	tables = []expectedTable{
		{[]annotatedcsv.Column{
			{Type: "string", Default: "_result", Name: "result", Group: false},
			{Type: "long", Default: "", Name: "table", Group: false},
			{Type: "dateTime:wrongLayout", Default: "", Name: "_start", Group: true},
			{Type: "dateTime", Default: "", Name: "_stop", Group: true},
			{Type: "dateTime:RFC3339Nano", Default: "", Name: "_time", Group: false},
			{Type: "double", Default: "", Name: "_value", Group: false},
			{Type: "string", Default: "", Name: "_field", Group: true},
			{Type: "string", Default: "", Name: "_measurement", Group: true},
			{Type: "string", Default: "", Name: "a", Group: true},
			{Type: "string", Default: "", Name: "b", Group: true},
		},
			[][]interface{}{
				{
					"_result",
					int64(0),
					"2020-02-17T22:19:49.747562847Z",
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
					"2020-02-17T22:19:49.747562847Z",
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
	}

	verifyTables(t, csvTable, tables)

}

func TestFailedConversion(t *testing.T) {
	csvTable := `#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#group,false,false,true,true,false,false,true,true,true,true
#default,_result,zero,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,a,b
,,,2020-02-17T22:19:49.747562847Z,2020-02-18T22:19:49.747562847Z,2020-02-18T10:34:08.135814545Z,seven,f,test,1,adsfasdf
,,0,2020-02-17T22:19:49.747562847Z,2020-02-18T22:19:49.747562847Z,2020-02-18T22:08:44.850214724Z,six,f,test,1,adsfasdf

`
	verifyParsingError(t, csvTable, `cannot convert value "zero" in column of type "long" to Go type interface {} at line 5: strconv.ParseInt: parsing "zero": invalid syntax`, false)

	csvTable = `#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#group,false,false,true,true,false,false,true,true,true,true
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,a,b
,,0,2020-02-17T22:19:49.747562847Z,2020-02-18T22:19:49.747562847Z,2020-02-18T10:34:08.135814545Z,seven,f,test,1,adsfasdf
,,0,2020-02-17T22:19:49.747562847Z,2020-02-18T22:19:49.747562847Z,2020-02-18T22:08:44.850214724Z,six,f,test,1,adsfasdf

`
	verifyParsingError(t, csvTable, `cannot convert value "seven" in column of type "double" to Go type interface {} at line 5: strconv.ParseFloat: parsing "seven": invalid syntax`, false)
}

func TestDifferentNumberOfColumns(t *testing.T) {
	// different #columns in group (8)
	csvTable := `#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,int,string,duration,base64Binary,dateTime:RFC3339
#group,false,false,true,true,false,true,true,false
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,deviceId,sensor,elapsed,note,start
,,0,2020-04-28T12:36:50.990018157Z,2020-04-28T12:51:50.990018157Z,2020-04-28T12:38:11.480545389Z,1467463,BME280,1m1s,ZGF0YWluYmFzZTY0,2020-04-27T00:00:00Z
`
	verifyParsingError(t, csvTable, "inconsistent table header (got 8 items want 10)", true)

	// different #columns in default (8)
	csvTable = `#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,int,string,duration,base64Binary,dateTime:RFC3339
#group,false,false,true,true,false,true,true,false,false,true
#default,_result,,,,,,,
,result,table,_start,_stop,_time,deviceId,sensor,elapsed,note,start
,,0,2020-04-28T12:36:50.990018157Z,2020-04-28T12:51:50.990018157Z,2020-04-28T12:38:11.480545389Z,1467463,BME280,1m1s,ZGF0YWluYmFzZTY0,2020-04-27T00:00:00Z
`
	verifyParsingError(t, csvTable, "inconsistent table header (got 8 items want 10)", true)

	// different #columns in dataType(10)
	csvTable = `#default,_result,,,,,,,
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,duration,base64Binary,dateTime:RFC3339
#group,false,false,true,true,false,true,true,false,false,true,true
,result,table,_start,_stop,_time,deviceId,sensor,elapsed,note,start
,,0,2020-04-28T12:36:50.990018157Z,2020-04-28T12:51:50.990018157Z,2020-04-28T12:38:11.480545389Z,1467463,BME280,1m1s,ZGF0YWluYmFzZTY0,2020-04-27T00:00:00Z
`

	verifyParsingError(t, csvTable, "inconsistent table header (got 10 items want 8)", true)

	// different #columns in data row(8)
	csvTable = `#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,int,string,duration,base64Binary,dateTime:RFC3339
#group,false,false,true,true,false,true,true,false,true,false
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,deviceId,sensor,elapsed,note,start
,,0,2020-04-28T12:36:50.990018157Z,2020-04-28T12:51:50.990018157Z,2020-04-28T12:38:11.480545389Z,1467463,BME280,1m1s,ZGF0YWluYmFzZTY0,2020-04-27T00:00:00Z,2345234
`
	verifyParsingError(t, csvTable, "inconsistent number of columns at line 5 (got 11 items want 10)", false)
}

func TestCSVError(t *testing.T) {
	csvErrTable := `#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
,result,"table,_start,_stop,_time,_value,_field,_measurement,a,b
,,0,2020-02-17T22:19:49.747562847Z,2020-02-18T22:19:49.747562847Z,2020-02-18T10:34:08.135814545Z,1.4,f,test,1,adsfasdf
,,0,2020-02-17T22:19:49.747562847Z,2020-02-18T22:19:49.747562847Z,2020-02-18T22:08:44.850214724Z,6.6,f,test,1,adsfasdf
`
	reader := strings.NewReader(csvErrTable)
	res := annotatedcsv.NewReader(reader)

	require.False(t, res.NextSection())
	require.NotNil(t, res.Err())

	csvErrTable = `#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
,result,table,_start,_stop,_time,_value,_field,_measurement,a,b
,",0,2020-02-17T22:19:49.747562847Z,2020-02-18T22:19:49.747562847Z,2020-02-18T10:34:08.135814545Z,1.4,f,test,1,adsfasdf
,,0,2020-02-17T22:19:49.747562847Z,2020-02-18T22:19:49.747562847Z,2020-02-18T22:08:44.850214724Z,6.6,f,test,1,adsfasdf
`
	reader = strings.NewReader(csvErrTable)
	res = annotatedcsv.NewReader(reader)

	require.True(t, res.NextSection())
	require.False(t, res.NextRow())
	require.NotNil(t, res.Err())
}

func verifyParsingError(t *testing.T, csvTable, error string, inHeader bool) {
	reader := strings.NewReader(csvTable)
	res := annotatedcsv.NewReader(reader)

	if inHeader {
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
}

func verifyTables(t *testing.T, csvTable string, expectedTables []expectedTable) {
	reader := strings.NewReader(csvTable)
	res := annotatedcsv.NewReader(reader)

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
