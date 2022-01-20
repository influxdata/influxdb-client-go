// Copyright 2020-2021 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package query

import (
	"testing"
	"time"

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

func TestTable(t *testing.T) {
	table := &FluxTableMetadata{position: 1}
	table.AddColumn(&FluxColumn{dataType: "string", defaultValue: "_result", name: "result", group: false, index: 0})
	table.AddColumn(&FluxColumn{dataType: "long", defaultValue: "10", name: "_table", group: false, index: 1})
	table.AddColumn(&FluxColumn{dataType: "dateTime:RFC3339", defaultValue: "", name: "_start", group: true, index: 2})
	table.AddColumn(&FluxColumn{dataType: "double", defaultValue: "1.1", name: "_value", group: false, index: 3})
	table.AddColumn(&FluxColumn{dataType: "string", defaultValue: "", name: "_field", group: true, index: 4})
	require.Len(t, table.columns, 5)

	assert.Equal(t, 1, table.Position())
	require.NotNil(t, table.Column(0))
	assert.Equal(t, "_result", table.Column(0).DefaultValue())
	assert.Equal(t, "string", table.Column(0).DataType())
	assert.Equal(t, "result", table.Column(0).Name())
	assert.Equal(t, 0, table.Column(0).Index())
	assert.Equal(t, false, table.Column(0).IsGroup())

	require.NotNil(t, table.Column(1))
	assert.Equal(t, "10", table.Column(1).DefaultValue())
	assert.Equal(t, "long", table.Column(1).DataType())
	assert.Equal(t, "_table", table.Column(1).Name())
	assert.Equal(t, 1, table.Column(1).Index())
	assert.Equal(t, false, table.Column(1).IsGroup())

	require.NotNil(t, table.Column(2))
	assert.Equal(t, "", table.Column(2).DefaultValue())
	assert.Equal(t, "dateTime:RFC3339", table.Column(2).DataType())
	assert.Equal(t, "_start", table.Column(2).Name())
	assert.Equal(t, 2, table.Column(2).Index())
	assert.Equal(t, true, table.Column(2).IsGroup())

	require.NotNil(t, table.Column(3))
	assert.Equal(t, "1.1", table.Column(3).DefaultValue())
	assert.Equal(t, "double", table.Column(3).DataType())
	assert.Equal(t, "_value", table.Column(3).Name())
	assert.Equal(t, 3, table.Column(3).Index())
	assert.Equal(t, false, table.Column(3).IsGroup())

	require.NotNil(t, table.Column(4))
	assert.Equal(t, "", table.Column(4).DefaultValue())
	assert.Equal(t, "string", table.Column(4).DataType())
	assert.Equal(t, "_field", table.Column(4).Name())
	assert.Equal(t, 4, table.Column(4).Index())
	assert.Equal(t, true, table.Column(4).IsGroup())
}

func TestRecord(t *testing.T) {
	record := &FluxRecord{table: 2,
		values: map[string]interface{}{
			"result":       "_result",
			"table":        int64(2),
			"_start":       mustParseTime("2020-02-17T22:19:49.747562847Z"),
			"_stop":        mustParseTime("2020-02-18T22:19:49.747562847Z"),
			"_time":        mustParseTime("2020-02-18T10:34:08.135814545Z"),
			"_value":       1.4,
			"_field":       "f",
			"_measurement": "test",
			"a":            "1",
			"b":            "adsfasdf",
		},
	}
	require.Len(t, record.values, 10)
	assert.Equal(t, mustParseTime("2020-02-17T22:19:49.747562847Z"), record.Start())
	assert.Equal(t, mustParseTime("2020-02-18T22:19:49.747562847Z"), record.Stop())
	assert.Equal(t, mustParseTime("2020-02-18T10:34:08.135814545Z"), record.Time())
	assert.Equal(t, "_result", record.Result())
	assert.Equal(t, "f", record.Field())
	assert.Equal(t, 1.4, record.Value())
	assert.Equal(t, "test", record.Measurement())
	assert.Equal(t, 2, record.Table())

	agRec := &FluxRecord{table: 0,
		values: map[string]interface{}{
			"room":   "bathroom",
			"sensor": "SHT",
			"temp":   24.3,
			"hum":    42,
		},
	}
	require.Len(t, agRec.values, 4)
	assert.Equal(t, time.Time{}, agRec.Start())
	assert.Equal(t, time.Time{}, agRec.Stop())
	assert.Equal(t, time.Time{}, agRec.Time())
	assert.Equal(t, "", agRec.Field())
	assert.Equal(t, "", agRec.Result())
	assert.Nil(t, agRec.Value())
	assert.Equal(t, "", agRec.Measurement())
	assert.Equal(t, 0, agRec.Table())
	assert.Equal(t, 24.3, agRec.ValueByKey("temp"))
	assert.Equal(t, 42, agRec.ValueByKey("hum"))
	assert.Nil(t, agRec.ValueByKey("notexist"))
}
