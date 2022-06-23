// Copyright 2020-2021 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package influxclient

import (
	"fmt"
	"sort"
	"time"

	"github.com/influxdata/line-protocol/v2/lineprotocol"
)

// Tag holds the keys and values for a bunch of Tag k/v pairs.
type Tag struct {
	Key   string
	Value string
}

// Field holds the keys and values for a bunch of Metric Field k/v pairs where Value can be a uint64, int64, int, float32, float64, string, or bool.
type Field struct {
	Key   string
	Value lineprotocol.Value
}

// Point is represents InfluxDB time series point, holding tags and fields
type Point struct {
	Measurement string
	Tags        []Tag
	Fields      []Field
	Timestamp   time.Time
}

// NewPointWithMeasurement is a convenient function for creating a Point from measurement name for later adding data
func NewPointWithMeasurement(measurement string) *Point {
	return &Point{
		Measurement: measurement,
	}
}

// NewPoint is a convenient function for creating a Point from measurement name, tags, fields and a timestamp.
func NewPoint(measurement string, tags map[string]string, fields map[string]interface{}, ts time.Time) *Point {
	m := &Point{
		Measurement: measurement,
		Timestamp:   ts,
	}
	if len(tags) > 0 {
		m.Tags = make([]Tag, 0, len(tags))
		for k, v := range tags {
			m.AddTag(k, v)
		}
		m.SortTags()
	}
	if len(fields) > 0 {
		m.Fields = make([]Field, 0, len(fields))
		for k, v := range fields {
			m.AddField(k, v)
		}
		m.SortFields()
	}
	return m
}

// SortTags orders the tags of a point alphanumerically by key.
// This is just here as a helper, to make it easy to keep tags sorted if you are creating a Point manually.
func (m *Point) SortTags() *Point {
	sort.Slice(m.Tags, func(i, j int) bool { return m.Tags[i].Key < m.Tags[j].Key })
	return m
}

// SortFields orders the fields of a point alphanumerically by key.
func (m *Point) SortFields() *Point {
	sort.Slice(m.Fields, func(i, j int) bool { return m.Fields[i].Key < m.Fields[j].Key })
	return m
}

// AddTag adds a tag to a point.
func (m *Point) AddTag(k, v string) *Point {
	for i, tag := range m.Tags {
		if k == tag.Key {
			m.Tags[i].Value = v
			return m
		}
	}
	m.Tags = append(m.Tags, Tag{Key: k, Value: v})
	return m
}

// AddField adds a field to a point.
func (m *Point) AddField(k string, v interface{}) *Point {
	val, _ := lineprotocol.NewValue(convertField(v))
	for i, field := range m.Fields {
		if k == field.Key {
			m.Fields[i].Value = val
			return m
		}
	}

	m.Fields = append(m.Fields, Field{Key: k, Value: val})
	return m
}

// SetTimestamp is helper function for complete fluent interface
func (m *Point) SetTimestamp(t time.Time) *Point {
	m.Timestamp = t
	return m
}

func (m *Point) MarshalBinary(precision lineprotocol.Precision) ([]byte, error) {
	var enc lineprotocol.Encoder
	enc.SetPrecision(precision)
	enc.StartLine(m.Measurement)
	m.SortTags()
	for _, t := range m.Tags {
		enc.AddTag(t.Key, t.Value)
	}
	m.SortFields()
	for _, f := range m.Fields {
		enc.AddField(f.Key, f.Value)
	}
	enc.EndLine(m.Timestamp)
	if err := enc.Err(); err != nil {
		return nil, fmt.Errorf("encoding error: %v", err)
	}
	return enc.Bytes(), nil
}

// convertField converts any primitive type to types supported by line protocol
func convertField(v interface{}) interface{} {
	switch v := v.(type) {
	case bool, int64, uint64, string, float64:
		return v
	case int:
		return int64(v)
	case uint:
		return uint64(v)
	case []byte:
		return string(v)
	case int32:
		return int64(v)
	case int16:
		return int64(v)
	case int8:
		return int64(v)
	case uint32:
		return uint64(v)
	case uint16:
		return uint64(v)
	case uint8:
		return uint64(v)
	case float32:
		return float64(v)
	case time.Time:
		return v.Format(time.RFC3339Nano)
	case time.Duration:
		return v.String()
	default:
		return fmt.Sprintf("%v", v)
	}
}
