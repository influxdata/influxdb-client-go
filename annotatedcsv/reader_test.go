package annotatedcsv_test

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"testing"
	"time"

	"github.com/influxdata/influxdb-client-go/annotatedcsv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type expectedTable struct {
	columns []annotatedcsv.Column
	rows    [][]interface{}
}

func TestQueryResultSingleTable(t *testing.T) {
	csvTable := `#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#group,false,false,true,true,false,false,true,true,true,true
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,a,b
,,0,2020-02-17T22:19:49.747562847Z,2020-02-18T22:19:49.747562847Z,2020-02-18T10:34:08.135814545Z,1.4,f,test,1,adsfasdf
,,0,2020-02-17T22:19:49.747562847Z,2020-02-18T22:19:49.747562847Z,2020-02-18T22:08:44.850214724Z,6.6,f,test,1,adsfasdf

`
	tables := []expectedTable{
		{[]annotatedcsv.Column{
			{Type: "string", Default: "_result", Name: "result", Group: false},
			{Type: "long", Default: nil, Name: "table", Group: false},
			{Type: "dateTime:RFC3339", Default: nil, Name: "_start", Group: true},
			{Type: "dateTime:RFC3339", Default: nil, Name: "_stop", Group: true},
			{Type: "dateTime:RFC3339", Default: nil, Name: "_time", Group: false},
			{Type: "double", Default: nil, Name: "_value", Group: false},
			{Type: "string", Default: nil, Name: "_field", Group: true},
			{Type: "string", Default: nil, Name: "_measurement", Group: true},
			{Type: "string", Default: nil, Name: "a", Group: true},
			{Type: "string", Default: nil, Name: "b", Group: true},
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

func TestQueryResultMultiTables(t *testing.T) {
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
			[]annotatedcsv.Column{
				{Type: "string", Default: "_result", Name: "result", Group: false},
				{Type: "long", Default: nil, Name: "table", Group: false},
				{Type: "dateTime:RFC3339", Default: nil, Name: "_start", Group: true},
				{Type: "dateTime:RFC3339", Default: nil, Name: "_stop", Group: true},
				{Type: "dateTime:RFC3339", Default: nil, Name: "_time", Group: false},
				{Type: "double", Default: nil, Name: "_value", Group: false},
				{Type: "string", Default: nil, Name: "_field", Group: true},
				{Type: "string", Default: nil, Name: "_measurement", Group: true},
				{Type: "string", Default: nil, Name: "a", Group: true},
				{Type: "string", Default: nil, Name: "b", Group: true},
			},
			[][]interface{}{
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
			[]annotatedcsv.Column{
				{Type: "string", Default: "_result", Name: "result", Group: false},
				{Type: "long", Default: nil, Name: "table", Group: false},
				{Type: "dateTime:RFC3339", Default: nil, Name: "_start", Group: true},
				{Type: "dateTime:RFC3339", Default: nil, Name: "_stop", Group: true},
				{Type: "dateTime:RFC3339", Default: nil, Name: "_time", Group: false},
				{Type: "long", Default: nil, Name: "_value", Group: false},
				{Type: "string", Default: nil, Name: "_field", Group: true},
				{Type: "string", Default: nil, Name: "_measurement", Group: true},
				{Type: "string", Default: nil, Name: "a", Group: true},
				{Type: "string", Default: nil, Name: "b", Group: true},
			},
			[][]interface{}{
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
			[]annotatedcsv.Column{
				{Type: "string", Default: "_result", Name: "result", Group: false},
				{Type: "long", Default: nil, Name: "table", Group: false},
				{Type: "dateTime:RFC3339", Default: nil, Name: "_start", Group: true},
				{Type: "dateTime:RFC3339", Default: nil, Name: "_stop", Group: true},
				{Type: "dateTime:RFC3339", Default: nil, Name: "_time", Group: false},
				{Type: "boolean", Default: nil, Name: "_value", Group: false},
				{Type: "string", Default: nil, Name: "_field", Group: true},
				{Type: "string", Default: nil, Name: "_measurement", Group: true},
				{Type: "string", Default: nil, Name: "a", Group: true},
				{Type: "string", Default: nil, Name: "b", Group: true},
			},
			[][]interface{}{
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
			[]annotatedcsv.Column{
				{Type: "string", Default: "_result", Name: "result", Group: false},
				{Type: "long", Default: nil, Name: "table", Group: false},
				{Type: "dateTime:RFC3339Nano", Default: nil, Name: "_start", Group: true},
				{Type: "dateTime:RFC3339Nano", Default: nil, Name: "_stop", Group: true},
				{Type: "dateTime:RFC3339Nano", Default: nil, Name: "_time", Group: false},
				{Type: "unsignedLong", Default: nil, Name: "_value", Group: false},
				{Type: "string", Default: nil, Name: "_field", Group: true},
				{Type: "string", Default: nil, Name: "_measurement", Group: true},
				{Type: "string", Default: nil, Name: "a", Group: true},
				{Type: "string", Default: nil, Name: "b", Group: true},
			},
			[][]interface{}{
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
			[]annotatedcsv.Column{
				{Type: "string", Default: "_result", Name: "result", Group: false},
				{Type: "long", Default: nil, Name: "table", Group: false},
				{Type: "dateTime:RFC3339", Default: nil, Name: "_start", Group: true},
				{Type: "dateTime:RFC3339", Default: nil, Name: "_stop", Group: true},
				{Type: "dateTime:RFC3339", Default: nil, Name: "_time", Group: false},
				{Type: "double", Default: nil, Name: "_value", Group: false},
				{Type: "string", Default: nil, Name: "_field", Group: true},
				{Type: "string", Default: nil, Name: "_measurement", Group: true},
				{Type: "string", Default: nil, Name: "a", Group: true},
				{Type: "string", Default: nil, Name: "b", Group: true},
			},
			[][]interface{}{
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
}

func TestAdvanceInTable(t *testing.T) {
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

d
#datatype,string,long,dateTime:RFC3339Nano,dateTime:RFC3339Nano,dateTime:RFC3339Nano,unsignedLong,string,string,string,string
#group,false,false,true,true,false,false,true,true,true,true
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,a,b
,,3,2020-02-17T22:19:49.747562847Z,2020-02-18T22:19:49.747562847Z,2020-02-18T22:08:44.62797864Z,0,i,test,0,adsfasdf
,,3,2020-02-17T22:19:49.747562847Z,2020-02-18T22:19:49.747562847Z,2020-02-18T22:08:44.969100374Z,2,i,test,0,adsfasdf

`
	tables := []expectedTable{
		{ // Table 1
			[]annotatedcsv.Column{
				{Type: "string", Default: "_result", Name: "result", Group: false},
				{Type: "long", Default: nil, Name: "table", Group: false},
				{Type: "dateTime:RFC3339", Default: nil, Name: "_start", Group: true},
				{Type: "dateTime:RFC3339", Default: nil, Name: "_stop", Group: true},
				{Type: "dateTime:RFC3339", Default: nil, Name: "_time", Group: false},
				{Type: "double", Default: nil, Name: "_value", Group: false},
				{Type: "string", Default: nil, Name: "_field", Group: true},
				{Type: "string", Default: nil, Name: "_measurement", Group: true},
				{Type: "string", Default: nil, Name: "a", Group: true},
				{Type: "string", Default: nil, Name: "b", Group: true},
			},
			[][]interface{}{
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
			[]annotatedcsv.Column{
				{Type: "string", Default: "_result", Name: "result", Group: false},
				{Type: "long", Default: nil, Name: "table", Group: false},
				{Type: "dateTime:RFC3339", Default: nil, Name: "_start", Group: true},
				{Type: "dateTime:RFC3339", Default: nil, Name: "_stop", Group: true},
				{Type: "dateTime:RFC3339", Default: nil, Name: "_time", Group: false},
				{Type: "long", Default: nil, Name: "_value", Group: false},
				{Type: "string", Default: nil, Name: "_field", Group: true},
				{Type: "string", Default: nil, Name: "_measurement", Group: true},
				{Type: "string", Default: nil, Name: "a", Group: true},
				{Type: "string", Default: nil, Name: "b", Group: true},
			},
			[][]interface{}{
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
			[]annotatedcsv.Column{
				{Type: "string", Default: "_result", Name: "result", Group: false},
				{Type: "long", Default: nil, Name: "table", Group: false},
				{Type: "dateTime:RFC3339", Default: nil, Name: "_start", Group: true},
				{Type: "dateTime:RFC3339", Default: nil, Name: "_stop", Group: true},
				{Type: "dateTime:RFC3339", Default: nil, Name: "_time", Group: false},
				{Type: "boolean", Default: nil, Name: "_value", Group: false},
				{Type: "string", Default: nil, Name: "_field", Group: true},
				{Type: "string", Default: nil, Name: "_measurement", Group: true},
				{Type: "string", Default: nil, Name: "a", Group: true},
				{Type: "string", Default: nil, Name: "b", Group: true},
			},
			[][]interface{}{
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
			[]annotatedcsv.Column{
				{Type: "string", Default: "_result", Name: "result", Group: false},
				{Type: "long", Default: nil, Name: "table", Group: false},
				{Type: "dateTime:RFC3339Nano", Default: nil, Name: "_start", Group: true},
				{Type: "dateTime:RFC3339Nano", Default: nil, Name: "_stop", Group: true},
				{Type: "dateTime:RFC3339Nano", Default: nil, Name: "_time", Group: false},
				{Type: "unsignedLong", Default: nil, Name: "_value", Group: false},
				{Type: "string", Default: nil, Name: "_field", Group: true},
				{Type: "string", Default: nil, Name: "_measurement", Group: true},
				{Type: "string", Default: nil, Name: "a", Group: true},
				{Type: "string", Default: nil, Name: "b", Group: true},
			},
			[][]interface{}{
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

	reader := strings.NewReader(csvTableMultiStructure)
	res := annotatedcsv.NewReader(ioutil.NopCloser(reader))

	//test skip first table header
	require.True(t, res.NextRow())
	require.Nil(t, res.Err())
	require.Equal(t, tables[0].rows[0], res.Row())
	_ = res.Close()

	reader = strings.NewReader(csvTableMultiStructure)
	res = annotatedcsv.NewReader(ioutil.NopCloser(reader))

	//test skip tables
	require.True(t, res.NextSection())
	require.Nil(t, res.Err())
	require.True(t, res.NextSection())
	require.Nil(t, res.Err())
	require.True(t, res.NextRow())
	require.Nil(t, res.Err())
	require.True(t, res.NextSection())
	require.Nil(t, res.Err())
	require.True(t, res.NextRow())
	require.Nil(t, res.Err())
	require.Equal(t, tables[2].rows[0], res.Row())

	_ = res.Close()

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
	res = annotatedcsv.NewReader(ioutil.NopCloser(reader))

	//test skip first table header
	require.True(t, res.NextRow())
	require.True(t, res.NextRow())
	require.Nil(t, res.Err())
	require.Equal(t, tables[0].rows[1], res.Row())
	_ = res.Close()

	reader = strings.NewReader(csvTableMultiTables)
	res = annotatedcsv.NewReader(ioutil.NopCloser(reader))
	require.True(t, res.NextSection())
	require.Nil(t, res.Err())
	require.False(t, res.NextSection())
	require.Nil(t, res.Err())
}

func TestValueByName(t *testing.T) {
	csvTable := `#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,duration,base64Binary,dateTime:RFC3339
#group,false,false,true,true,false,true,true,false,false,false
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,deviceId,sensor,elapsed,note,start
,,0,2020-04-28T12:36:50.990018157Z,2020-04-28T12:51:50.990018157Z,2020-04-28T12:38:11.480545389Z,1467463,BME280,1m1s,ZGF0YWluYmFzZTY0,2020-04-27T00:00:00Z
,,1,2020-04-28T12:36:50.990018157Z,2020-04-28T12:51:50.990018157Z,2020-04-28T12:39:36.330153686Z,1467463,BME280,1h20m30.13245s,eHh4eHhjY2NjY2NkZGRkZA==,2020-04-28T00:00:00Z

#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#group,false,false,true,true,false,false,true,true,true,true
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,a,b
,,0,2020-02-17T22:19:49.747562847Z,2020-02-18T22:19:49.747562847Z,2020-02-18T10:34:08.135814545Z,1.4,f,test,1,adsfasdf
,,0,2020-02-17T22:19:49.747562847Z,2020-02-18T22:19:49.747562847Z,2020-02-18T22:08:44.850214724Z,6.6,f,test,1,adsfasdf

`
	tables := []expectedTable{
		{ // Table 1
			[]annotatedcsv.Column{
				{Type: "string", Default: "_result", Name: "result", Group: false},
				{Type: "long", Default: nil, Name: "table", Group: false},
				{Type: "dateTime:RFC3339", Default: nil, Name: "_start", Group: true},
				{Type: "dateTime:RFC3339", Default: nil, Name: "_stop", Group: true},
				{Type: "dateTime:RFC3339", Default: nil, Name: "_time", Group: false},
				{Type: "long", Default: nil, Name: "deviceId", Group: true},
				{Type: "string", Default: nil, Name: "sensor", Group: true},
				{Type: "duration", Default: nil, Name: "elapsed", Group: false},
				{Type: "base64Binary", Default: nil, Name: "note", Group: false},
				{Type: "dateTime:RFC3339", Default: nil, Name: "start", Group: false},
			},
			[][]interface{}{
				{
					"_result",
					int64(0),
					mustParseTime("2020-04-28T12:36:50.990018157Z"),
					mustParseTime("2020-04-28T12:51:50.990018157Z"),
					mustParseTime("2020-04-28T12:38:11.480545389Z"),
					int64(1467463),
					"BME280",
					time.Minute + time.Second,
					[]byte("datainbase64"),
					time.Date(2020, 4, 27, 0, 0, 0, 0, time.UTC),
				},
				{
					"_result",
					int64(1),
					mustParseTime("2020-04-28T12:36:50.990018157Z"),
					mustParseTime("2020-04-28T12:51:50.990018157Z"),
					mustParseTime("2020-04-28T12:39:36.330153686Z"),
					int64(1467463),
					"BME280",
					time.Hour + 20*time.Minute + 30*time.Second + 132450000*time.Nanosecond,
					[]byte("xxxxxccccccddddd"),
					time.Date(2020, 4, 28, 0, 0, 0, 0, time.UTC),
				},
			},
		},
		{[]annotatedcsv.Column{
			{Type: "string", Default: "_result", Name: "result", Group: false},
			{Type: "long", Default: nil, Name: "table", Group: false},
			{Type: "dateTime:RFC3339", Default: nil, Name: "_start", Group: true},
			{Type: "dateTime:RFC3339", Default: nil, Name: "_stop", Group: true},
			{Type: "dateTime:RFC3339", Default: nil, Name: "_time", Group: false},
			{Type: "double", Default: nil, Name: "_value", Group: false},
			{Type: "string", Default: nil, Name: "_field", Group: true},
			{Type: "string", Default: nil, Name: "_measurement", Group: true},
			{Type: "string", Default: nil, Name: "a", Group: true},
			{Type: "string", Default: nil, Name: "b", Group: true},
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

	reader := strings.NewReader(csvTable)
	res := annotatedcsv.NewReader(ioutil.NopCloser(reader))

	require.True(t, res.NextSection() && res.NextRow(), res.Err())
	require.Nil(t, res.Err())
	assert.Equal(t, []byte("datainbase64"), res.ValueByName("note"))
	assert.Nil(t, res.ValueByName(""))
	assert.Nil(t, res.ValueByName("invalid"))
	assert.Nil(t, res.ValueByName("a"))

	require.True(t, res.NextSection() && res.NextRow(), res.Err())
	assert.Equal(t, "1", res.ValueByName("a"))
	assert.Nil(t, res.ValueByName("note"))

}

func TestErrorInRow(t *testing.T) {
	csvTableError := `#datatype,string,string
#group,true,true
#default,,
,error,reference
,failed to create physical plan: invalid time bounds from procedure from: bounds contain zero time,897`

	verifyParsingError(t, csvTableError, "failed to create physical plan: invalid time bounds from procedure from: bounds contain zero time,897", true)

	csvTableErrorNoReference := `#datatype,string,string
#group,true,true
#default,,
,error,reference
,failed to create physical plan: invalid time bounds from procedure from: bounds contain zero time,`
	verifyParsingError(t, csvTableErrorNoReference, "failed to create physical plan: invalid time bounds from procedure from: bounds contain zero time", true)

	csvTableErrorNoMessage := `#datatype,string,string
#group,true,true
#default,,
,error,reference
,,`
	verifyParsingError(t, csvTableErrorNoMessage, "unknown query error", true)
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
			{Type: "long", Default: nil, Name: "table", Group: false},
			{Type: "dateTime:RFC3339", Default: nil, Name: "_start", Group: true},
			{Type: "dateTime:RFC3339", Default: nil, Name: "_stop", Group: true},
			{Type: "dateTime:RFC3339", Default: nil, Name: "_time", Group: false},
			{Type: "int", Default: nil, Name: "_value", Group: false},
			{Type: "string", Default: nil, Name: "_field", Group: true},
			{Type: "string", Default: nil, Name: "_measurement", Group: true},
			{Type: "string", Default: nil, Name: "a", Group: true},
			{Type: "string", Default: nil, Name: "b", Group: true},
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
			{Type: "long", Default: nil, Name: "table", Group: false},
			{Type: "dateTime:RFC3339", Default: nil, Name: "_start", Group: true},
			{Type: "dateTime:RFC3339", Default: nil, Name: "_stop", Group: true},
			{Type: "dateTime:RFC3339", Default: nil, Name: "_time", Group: false},
			{Type: "double", Default: nil, Name: "_value", Group: false},
			{Type: "string", Default: nil, Name: "_field", Group: true},
			{Type: "string", Default: nil, Name: "_measurement", Group: true},
			{Type: "string", Default: nil, Name: "a", Group: true},
			{Type: "string", Default: nil, Name: "b", Group: true},
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
			{Type: "string", Default: nil, Name: "result", Group: false},
			{Type: "long", Default: nil, Name: "table", Group: false},
			{Type: "dateTime:RFC3339", Default: nil, Name: "_start", Group: false},
			{Type: "dateTime:RFC3339", Default: nil, Name: "_stop", Group: false},
			{Type: "dateTime:RFC3339", Default: nil, Name: "_time", Group: false},
			{Type: "double", Default: nil, Name: "_value", Group: false},
			{Type: "string", Default: nil, Name: "_field", Group: false},
			{Type: "string", Default: nil, Name: "_measurement", Group: false},
			{Type: "string", Default: nil, Name: "a", Group: false},
			{Type: "string", Default: nil, Name: "b", Group: false},
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

	verifyParsingError(t, csvTable1, "datatype annotation not found", true)

	csvTable2 := `
#default,_result,,,,,,,,,
#group,false,false,true,true,false,true,true,false,false,false
,result,table,_start,_stop,_time,deviceId,sensor,elapsed,note,start
,,0,2020-04-28T12:36:50.990018157Z,2020-04-28T12:51:50.990018157Z,2020-04-28T12:38:11.480545389Z,1467463,BME280,1m1s,ZGF0YWluYmFzZTY0,2020-04-27T00:00:00Z
,,0,2020-04-28T12:36:50.990018157Z,2020-04-28T12:51:50.990018157Z,2020-04-28T12:39:36.330153686Z,1467463,BME280,1h20m30.13245s,eHh4eHhjY2NjY2NkZGRkZA==,2020-04-28T00:00:00Z
`
	verifyParsingError(t, csvTable2, "datatype annotation not found", true)
}

func TestMissingAnnotations(t *testing.T) {
	csvTable := `
,result,table,_start,_stop,_time,deviceId,sensor,elapsed,note,start
,,0,2020-04-28T12:36:50.990018157Z,2020-04-28T12:51:50.990018157Z,2020-04-28T12:38:11.480545389Z,1467463,BME280,1m1s,ZGF0YWluYmFzZTY0,2020-04-27T00:00:00Z
,,0,2020-04-28T12:36:50.990018157Z,2020-04-28T12:51:50.990018157Z,2020-04-28T12:39:36.330153686Z,1467463,BME280,1h20m30.13245s,eHh4eHhjY2NjY2NkZGRkZA==,2020-04-28T00:00:00Z

`
	verifyParsingError(t, csvTable, "datatype annotation not found", true)
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
			{Type: "long", Default: nil, Name: "table", Group: false},
			{Type: "dateTime:RFC3339", Default: nil, Name: "_start", Group: true},
			{Type: "dateTime:RFC3339", Default: nil, Name: "_stop", Group: true},
			{Type: "dateTime:RFC3339", Default: nil, Name: "_time", Group: false},
			{Type: "double", Default: nil, Name: "_value", Group: false},
			{Type: "string", Default: nil, Name: "_field", Group: true},
			{Type: "string", Default: nil, Name: "_measurement", Group: true},
			{Type: "string", Default: nil, Name: "a", Group: true},
			{Type: "string", Default: nil, Name: "b", Group: true},
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
			{Type: "long", Default: nil, Name: "table", Group: false},
			{Type: "dateTime:RFC3339", Default: nil, Name: "_start", Group: true},
			{Type: "dateTime", Default: nil, Name: "_stop", Group: true},
			{Type: "dateTime:RFC3339Nano", Default: nil, Name: "_time", Group: false},
			{Type: "double", Default: nil, Name: "_value", Group: false},
			{Type: "string", Default: nil, Name: "_field", Group: true},
			{Type: "string", Default: nil, Name: "_measurement", Group: true},
			{Type: "string", Default: nil, Name: "a", Group: true},
			{Type: "string", Default: nil, Name: "b", Group: true},
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
	verifyParsingError(t, csvTable, `cannot convert value "2020-02-17T22:19:49.747562847Z" to type "dateTime:wrongLayout" at line 5: unknown time format "dateTime:wrongLayout"`, false)

}

func TestFailedConversion(t *testing.T) {
	csvTable := `#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#group,false,false,true,true,false,false,true,true,true,true
#default,_result,zero,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,a,b
,,0,2020-02-17T22:19:49.747562847Z,2020-02-18T22:19:49.747562847Z,2020-02-18T10:34:08.135814545Z,seven,f,test,1,adsfasdf
,,0,2020-02-17T22:19:49.747562847Z,2020-02-18T22:19:49.747562847Z,2020-02-18T22:08:44.850214724Z,six,f,test,1,adsfasdf

`
	verifyParsingError(t, csvTable, `cannot convert default value "zero" to type "long": strconv.ParseInt: parsing "zero": invalid syntax`, true)

	csvTable = `#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#group,false,false,true,true,false,false,true,true,true,true
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,a,b
,,0,2020-02-17T22:19:49.747562847Z,2020-02-18T22:19:49.747562847Z,2020-02-18T10:34:08.135814545Z,seven,f,test,1,adsfasdf
,,0,2020-02-17T22:19:49.747562847Z,2020-02-18T22:19:49.747562847Z,2020-02-18T22:08:44.850214724Z,six,f,test,1,adsfasdf

`
	verifyParsingError(t, csvTable, `cannot convert value "seven" to type "double" at line 5: strconv.ParseFloat: parsing "seven": invalid syntax`, false)
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
	res := annotatedcsv.NewReader(ioutil.NopCloser(reader))

	require.False(t, res.NextRow())
	require.NotNil(t, res.Err())

	csvErrTable = `#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
,result,table,_start,_stop,_time,_value,_field,_measurement,a,b
,",0,2020-02-17T22:19:49.747562847Z,2020-02-18T22:19:49.747562847Z,2020-02-18T10:34:08.135814545Z,1.4,f,test,1,adsfasdf
,,0,2020-02-17T22:19:49.747562847Z,2020-02-18T22:19:49.747562847Z,2020-02-18T22:08:44.850214724Z,6.6,f,test,1,adsfasdf
`
	reader = strings.NewReader(csvErrTable)
	res = annotatedcsv.NewReader(ioutil.NopCloser(reader))

	require.False(t, res.NextRow())
	require.NotNil(t, res.Err())
}

type errCloser struct {
	io.Reader
}

func (errCloser) Close() error {
	return errors.New("close error")
}

func newErrCloser(r io.Reader) io.ReadCloser {
	return errCloser{r}
}

func TestCloseError(t *testing.T) {
	csvTable := `#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
,result,table,_start,_stop,_time,_value,_field,_measurement,a,b
,,0,2020-02-17T22:19:49.747562847Z,2020-02-18T22:19:49.747562847Z,2020-02-18T10:34:08.135814545Z,1.4,f,test,1,adsfasdf
,,0,2020-02-17T22:19:49.747562847Z,2020-02-18T22:19:49.747562847Z,2020-02-18T22:08:44.850214724Z,6.6,f,test,1,adsfasdf
`
	csvErrTable := `#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
,result,table,_start,_stop,_time,_value,_field,_measurement,a,b
,",0,2020-02-17T22:19:49.747562847Z,2020-02-18T22:19:49.747562847Z,2020-02-18T10:34:08.135814545Z,1.4,f,test,1,adsfasdf
,,0,2020-02-17T22:19:49.747562847Z,2020-02-18T22:19:49.747562847Z,2020-02-18T22:08:44.850214724Z,6.6,f,test,1,adsfasdf
`
	reader := strings.NewReader(csvTable)
	res := annotatedcsv.NewReader(newErrCloser(reader))

	require.True(t, res.NextSection())
	require.True(t, res.NextRow())
	require.True(t, res.NextRow())
	require.False(t, res.NextRow())
	require.Nil(t, res.Err())

	reader = strings.NewReader(csvErrTable)
	res = annotatedcsv.NewReader(newErrCloser(reader))

	require.False(t, res.NextRow())
	require.NotNil(t, res.Err())
}

func mustParseTime(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return t
}

func verifyTables(t *testing.T, csv string, tables []expectedTable) {
	reader := strings.NewReader(csv)
	res := annotatedcsv.NewReader(ioutil.NopCloser(reader))

	for _, table := range tables {
		require.True(t, res.NextSection(), res.Err())
		require.Nil(t, res.Err())
		require.Equal(t, table.columns, res.Columns())
		for _, row := range table.rows {
			require.True(t, res.NextRow(), res.Err())
			require.Nil(t, res.Err())
			for i, v := range res.Row() {
				if table.columns[i].Type == "base64Binary" {
					require.Equal(t, row[i], v)
				} else {
					require.True(t, row[i] == v, fmt.Sprintf("%v vs %v", row[i], v))
				}
			}
			for i, c := range table.columns {
				v := res.ValueByName(c.Name)
				if c.Type == "base64Binary" {
					require.Equal(t, row[i], v)
				} else {
					require.True(t, row[i] == v)
				}
			}
		}
		require.False(t, res.NextRow(), res.Err())
		require.Nil(t, res.Err())
	}

	require.False(t, res.NextSection(), res.Err())
	require.Nil(t, res.Err())

	require.Nil(t, res.Columns())
	require.Nil(t, res.Row())
	require.Nil(t, res.ValueByName("table"))
}

func verifyParsingError(t *testing.T, csvTable, error string, inHeader bool) {
	reader := strings.NewReader(csvTable)
	res := annotatedcsv.NewReader(ioutil.NopCloser(reader))

	if inHeader {
		require.False(t, res.NextSection())
	} else {
		require.True(t, res.NextSection())
	}
	require.False(t, res.NextRow())
	require.NotNil(t, res.Err())
	assert.Equal(t, error, res.Err().Error())

}
