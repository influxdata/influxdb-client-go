// Copyright 2020-2021 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package api

import (
	"context"
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	http2 "github.com/influxdata/influxdb-client-go/v2/api/http"
	"github.com/influxdata/influxdb-client-go/v2/api/query"
	"github.com/influxdata/influxdb-client-go/v2/internal/gzip"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mustParseTime(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return t
}

func TestQueryCVSResultSingleTable(t *testing.T) {
	csvTable := `#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#group,false,false,true,true,false,false,true,true,true,true
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,a,b
,,0,2020-02-17T22:19:49.747562847Z,2020-02-18T22:19:49.747562847Z,2020-02-18T10:34:08.135814545Z,1.4,f,test,1,adsfasdf
,,0,2020-02-17T22:19:49.747562847Z,2020-02-18T22:19:49.747562847Z,2020-02-18T22:08:44.850214724Z,6.6,f,test,1,adsfasdf

`
	expectedTable := query.NewFluxTableMetadataFull(0,
		[]*query.FluxColumn{
			query.NewFluxColumnFull("string", "_result", "result", false, 0),
			query.NewFluxColumnFull("long", "", "table", false, 1),
			query.NewFluxColumnFull("dateTime:RFC3339", "", "_start", true, 2),
			query.NewFluxColumnFull("dateTime:RFC3339", "", "_stop", true, 3),
			query.NewFluxColumnFull("dateTime:RFC3339", "", "_time", false, 4),
			query.NewFluxColumnFull("double", "", "_value", false, 5),
			query.NewFluxColumnFull("string", "", "_field", true, 6),
			query.NewFluxColumnFull("string", "", "_measurement", true, 7),
			query.NewFluxColumnFull("string", "", "a", true, 8),
			query.NewFluxColumnFull("string", "", "b", true, 9),
		},
	)
	expectedRecord1 := query.NewFluxRecord(0,
		map[string]interface{}{
			"result":       "_result",
			"table":        int64(0),
			"_start":       mustParseTime("2020-02-17T22:19:49.747562847Z"),
			"_stop":        mustParseTime("2020-02-18T22:19:49.747562847Z"),
			"_time":        mustParseTime("2020-02-18T10:34:08.135814545Z"),
			"_value":       1.4,
			"_field":       "f",
			"_measurement": "test",
			"a":            "1",
			"b":            "adsfasdf",
		},
	)

	expectedRecord2 := query.NewFluxRecord(0,
		map[string]interface{}{
			"result":       "_result",
			"table":        int64(0),
			"_start":       mustParseTime("2020-02-17T22:19:49.747562847Z"),
			"_stop":        mustParseTime("2020-02-18T22:19:49.747562847Z"),
			"_time":        mustParseTime("2020-02-18T22:08:44.850214724Z"),
			"_value":       6.6,
			"_field":       "f",
			"_measurement": "test",
			"a":            "1",
			"b":            "adsfasdf",
		},
	)

	reader := strings.NewReader(csvTable)
	csvReader := csv.NewReader(reader)
	csvReader.FieldsPerRecord = -1
	queryResult := &QueryTableResult{Closer: ioutil.NopCloser(reader), csvReader: csvReader}
	require.True(t, queryResult.Next(), queryResult.Err())
	require.Nil(t, queryResult.Err())

	require.Equal(t, queryResult.table, expectedTable)
	assert.True(t, queryResult.tableChanged)
	require.NotNil(t, queryResult.Record())
	require.Equal(t, queryResult.Record(), expectedRecord1)

	require.True(t, queryResult.Next(), queryResult.Err())
	require.Nil(t, queryResult.Err())
	assert.False(t, queryResult.tableChanged)
	require.NotNil(t, queryResult.Record())
	require.Equal(t, queryResult.Record(), expectedRecord2)

	require.False(t, queryResult.Next())
	require.Nil(t, queryResult.Err())
}

func TestQueryCVSResultMultiTables(t *testing.T) {
	csvTable := `#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#group,false,false,true,true,false,false,true,true,true,true
#default,_result1,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,a,b
,,0,2020-02-17T22:19:49.747562847Z,2020-02-18T22:19:49.747562847Z,2020-02-18T10:34:08.135814545Z,1.4,f,test,1,adsfasdf
,,0,2020-02-17T22:19:49.747562847Z,2020-02-18T22:19:49.747562847Z,2020-02-18T22:08:44.850214724Z,6.6,f,test,1,adsfasdf

#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#group,false,false,true,true,false,false,true,true,true,true
#default,_result2,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,a,b
,,1,2020-02-17T22:19:49.747562847Z,2020-02-18T22:19:49.747562847Z,2020-02-18T10:34:08.135814545Z,4,i,test,1,adsfasdf
,,1,2020-02-17T22:19:49.747562847Z,2020-02-18T22:19:49.747562847Z,2020-02-18T22:08:44.850214724Z,-1,i,test,1,adsfasdf

#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,boolean,string,string,string,string
#group,false,false,true,true,false,false,true,true,true,true
#default,_result3,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,a,b
,,2,2020-02-17T22:19:49.747562847Z,2020-02-18T22:19:49.747562847Z,2020-02-18T22:08:44.62797864Z,false,f,test,0,adsfasdf
,,2,2020-02-17T22:19:49.747562847Z,2020-02-18T22:19:49.747562847Z,2020-02-18T22:08:44.969100374Z,true,f,test,0,adsfasdf

#datatype,string,long,dateTime:RFC3339Nano,dateTime:RFC3339Nano,dateTime:RFC3339Nano,unsignedLong,string,string,string,string
#group,false,false,true,true,false,false,true,true,true,true
#default,_result4,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,a,b
,,3,2020-02-17T22:19:49.747562847Z,2020-02-18T22:19:49.747562847Z,2020-02-18T22:08:44.62797864Z,0,i,test,0,adsfasdf
,,3,2020-02-17T22:19:49.747562847Z,2020-02-18T22:19:49.747562847Z,2020-02-18T22:08:44.969100374Z,2,i,test,0,adsfasdf

`
	expectedTable1 := query.NewFluxTableMetadataFull(0,
		[]*query.FluxColumn{
			query.NewFluxColumnFull("string", "_result1", "result", false, 0),
			query.NewFluxColumnFull("long", "", "table", false, 1),
			query.NewFluxColumnFull("dateTime:RFC3339", "", "_start", true, 2),
			query.NewFluxColumnFull("dateTime:RFC3339", "", "_stop", true, 3),
			query.NewFluxColumnFull("dateTime:RFC3339", "", "_time", false, 4),
			query.NewFluxColumnFull("double", "", "_value", false, 5),
			query.NewFluxColumnFull("string", "", "_field", true, 6),
			query.NewFluxColumnFull("string", "", "_measurement", true, 7),
			query.NewFluxColumnFull("string", "", "a", true, 8),
			query.NewFluxColumnFull("string", "", "b", true, 9),
		},
	)
	expectedRecord11 := query.NewFluxRecord(0,
		map[string]interface{}{
			"result":       "_result1",
			"table":        int64(0),
			"_start":       mustParseTime("2020-02-17T22:19:49.747562847Z"),
			"_stop":        mustParseTime("2020-02-18T22:19:49.747562847Z"),
			"_time":        mustParseTime("2020-02-18T10:34:08.135814545Z"),
			"_value":       1.4,
			"_field":       "f",
			"_measurement": "test",
			"a":            "1",
			"b":            "adsfasdf",
		},
	)
	expectedRecord12 := query.NewFluxRecord(0,
		map[string]interface{}{
			"result":       "_result1",
			"table":        int64(0),
			"_start":       mustParseTime("2020-02-17T22:19:49.747562847Z"),
			"_stop":        mustParseTime("2020-02-18T22:19:49.747562847Z"),
			"_time":        mustParseTime("2020-02-18T22:08:44.850214724Z"),
			"_value":       6.6,
			"_field":       "f",
			"_measurement": "test",
			"a":            "1",
			"b":            "adsfasdf",
		},
	)

	expectedTable2 := query.NewFluxTableMetadataFull(1,
		[]*query.FluxColumn{
			query.NewFluxColumnFull("string", "_result2", "result", false, 0),
			query.NewFluxColumnFull("long", "", "table", false, 1),
			query.NewFluxColumnFull("dateTime:RFC3339", "", "_start", true, 2),
			query.NewFluxColumnFull("dateTime:RFC3339", "", "_stop", true, 3),
			query.NewFluxColumnFull("dateTime:RFC3339", "", "_time", false, 4),
			query.NewFluxColumnFull("long", "", "_value", false, 5),
			query.NewFluxColumnFull("string", "", "_field", true, 6),
			query.NewFluxColumnFull("string", "", "_measurement", true, 7),
			query.NewFluxColumnFull("string", "", "a", true, 8),
			query.NewFluxColumnFull("string", "", "b", true, 9),
		},
	)
	expectedRecord21 := query.NewFluxRecord(1,
		map[string]interface{}{
			"result":       "_result2",
			"table":        int64(1),
			"_start":       mustParseTime("2020-02-17T22:19:49.747562847Z"),
			"_stop":        mustParseTime("2020-02-18T22:19:49.747562847Z"),
			"_time":        mustParseTime("2020-02-18T10:34:08.135814545Z"),
			"_value":       int64(4),
			"_field":       "i",
			"_measurement": "test",
			"a":            "1",
			"b":            "adsfasdf",
		},
	)
	expectedRecord22 := query.NewFluxRecord(1,
		map[string]interface{}{
			"result":       "_result2",
			"table":        int64(1),
			"_start":       mustParseTime("2020-02-17T22:19:49.747562847Z"),
			"_stop":        mustParseTime("2020-02-18T22:19:49.747562847Z"),
			"_time":        mustParseTime("2020-02-18T22:08:44.850214724Z"),
			"_value":       int64(-1),
			"_field":       "i",
			"_measurement": "test",
			"a":            "1",
			"b":            "adsfasdf",
		},
	)

	expectedTable3 := query.NewFluxTableMetadataFull(2,
		[]*query.FluxColumn{
			query.NewFluxColumnFull("string", "_result3", "result", false, 0),
			query.NewFluxColumnFull("long", "", "table", false, 1),
			query.NewFluxColumnFull("dateTime:RFC3339", "", "_start", true, 2),
			query.NewFluxColumnFull("dateTime:RFC3339", "", "_stop", true, 3),
			query.NewFluxColumnFull("dateTime:RFC3339", "", "_time", false, 4),
			query.NewFluxColumnFull("boolean", "", "_value", false, 5),
			query.NewFluxColumnFull("string", "", "_field", true, 6),
			query.NewFluxColumnFull("string", "", "_measurement", true, 7),
			query.NewFluxColumnFull("string", "", "a", true, 8),
			query.NewFluxColumnFull("string", "", "b", true, 9),
		},
	)
	expectedRecord31 := query.NewFluxRecord(2,
		map[string]interface{}{
			"result":       "_result3",
			"table":        int64(2),
			"_start":       mustParseTime("2020-02-17T22:19:49.747562847Z"),
			"_stop":        mustParseTime("2020-02-18T22:19:49.747562847Z"),
			"_time":        mustParseTime("2020-02-18T22:08:44.62797864Z"),
			"_value":       false,
			"_field":       "f",
			"_measurement": "test",
			"a":            "0",
			"b":            "adsfasdf",
		},
	)
	expectedRecord32 := query.NewFluxRecord(2,
		map[string]interface{}{
			"result":       "_result3",
			"table":        int64(2),
			"_start":       mustParseTime("2020-02-17T22:19:49.747562847Z"),
			"_stop":        mustParseTime("2020-02-18T22:19:49.747562847Z"),
			"_time":        mustParseTime("2020-02-18T22:08:44.969100374Z"),
			"_value":       true,
			"_field":       "f",
			"_measurement": "test",
			"a":            "0",
			"b":            "adsfasdf",
		},
	)

	expectedTable4 := query.NewFluxTableMetadataFull(3,
		[]*query.FluxColumn{
			query.NewFluxColumnFull("string", "_result4", "result", false, 0),
			query.NewFluxColumnFull("long", "", "table", false, 1),
			query.NewFluxColumnFull("dateTime:RFC3339Nano", "", "_start", true, 2),
			query.NewFluxColumnFull("dateTime:RFC3339Nano", "", "_stop", true, 3),
			query.NewFluxColumnFull("dateTime:RFC3339Nano", "", "_time", false, 4),
			query.NewFluxColumnFull("unsignedLong", "", "_value", false, 5),
			query.NewFluxColumnFull("string", "", "_field", true, 6),
			query.NewFluxColumnFull("string", "", "_measurement", true, 7),
			query.NewFluxColumnFull("string", "", "a", true, 8),
			query.NewFluxColumnFull("string", "", "b", true, 9),
		},
	)
	expectedRecord41 := query.NewFluxRecord(3,
		map[string]interface{}{
			"result":       "_result4",
			"table":        int64(3),
			"_start":       mustParseTime("2020-02-17T22:19:49.747562847Z"),
			"_stop":        mustParseTime("2020-02-18T22:19:49.747562847Z"),
			"_time":        mustParseTime("2020-02-18T22:08:44.62797864Z"),
			"_value":       uint64(0),
			"_field":       "i",
			"_measurement": "test",
			"a":            "0",
			"b":            "adsfasdf",
		},
	)
	expectedRecord42 := query.NewFluxRecord(3,
		map[string]interface{}{
			"result":       "_result4",
			"table":        int64(3),
			"_start":       mustParseTime("2020-02-17T22:19:49.747562847Z"),
			"_stop":        mustParseTime("2020-02-18T22:19:49.747562847Z"),
			"_time":        mustParseTime("2020-02-18T22:08:44.969100374Z"),
			"_value":       uint64(2),
			"_field":       "i",
			"_measurement": "test",
			"a":            "0",
			"b":            "adsfasdf",
		},
	)

	reader := strings.NewReader(csvTable)
	csvReader := csv.NewReader(reader)
	csvReader.FieldsPerRecord = -1
	queryResult := &QueryTableResult{Closer: ioutil.NopCloser(reader), csvReader: csvReader}
	assert.Equal(t, -1, queryResult.TablePosition())
	require.True(t, queryResult.Next(), queryResult.Err())
	require.Nil(t, queryResult.Err())

	require.Equal(t, queryResult.TableMetadata(), expectedTable1)
	require.NotNil(t, queryResult.Record())
	require.Equal(t, queryResult.Record(), expectedRecord11)
	assert.True(t, queryResult.tableChanged)
	assert.Equal(t, 0, queryResult.TablePosition())

	require.True(t, queryResult.Next(), queryResult.Err())
	require.Nil(t, queryResult.Err())
	require.Equal(t, queryResult.TableMetadata(), expectedTable1)
	assert.False(t, queryResult.tableChanged)
	require.NotNil(t, queryResult.Record())
	require.Equal(t, queryResult.Record(), expectedRecord12)
	assert.Equal(t, 0, queryResult.TablePosition())

	require.True(t, queryResult.Next(), queryResult.Err())
	require.Nil(t, queryResult.Err())

	assert.True(t, queryResult.tableChanged)
	assert.Equal(t, 1, queryResult.TablePosition())
	require.Equal(t, queryResult.table, expectedTable2)
	require.NotNil(t, queryResult.Record())
	require.Equal(t, queryResult.Record(), expectedRecord21)

	require.True(t, queryResult.Next(), queryResult.Err())
	require.Nil(t, queryResult.Err())

	assert.False(t, queryResult.tableChanged)
	assert.Equal(t, 1, queryResult.TablePosition())
	require.Equal(t, queryResult.table, expectedTable2)
	require.NotNil(t, queryResult.Record())
	require.Equal(t, queryResult.Record(), expectedRecord22)

	require.True(t, queryResult.Next(), queryResult.Err())
	require.Nil(t, queryResult.Err(), queryResult.Err())

	assert.True(t, queryResult.tableChanged)
	assert.Equal(t, 2, queryResult.TablePosition())
	require.Equal(t, queryResult.table, expectedTable3)
	require.NotNil(t, queryResult.Record())
	require.Equal(t, queryResult.Record(), expectedRecord31)

	require.True(t, queryResult.Next(), queryResult.Err())
	require.Nil(t, queryResult.Err())

	assert.False(t, queryResult.tableChanged)
	assert.Equal(t, 2, queryResult.TablePosition())
	require.Equal(t, queryResult.table, expectedTable3)
	require.NotNil(t, queryResult.Record())
	require.Equal(t, queryResult.Record(), expectedRecord32)

	require.True(t, queryResult.Next(), queryResult.Err())
	require.Nil(t, queryResult.Err())

	assert.True(t, queryResult.tableChanged)
	assert.Equal(t, 3, queryResult.TablePosition())
	require.Equal(t, queryResult.table, expectedTable4)
	require.NotNil(t, queryResult.Record())
	require.Equal(t, queryResult.Record(), expectedRecord41)

	require.True(t, queryResult.Next(), queryResult.Err())
	require.Nil(t, queryResult.Err())

	assert.False(t, queryResult.tableChanged)
	assert.Equal(t, 3, queryResult.TablePosition())
	require.Equal(t, queryResult.table, expectedTable4)
	require.NotNil(t, queryResult.Record())
	require.Equal(t, queryResult.Record(), expectedRecord42)
	assert.Equal(t, "_result4", queryResult.Record().Result())

	require.False(t, queryResult.Next())
	require.Nil(t, queryResult.Err())
}

func TestQueryCVSResultSingleTableMultiColumnsNoValue(t *testing.T) {
	csvTable := `#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,duration,base64Binary,dateTime:RFC3339
#group,false,false,true,true,false,true,true,false,false,false
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,deviceId,sensor,elapsed,note,start
,,0,2020-04-28T12:36:50.990018157Z,2020-04-28T12:51:50.990018157Z,2020-04-28T12:38:11.480545389Z,1467463,BME280,1m1s,ZGF0YWluYmFzZTY0,2020-04-27T00:00:00Z
,,1,2020-04-28T12:36:50.990018157Z,2020-04-28T12:51:50.990018157Z,2020-04-28T12:39:36.330153686Z,1467463,BME280,1h20m30.13245s,eHh4eHhjY2NjY2NkZGRkZA==,2020-04-28T00:00:00Z
`
	expectedTable := query.NewFluxTableMetadataFull(0,
		[]*query.FluxColumn{
			query.NewFluxColumnFull("string", "_result", "result", false, 0),
			query.NewFluxColumnFull("long", "", "table", false, 1),
			query.NewFluxColumnFull("dateTime:RFC3339", "", "_start", true, 2),
			query.NewFluxColumnFull("dateTime:RFC3339", "", "_stop", true, 3),
			query.NewFluxColumnFull("dateTime:RFC3339", "", "_time", false, 4),
			query.NewFluxColumnFull("long", "", "deviceId", true, 5),
			query.NewFluxColumnFull("string", "", "sensor", true, 6),
			query.NewFluxColumnFull("duration", "", "elapsed", false, 7),
			query.NewFluxColumnFull("base64Binary", "", "note", false, 8),
			query.NewFluxColumnFull("dateTime:RFC3339", "", "start", false, 9),
		},
	)
	expectedRecord1 := query.NewFluxRecord(0,
		map[string]interface{}{
			"result":   "_result",
			"table":    int64(0),
			"_start":   mustParseTime("2020-04-28T12:36:50.990018157Z"),
			"_stop":    mustParseTime("2020-04-28T12:51:50.990018157Z"),
			"_time":    mustParseTime("2020-04-28T12:38:11.480545389Z"),
			"deviceId": int64(1467463),
			"sensor":   "BME280",
			"elapsed":  time.Minute + time.Second,
			"note":     []byte("datainbase64"),
			"start":    time.Date(2020, 4, 27, 0, 0, 0, 0, time.UTC),
		},
	)

	expectedRecord2 := query.NewFluxRecord(0,
		map[string]interface{}{
			"result":   "_result",
			"table":    int64(1),
			"_start":   mustParseTime("2020-04-28T12:36:50.990018157Z"),
			"_stop":    mustParseTime("2020-04-28T12:51:50.990018157Z"),
			"_time":    mustParseTime("2020-04-28T12:39:36.330153686Z"),
			"deviceId": int64(1467463),
			"sensor":   "BME280",
			"elapsed":  time.Hour + 20*time.Minute + 30*time.Second + 132450000*time.Nanosecond,
			"note":     []byte("xxxxxccccccddddd"),
			"start":    time.Date(2020, 4, 28, 0, 0, 0, 0, time.UTC),
		},
	)

	reader := strings.NewReader(csvTable)
	csvReader := csv.NewReader(reader)
	csvReader.FieldsPerRecord = -1
	queryResult := &QueryTableResult{Closer: ioutil.NopCloser(reader), csvReader: csvReader}
	require.True(t, queryResult.Next(), queryResult.Err())
	require.Nil(t, queryResult.Err())

	require.Equal(t, queryResult.table, expectedTable)
	assert.True(t, queryResult.tableChanged)
	require.NotNil(t, queryResult.Record())
	assert.Equal(t, queryResult.Record(), expectedRecord1)
	assert.Nil(t, queryResult.Record().Value())
	assert.Equal(t, 0, queryResult.TablePosition())

	require.True(t, queryResult.Next(), queryResult.Err())
	require.Nil(t, queryResult.Err())
	assert.False(t, queryResult.tableChanged)
	assert.Equal(t, 0, queryResult.TablePosition())
	require.NotNil(t, queryResult.Record())
	assert.Equal(t, queryResult.Record(), expectedRecord2)

	require.False(t, queryResult.Next())
	require.Nil(t, queryResult.Err())
}

func TestQueryRawResult(t *testing.T) {
	csvRows := []string{`#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string`,
		`#group,false,false,true,true,false,false,true,true,true,true`,
		`#default,_result,,,,,,,,,`,
		`,result,table,_start,_stop,_time,_value,_field,_measurement,a,b`,
		`,,0,2020-02-17T22:19:49.747562847Z,2020-02-18T22:19:49.747562847Z,2020-02-18T10:34:08.135814545Z,1.4,f,test,1,adsfasdf`,
		`,,0,2020-02-17T22:19:49.747562847Z,2020-02-18T22:19:49.747562847Z,2020-02-18T22:08:44.850214724Z,6.6,f,test,1,adsfasdf`,
		``,
		`#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string`,
		`#group,false,false,true,true,false,false,true,true,true,true`,
		`#default,_result2,,,,,,,,,`,
		`,result,table,_start,_stop,_time,_value,_field,_measurement,a,b`,
		`,,1,2020-02-17T22:19:49.747562847Z,2020-02-18T22:19:49.747562847Z,2020-02-18T10:34:08.135814545Z,4,i,test,1,adsfasdf`,
		`,,1,2020-02-17T22:19:49.747562847Z,2020-02-18T22:19:49.747562847Z,2020-02-18T22:08:44.850214724Z,1,i,test,1,adsfasdf`,
		``,
	}
	csvTable := strings.Join(csvRows, "\r\n")
	csvTable = fmt.Sprintf("%s\r\n", csvTable)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-time.After(100 * time.Millisecond)
		if r.Method == http.MethodPost {
			rbody, _ := ioutil.ReadAll(r.Body)
			fmt.Printf("Req: %s\n", string(rbody))
			body, err := gzip.CompressWithGzip(strings.NewReader(csvTable))
			if err == nil {
				var bytes []byte
				bytes, err = ioutil.ReadAll(body)
				if err == nil {
					w.Header().Set("Content-Type", "text/csv")
					w.Header().Set("Content-Encoding", "gzip")
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write(bytes)
				}
			}
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(err.Error()))
			}
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()
	queryAPI := NewQueryAPI("org", http2.NewService(server.URL, "a", http2.DefaultOptions()))

	result, err := queryAPI.QueryRaw(context.Background(), "flux", nil)
	require.Nil(t, err)
	require.NotNil(t, result)
	assert.Equal(t, csvTable, result)

}

func TestErrorInRow(t *testing.T) {
	csvRowsError := []string{
		`#datatype,string,string`,
		`#group,true,true`,
		`#default,,`,
		`,error,reference`,
		`,failed to create physical plan: invalid time bounds from procedure from: bounds contain zero time,897`}
	csvTable := makeCSVstring(csvRowsError)
	reader := strings.NewReader(csvTable)
	csvReader := csv.NewReader(reader)
	csvReader.FieldsPerRecord = -1
	queryResult := &QueryTableResult{Closer: ioutil.NopCloser(reader), csvReader: csvReader}

	require.False(t, queryResult.Next())
	require.NotNil(t, queryResult.Err())
	assert.Equal(t, "failed to create physical plan: invalid time bounds from procedure from: bounds contain zero time,897", queryResult.Err().Error())

	csvRowsErrorNoReference := []string{
		`#datatype,string,string`,
		`#group,true,true`,
		`#default,,`,
		`,error,reference`,
		`,failed to create physical plan: invalid time bounds from procedure from: bounds contain zero time,`}
	csvTable = makeCSVstring(csvRowsErrorNoReference)
	reader = strings.NewReader(csvTable)
	csvReader = csv.NewReader(reader)
	csvReader.FieldsPerRecord = -1
	queryResult = &QueryTableResult{Closer: ioutil.NopCloser(reader), csvReader: csvReader}

	require.False(t, queryResult.Next())
	require.NotNil(t, queryResult.Err())
	assert.Equal(t, "failed to create physical plan: invalid time bounds from procedure from: bounds contain zero time", queryResult.Err().Error())

	csvRowsErrorNoMessage := []string{
		`#datatype,string,string`,
		`#group,true,true`,
		`#default,,`,
		`,error,reference`,
		`,,`}
	csvTable = makeCSVstring(csvRowsErrorNoMessage)
	reader = strings.NewReader(csvTable)
	csvReader = csv.NewReader(reader)
	csvReader.FieldsPerRecord = -1
	queryResult = &QueryTableResult{Closer: ioutil.NopCloser(reader), csvReader: csvReader}

	require.False(t, queryResult.Next())
	require.NotNil(t, queryResult.Err())
	assert.Equal(t, "unknown query error", queryResult.Err().Error())
}

func TestInvalidDataType(t *testing.T) {
	csvTable := `#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,int,string,duration,base64Binary,dateTime:RFC3339
#group,false,false,true,true,false,true,true,false,false,false
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,deviceId,sensor,elapsed,note,start
,,0,2020-04-28T12:36:50.990018157Z,2020-04-28T12:51:50.990018157Z,2020-04-28T12:38:11.480545389Z,1467463,BME280,1m1s,ZGF0YWluYmFzZTY0,2020-04-27T00:00:00Z
,,0,2020-04-28T12:36:50.990018157Z,2020-04-28T12:51:50.990018157Z,2020-04-28T12:39:36.330153686Z,1467463,BME280,1h20m30.13245s,eHh4eHhjY2NjY2NkZGRkZA==,2020-04-28T00:00:00Z
`

	reader := strings.NewReader(csvTable)
	csvReader := csv.NewReader(reader)
	csvReader.FieldsPerRecord = -1
	queryResult := &QueryTableResult{Closer: ioutil.NopCloser(reader), csvReader: csvReader}
	require.False(t, queryResult.Next())
	require.NotNil(t, queryResult.Err())
	assert.Equal(t, "deviceId has unknown data type int", queryResult.Err().Error())
}

func TestReorderedAnnotations(t *testing.T) {
	expectedTable := query.NewFluxTableMetadataFull(0,
		[]*query.FluxColumn{
			query.NewFluxColumnFull("string", "_result", "result", false, 0),
			query.NewFluxColumnFull("long", "", "table", false, 1),
			query.NewFluxColumnFull("dateTime:RFC3339", "", "_start", true, 2),
			query.NewFluxColumnFull("dateTime:RFC3339", "", "_stop", true, 3),
			query.NewFluxColumnFull("dateTime:RFC3339", "", "_time", false, 4),
			query.NewFluxColumnFull("double", "", "_value", false, 5),
			query.NewFluxColumnFull("string", "", "_field", true, 6),
			query.NewFluxColumnFull("string", "", "_measurement", true, 7),
			query.NewFluxColumnFull("string", "", "a", true, 8),
			query.NewFluxColumnFull("string", "", "b", true, 9),
		},
	)
	expectedRecord1 := query.NewFluxRecord(0,
		map[string]interface{}{
			"result":       "_result",
			"table":        int64(0),
			"_start":       mustParseTime("2020-02-17T22:19:49.747562847Z"),
			"_stop":        mustParseTime("2020-02-18T22:19:49.747562847Z"),
			"_time":        mustParseTime("2020-02-18T10:34:08.135814545Z"),
			"_value":       1.4,
			"_field":       "f",
			"_measurement": "test",
			"a":            "1",
			"b":            "adsfasdf",
		},
	)

	expectedRecord2 := query.NewFluxRecord(0,
		map[string]interface{}{
			"result":       "_result",
			"table":        int64(0),
			"_start":       mustParseTime("2020-02-17T22:19:49.747562847Z"),
			"_stop":        mustParseTime("2020-02-18T22:19:49.747562847Z"),
			"_time":        mustParseTime("2020-02-18T22:08:44.850214724Z"),
			"_value":       6.6,
			"_field":       "f",
			"_measurement": "test",
			"a":            "1",
			"b":            "adsfasdf",
		},
	)

	csvTable1 := `#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,a,b
,,0,2020-02-17T22:19:49.747562847Z,2020-02-18T22:19:49.747562847Z,2020-02-18T10:34:08.135814545Z,1.4,f,test,1,adsfasdf
,,0,2020-02-17T22:19:49.747562847Z,2020-02-18T22:19:49.747562847Z,2020-02-18T22:08:44.850214724Z,6.6,f,test,1,adsfasdf

`
	reader := strings.NewReader(csvTable1)
	csvReader := csv.NewReader(reader)
	csvReader.FieldsPerRecord = -1
	queryResult := &QueryTableResult{Closer: ioutil.NopCloser(reader), csvReader: csvReader}
	require.True(t, queryResult.Next(), queryResult.Err())
	require.Nil(t, queryResult.Err())

	require.Equal(t, queryResult.table, expectedTable)
	assert.True(t, queryResult.tableChanged)
	require.NotNil(t, queryResult.Record())
	require.Equal(t, queryResult.Record(), expectedRecord1)

	require.True(t, queryResult.Next(), queryResult.Err())
	require.Nil(t, queryResult.Err())
	assert.False(t, queryResult.tableChanged)
	require.NotNil(t, queryResult.Record())
	require.Equal(t, queryResult.Record(), expectedRecord2)

	require.False(t, queryResult.Next())
	require.Nil(t, queryResult.Err())

	csvTable2 := `#default,_result,,,,,,,,,
#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
,result,table,_start,_stop,_time,_value,_field,_measurement,a,b
,,0,2020-02-17T22:19:49.747562847Z,2020-02-18T22:19:49.747562847Z,2020-02-18T10:34:08.135814545Z,1.4,f,test,1,adsfasdf
,,0,2020-02-17T22:19:49.747562847Z,2020-02-18T22:19:49.747562847Z,2020-02-18T22:08:44.850214724Z,6.6,f,test,1,adsfasdf

`
	reader = strings.NewReader(csvTable2)
	csvReader = csv.NewReader(reader)
	csvReader.FieldsPerRecord = -1
	queryResult = &QueryTableResult{Closer: ioutil.NopCloser(reader), csvReader: csvReader}
	require.True(t, queryResult.Next(), queryResult.Err())
	require.Nil(t, queryResult.Err())

	require.Equal(t, queryResult.table, expectedTable)
	assert.True(t, queryResult.tableChanged)
	require.NotNil(t, queryResult.Record())
	require.Equal(t, queryResult.Record(), expectedRecord1)

	require.True(t, queryResult.Next(), queryResult.Err())
	require.Nil(t, queryResult.Err())
	assert.False(t, queryResult.tableChanged)
	require.NotNil(t, queryResult.Record())
	require.Equal(t, queryResult.Record(), expectedRecord2)

	require.False(t, queryResult.Next())
	require.Nil(t, queryResult.Err())
}

func TestDatatypeOnlyAnnotation(t *testing.T) {
	expectedTable := query.NewFluxTableMetadataFull(0,
		[]*query.FluxColumn{
			query.NewFluxColumnFull("string", "", "result", false, 0),
			query.NewFluxColumnFull("long", "", "table", false, 1),
			query.NewFluxColumnFull("dateTime:RFC3339", "", "_start", false, 2),
			query.NewFluxColumnFull("dateTime:RFC3339", "", "_stop", false, 3),
			query.NewFluxColumnFull("dateTime:RFC3339", "", "_time", false, 4),
			query.NewFluxColumnFull("double", "", "_value", false, 5),
			query.NewFluxColumnFull("string", "", "_field", false, 6),
			query.NewFluxColumnFull("string", "", "_measurement", false, 7),
			query.NewFluxColumnFull("string", "", "a", false, 8),
			query.NewFluxColumnFull("string", "", "b", false, 9),
		},
	)
	expectedRecord1 := query.NewFluxRecord(0,
		map[string]interface{}{
			"result":       nil,
			"table":        int64(0),
			"_start":       mustParseTime("2020-02-17T22:19:49.747562847Z"),
			"_stop":        mustParseTime("2020-02-18T22:19:49.747562847Z"),
			"_time":        mustParseTime("2020-02-18T10:34:08.135814545Z"),
			"_value":       1.4,
			"_field":       "f",
			"_measurement": "test",
			"a":            "1",
			"b":            "adsfasdf",
		},
	)

	expectedRecord2 := query.NewFluxRecord(0,
		map[string]interface{}{
			"result":       nil,
			"table":        int64(0),
			"_start":       mustParseTime("2020-02-17T22:19:49.747562847Z"),
			"_stop":        mustParseTime("2020-02-18T22:19:49.747562847Z"),
			"_time":        mustParseTime("2020-02-18T22:08:44.850214724Z"),
			"_value":       6.6,
			"_field":       "f",
			"_measurement": "test",
			"a":            "1",
			"b":            "adsfasdf",
		},
	)

	csvTable1 := `#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
,result,table,_start,_stop,_time,_value,_field,_measurement,a,b
,,0,2020-02-17T22:19:49.747562847Z,2020-02-18T22:19:49.747562847Z,2020-02-18T10:34:08.135814545Z,1.4,f,test,1,adsfasdf
,,0,2020-02-17T22:19:49.747562847Z,2020-02-18T22:19:49.747562847Z,2020-02-18T22:08:44.850214724Z,6.6,f,test,1,adsfasdf

`
	reader := strings.NewReader(csvTable1)
	csvReader := csv.NewReader(reader)
	csvReader.FieldsPerRecord = -1
	queryResult := &QueryTableResult{Closer: ioutil.NopCloser(reader), csvReader: csvReader}
	require.True(t, queryResult.Next(), queryResult.Err())
	require.Nil(t, queryResult.Err())

	require.Equal(t, queryResult.table, expectedTable)
	assert.True(t, queryResult.tableChanged)
	require.NotNil(t, queryResult.Record())
	require.Equal(t, queryResult.Record(), expectedRecord1)

	require.True(t, queryResult.Next(), queryResult.Err())
	require.Nil(t, queryResult.Err())
	assert.False(t, queryResult.tableChanged)
	require.NotNil(t, queryResult.Record())
	require.Equal(t, queryResult.Record(), expectedRecord2)

	require.False(t, queryResult.Next())
	require.Nil(t, queryResult.Err())
}

func TestMissingDatatypeAnnotation(t *testing.T) {
	csvTable1 := `
#group,false,false,true,true,false,true,true,false,false,false
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,deviceId,sensor,elapsed,note,start
,,0,2020-04-28T12:36:50.990018157Z,2020-04-28T12:51:50.990018157Z,2020-04-28T12:38:11.480545389Z,1467463,BME280,1m1s,ZGF0YWluYmFzZTY0,2020-04-27T00:00:00Z
,,0,2020-04-28T12:36:50.990018157Z,2020-04-28T12:51:50.990018157Z,2020-04-28T12:39:36.330153686Z,1467463,BME280,1h20m30.13245s,eHh4eHhjY2NjY2NkZGRkZA==,2020-04-28T00:00:00Z
`

	reader := strings.NewReader(csvTable1)
	csvReader := csv.NewReader(reader)
	csvReader.FieldsPerRecord = -1
	queryResult := &QueryTableResult{Closer: ioutil.NopCloser(reader), csvReader: csvReader}
	require.False(t, queryResult.Next())
	require.NotNil(t, queryResult.Err())
	assert.Equal(t, "parsing error, datatype annotation not found", queryResult.Err().Error())

	csvTable2 := `
#default,_result,,,,,,,,,
#group,false,false,true,true,false,true,true,false,false,false
,result,table,_start,_stop,_time,deviceId,sensor,elapsed,note,start
,,0,2020-04-28T12:36:50.990018157Z,2020-04-28T12:51:50.990018157Z,2020-04-28T12:38:11.480545389Z,1467463,BME280,1m1s,ZGF0YWluYmFzZTY0,2020-04-27T00:00:00Z
,,0,2020-04-28T12:36:50.990018157Z,2020-04-28T12:51:50.990018157Z,2020-04-28T12:39:36.330153686Z,1467463,BME280,1h20m30.13245s,eHh4eHhjY2NjY2NkZGRkZA==,2020-04-28T00:00:00Z
`

	reader = strings.NewReader(csvTable2)
	csvReader = csv.NewReader(reader)
	csvReader.FieldsPerRecord = -1
	queryResult = &QueryTableResult{Closer: ioutil.NopCloser(reader), csvReader: csvReader}
	require.False(t, queryResult.Next())
	require.NotNil(t, queryResult.Err())
	assert.Equal(t, "parsing error, datatype annotation not found", queryResult.Err().Error())
}

func TestMissingAnnotations(t *testing.T) {
	csvTable3 := `
,result,table,_start,_stop,_time,deviceId,sensor,elapsed,note,start
,,0,2020-04-28T12:36:50.990018157Z,2020-04-28T12:51:50.990018157Z,2020-04-28T12:38:11.480545389Z,1467463,BME280,1m1s,ZGF0YWluYmFzZTY0,2020-04-27T00:00:00Z
,,0,2020-04-28T12:36:50.990018157Z,2020-04-28T12:51:50.990018157Z,2020-04-28T12:39:36.330153686Z,1467463,BME280,1h20m30.13245s,eHh4eHhjY2NjY2NkZGRkZA==,2020-04-28T00:00:00Z

`
	reader := strings.NewReader(csvTable3)
	csvReader := csv.NewReader(reader)
	csvReader.FieldsPerRecord = -1
	queryResult := &QueryTableResult{Closer: ioutil.NopCloser(reader), csvReader: csvReader}
	require.False(t, queryResult.Next())
	require.NotNil(t, queryResult.Err())
	assert.Equal(t, "parsing error, annotations not found", queryResult.Err().Error())
}

func TestDifferentNumberOfColumns(t *testing.T) {
	csvTable := `#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,int,string,duration,base64Binary,dateTime:RFC3339
#group,false,false,true,true,false,true,true,false,false,
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,deviceId,sensor,elapsed,note,start
,,0,2020-04-28T12:36:50.990018157Z,2020-04-28T12:51:50.990018157Z,2020-04-28T12:38:11.480545389Z,1467463,BME280,1m1s,ZGF0YWluYmFzZTY0,2020-04-27T00:00:00Z,2345234
`

	reader := strings.NewReader(csvTable)
	csvReader := csv.NewReader(reader)
	csvReader.FieldsPerRecord = -1
	queryResult := &QueryTableResult{Closer: ioutil.NopCloser(reader), csvReader: csvReader}
	require.False(t, queryResult.Next())
	require.NotNil(t, queryResult.Err())
	assert.Equal(t, "parsing error, row has different number of columns than the table: 11 vs 10", queryResult.Err().Error())

	csvTable2 := `#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,int,string,duration,base64Binary,dateTime:RFC3339
#group,false,false,true,true,false,true,true,false,false,
#default,_result,,,,,,,
,result,table,_start,_stop,_time,deviceId,sensor,elapsed,note,start
,,0,2020-04-28T12:36:50.990018157Z,2020-04-28T12:51:50.990018157Z,2020-04-28T12:38:11.480545389Z,1467463,BME280,1m1s,ZGF0YWluYmFzZTY0,2020-04-27T00:00:00Z,2345234
`

	reader = strings.NewReader(csvTable2)
	csvReader = csv.NewReader(reader)
	csvReader.FieldsPerRecord = -1
	queryResult = &QueryTableResult{Closer: ioutil.NopCloser(reader), csvReader: csvReader}
	require.False(t, queryResult.Next())
	require.NotNil(t, queryResult.Err())
	assert.Equal(t, "parsing error, row has different number of columns than the table: 8 vs 10", queryResult.Err().Error())

	csvTable3 := `#default,_result,,,,,,,
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,int,string,duration,base64Binary,dateTime:RFC3339
#group,false,false,true,true,false,true,true,false,false,
,result,table,_start,_stop,_time,deviceId,sensor,elapsed,note,start
,,0,2020-04-28T12:36:50.990018157Z,2020-04-28T12:51:50.990018157Z,2020-04-28T12:38:11.480545389Z,1467463,BME280,1m1s,ZGF0YWluYmFzZTY0,2020-04-27T00:00:00Z,2345234
`

	reader = strings.NewReader(csvTable3)
	csvReader = csv.NewReader(reader)
	csvReader.FieldsPerRecord = -1
	queryResult = &QueryTableResult{Closer: ioutil.NopCloser(reader), csvReader: csvReader}
	require.False(t, queryResult.Next())
	require.NotNil(t, queryResult.Err())
	assert.Equal(t, "parsing error, row has different number of columns than the table: 10 vs 8", queryResult.Err().Error())
}

func TestEmptyValue(t *testing.T) {
	csvTable := `#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#group,false,false,true,true,false,false,true,true,true,true
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,a,b
,,0,2020-02-17T22:19:49.747562847Z,2020-02-18T22:19:49.747562847Z,2020-02-18T10:34:08.135814545Z,,f,test,1,adsfasdf
,,0,2020-02-17T22:19:49.747562847Z,2020-02-18T22:19:49.747562847Z,2020-02-18T22:08:44.850214724Z,6.6,f,test,,adsfasdf
,,0,2020-02-17T22:19:49.747562847Z,2020-02-18T22:19:49.747562847Z,2020-02-18T22:11:32.225467895Z,1122.45,f,test,3,
`

	reader := strings.NewReader(csvTable)
	csvReader := csv.NewReader(reader)
	csvReader.FieldsPerRecord = -1
	queryResult := &QueryTableResult{Closer: ioutil.NopCloser(reader), csvReader: csvReader}

	require.True(t, queryResult.Next(), queryResult.Err())
	require.Nil(t, queryResult.Err())

	require.NotNil(t, queryResult.Record())
	assert.Nil(t, queryResult.Record().Value())

	require.True(t, queryResult.Next(), queryResult.Err())
	require.NotNil(t, queryResult.Record())
	assert.Nil(t, queryResult.Record().ValueByKey("a"))

	require.True(t, queryResult.Next(), queryResult.Err())
	require.NotNil(t, queryResult.Record())
	assert.Nil(t, queryResult.Record().ValueByKey("b"))

	require.False(t, queryResult.Next())
	require.Nil(t, queryResult.Err())
}

func TestFluxError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-time.After(100 * time.Millisecond)
		if r.Method == http.MethodPost {
			_, _ = ioutil.ReadAll(r.Body)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"code":"invalid","message":"compilation failed: loc 4:17-4:86: expected an operator between two expressions"}`))
		}
	}))
	defer server.Close()
	queryAPI := NewQueryAPI("org", http2.NewService(server.URL, "a", http2.DefaultOptions()))

	result, err := queryAPI.QueryRaw(context.Background(), "errored flux", nil)
	assert.Equal(t, "", result)
	require.NotNil(t, err)
	assert.Equal(t, "invalid: compilation failed: loc 4:17-4:86: expected an operator between two expressions", err.Error())

	tableRes, err := queryAPI.Query(context.Background(), "errored flux")
	assert.Nil(t, tableRes)
	require.NotNil(t, err)
	assert.Equal(t, "invalid: compilation failed: loc 4:17-4:86: expected an operator between two expressions", err.Error())

}

func makeCSVstring(rows []string) string {
	csvTable := strings.Join(rows, "\r\n")
	return fmt.Sprintf("%s\r\n", csvTable)
}
