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

	assert.Equal(t, table.Position(), 1)
	require.NotNil(t, table.Column(0))
	assert.Equal(t, table.Column(0).DefaultValue(), "_result")
	assert.Equal(t, table.Column(0).DataType(), "string")
	assert.Equal(t, table.Column(0).Name(), "result")
	assert.Equal(t, table.Column(0).Index(), 0)
	assert.Equal(t, table.Column(0).IsGroup(), false)

	require.NotNil(t, table.Column(1))
	assert.Equal(t, table.Column(1).DefaultValue(), "10")
	assert.Equal(t, table.Column(1).DataType(), "long")
	assert.Equal(t, table.Column(1).Name(), "_table")
	assert.Equal(t, table.Column(1).Index(), 1)
	assert.Equal(t, table.Column(1).IsGroup(), false)

	require.NotNil(t, table.Column(2))
	assert.Equal(t, table.Column(2).DefaultValue(), "")
	assert.Equal(t, table.Column(2).DataType(), "dateTime:RFC3339")
	assert.Equal(t, table.Column(2).Name(), "_start")
	assert.Equal(t, table.Column(2).Index(), 2)
	assert.Equal(t, table.Column(2).IsGroup(), true)

	require.NotNil(t, table.Column(3))
	assert.Equal(t, table.Column(3).DefaultValue(), "1.1")
	assert.Equal(t, table.Column(3).DataType(), "double")
	assert.Equal(t, table.Column(3).Name(), "_value")
	assert.Equal(t, table.Column(3).Index(), 3)
	assert.Equal(t, table.Column(3).IsGroup(), false)

	require.NotNil(t, table.Column(4))
	assert.Equal(t, table.Column(4).DefaultValue(), "")
	assert.Equal(t, table.Column(4).DataType(), "string")
	assert.Equal(t, table.Column(4).Name(), "_field")
	assert.Equal(t, table.Column(4).Index(), 4)
	assert.Equal(t, table.Column(4).IsGroup(), true)
}

func TestRecord(t *testing.T) {
	record := &FluxRecord{table: 2,
		values: map[string]interface{}{
			"result":       "_result",
			"_table":       int64(0),
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
	assert.Equal(t, record.Start(), mustParseTime("2020-02-17T22:19:49.747562847Z"))
	assert.Equal(t, record.Stop(), mustParseTime("2020-02-18T22:19:49.747562847Z"))
	assert.Equal(t, record.Time(), mustParseTime("2020-02-18T10:34:08.135814545Z"))
	assert.Equal(t, record.Field(), "f")
	assert.Equal(t, record.Value(), 1.4)
	assert.Equal(t, record.Measurement(), "test")
	assert.Equal(t, record.Table(), 2)
}
