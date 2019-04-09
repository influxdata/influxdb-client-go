package client

import (
	"sort"
	"time"

	lp "github.com/influxdata/line-protocol"
)

type LMetric = lp.Metric

// Metric is an github.com/influxdata/line-protocol.Metric,
// that has methods to make it easy to add tags and fields
type Metric struct {
	Name   string
	Tags   []*lp.Tag
	Fields []*lp.Field
	TS     time.Time
}

// TagList returns a slice containing Tags of a Metric.
func (m *Metric) TagList() []*lp.Tag {
	return m.Tags
}

// FieldList returns a slice containing the Fields of a Metric.
func (m *Metric) FieldList() []*lp.Field {
	return m.Fields
}

// Time is the timestamp of a metric.
func (m *Metric) Time() time.Time {
	return m.TS
}

// SortTags orders the tags of a a metric alphnumerically by key.
func (m *Metric) SortTags() {
	sort.Slice(m.Tags, func(i, j int) bool { return m.Tags[i].Key < m.Tags[j].Key })
}

// AddTag adds an lp.Tag to a metric.
func (m *Metric) AddTag(k, v string) {
	for i, tag := range m.Tags {
		if k == tag.Key {
			m.Tags[i].Value = v
			return
		}
	}
	m.Tags = append(m.Tags, &lp.Tag{Key: k, Value: v})
}

// AddField adds an lp.Field to a metric.
func (m *Metric) AddField(k string, v interface{}) {
	for i, field := range m.Fields {
		if k == field.Key {
			m.Fields[i].Value = v
			return
		}
	}
	m.SortTags()
}

func convertField(v interface{}) interface{} {
	switch v := v.(type) {
	case float64:
		return v
	case int64:
		return v
	case string:
		return v
	case bool:
		return v
	case int:
		return int64(v)
	case uint:
		return uint64(v)
	case uint64:
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
	default:
		return nil
	}
}

// NewMetric creates a *Metric from tags, fields and a timestamp.
func NewMetric(
	fields map[string]interface{},
	name string,
	tags map[string]string,
	tm time.Time,
) *Metric {
	m := &Metric{
		Name:   name,
		Tags:   nil,
		Fields: nil,
		TS:     tm,
	}

	if len(tags) > 0 {
		m.Tags = make([]*lp.Tag, 0, len(tags))
		for k, v := range tags {
			m.Tags = append(m.Tags,
				&lp.Tag{Key: k, Value: v})
		}
		sort.Slice(m.Tags, func(i, j int) bool { return m.Tags[i].Key < m.Tags[j].Key })
	}

	m.Fields = make([]*lp.Field, 0, len(fields))
	for k, v := range fields {
		v := convertField(v)
		if v == nil {
			continue
		}
		m.AddField(k, v)
	}

	return m
}
