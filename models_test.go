package client

import (
	"testing"
	"time"

	cmp "github.com/google/go-cmp/cmp"
	lp "github.com/influxdata/line-protocol"
)

func TestRowMetric_TagList(t *testing.T) {
	type fields struct {
		Name   string
		Tags   []*lp.Tag
		Fields []*lp.Field
		TS     time.Time
	}
	tests := []struct {
		name   string
		fields fields
		want   []*lp.Tag
	}{
		{
			name: "test FieldList",
			fields: fields{
				Name:   "testmetric",
				Fields: []*lp.Field{{Key: "field1", Value: int64(1)}, {Key: "field2", Value: int64(2)}},
				Tags:   []*lp.Tag{{Key: "tag1", Value: "1"}, {Key: "tag2", Value: "2"}, {Key: "tag3", Value: "3"}},
				TS:     time.Date(2001, 1, 1, 1, 0, 0, 10, time.UTC),
			},
			want: []*lp.Tag{{Key: "tag1", Value: "1"}, {Key: "tag2", Value: "2"}, {Key: "tag3", Value: "3"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &RowMetric{
				NameStr: tt.fields.Name,
				Tags:    tt.fields.Tags,
				Fields:  tt.fields.Fields,
				TS:      tt.fields.TS,
			}
			if got := m.TagList(); !cmp.Equal(got, tt.want) {
				t.Errorf("Diff: %s", cmp.Diff(got, tt.want))
			}
		})
	}
}

func TestRowMetric_FieldList(t *testing.T) {
	type fields struct {
		Name   string
		Tags   []*lp.Tag
		Fields []*lp.Field
		TS     time.Time
	}
	tests := []struct {
		name   string
		fields fields
		want   []*lp.Field
	}{
		{
			name: "test FieldList",
			fields: fields{
				Name:   "testmetric",
				Fields: []*lp.Field{{Key: "field1", Value: int64(1)}, {Key: "field2", Value: int64(2)}},
				Tags:   []*lp.Tag{{Key: "tag1", Value: "1"}, {Key: "tag2", Value: "2"}, {Key: "tag3", Value: "3"}},
				TS:     time.Date(2001, 1, 1, 1, 0, 0, 10, time.UTC),
			},
			want: []*lp.Field{{Key: "field1", Value: int64(1)}, {Key: "field2", Value: int64(2)}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &RowMetric{
				NameStr: tt.fields.Name,
				Tags:    tt.fields.Tags,
				Fields:  tt.fields.Fields,
				TS:      tt.fields.TS,
			}
			if got := m.FieldList(); !cmp.Equal(got, tt.want) {
				t.Errorf("Diff: %s", cmp.Diff(got, tt.want))
			}
		})
	}
}

func TestRowMetric_Time(t *testing.T) {
	type fields struct {
		Name   string
		Tags   []*lp.Tag
		Fields []*lp.Field
		TS     time.Time
	}
	tests := []struct {
		name   string
		fields fields
		want   time.Time
	}{
		{
			name: "test Time",
			fields: fields{
				Name:   "testmetric",
				Fields: []*lp.Field{{Key: "field1", Value: int64(1)}, {Key: "field2", Value: int64(2)}},
				Tags:   []*lp.Tag{{Key: "tag1", Value: "1"}, {Key: "tag3", Value: "3"}, {Key: "tag2", Value: "2"}},
				TS:     time.Date(2001, 1, 1, 1, 0, 0, 10, time.UTC),
			},
			want: time.Date(2001, 1, 1, 1, 0, 0, 10, time.UTC),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &RowMetric{
				NameStr: tt.fields.Name,
				Tags:    tt.fields.Tags,
				Fields:  tt.fields.Fields,
				TS:      tt.fields.TS,
			}
			if got := m.Time(); !cmp.Equal(got, tt.want) {
				t.Errorf("Diff: %s", cmp.Diff(got, tt.want))
			}
		})
	}
}

func TestRowMetric_SortTags(t *testing.T) {
	type fields struct {
		NameStr string
		Tags    []*lp.Tag
		Fields  []*lp.Field
		TS      time.Time
	}
	tests := []struct {
		name   string
		fields fields
		want   *RowMetric
	}{
		{
			name: "test SortTags",
			fields: fields{
				NameStr: "testmetric",
				Fields:  []*lp.Field{{Key: "field1", Value: int64(1)}, {Key: "field2", Value: int64(2)}},
				Tags:    []*lp.Tag{{Key: "tag1", Value: "1"}, {Key: "tag3", Value: "3"}, {Key: "tag2", Value: "2"}},
				TS:      time.Date(2001, 1, 1, 1, 0, 0, 10, time.UTC),
			},
			want: &RowMetric{
				NameStr: "testmetric",
				Fields:  []*lp.Field{{Key: "field1", Value: int64(1)}, {Key: "field2", Value: int64(2)}},
				Tags:    []*lp.Tag{{Key: "tag1", Value: "1"}, {Key: "tag2", Value: "2"}, {Key: "tag3", Value: "3"}},
				TS:      time.Date(2001, 1, 1, 1, 0, 0, 10, time.UTC),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &RowMetric{
				NameStr: tt.fields.NameStr,
				Tags:    tt.fields.Tags,
				Fields:  tt.fields.Fields,
				TS:      tt.fields.TS,
			}
			m.SortTags()
			if !cmp.Equal(m, tt.want) {
				t.Errorf("Diff: %s", cmp.Diff(m, tt.want))
			}
		})
	}
}

func TestRowMetric_AddTag(t *testing.T) {
	type fields struct {
		NameStr string
		Tags    []*lp.Tag
		Fields  []*lp.Field
		TS      time.Time
	}
	type args struct {
		k string
		v string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *RowMetric
	}{
		{
			name: "test AddTag",
			fields: fields{
				NameStr: "testmetric",
				Fields:  []*lp.Field{{Key: "field1", Value: int64(1)}, {Key: "field2", Value: int64(2)}},
				Tags:    []*lp.Tag{{Key: "tag1", Value: "1"}, {Key: "tag2", Value: "2"}},
				TS:      time.Date(2001, 1, 1, 1, 0, 0, 10, time.UTC),
			},
			args: args{k: "newTag", v: "tag3"},
			want: &RowMetric{
				NameStr: "testmetric",
				Fields:  []*lp.Field{{Key: "field1", Value: int64(1)}, {Key: "field2", Value: int64(2)}},
				Tags:    []*lp.Tag{{Key: "tag1", Value: "1"}, {Key: "tag2", Value: "2"}, {Key: "newTag", Value: "tag3"}},
				TS:      time.Date(2001, 1, 1, 1, 0, 0, 10, time.UTC),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &RowMetric{
				NameStr: tt.fields.NameStr,
				Tags:    tt.fields.Tags,
				Fields:  tt.fields.Fields,
				TS:      tt.fields.TS,
			}
			m.AddTag(tt.args.k, tt.args.v)
			if !cmp.Equal(m, tt.want) {
				t.Errorf("Diff: %s", cmp.Diff(m, tt.want))
			}
		})
	}
}

func TestRowMetric_AddField(t *testing.T) {
	type fields struct {
		Name   string
		Tags   []*lp.Tag
		Fields []*lp.Field
		TS     time.Time
	}
	type args struct {
		k string
		v interface{}
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *RowMetric
	}{
		{
			name: "test AddField",
			fields: fields{
				Name:   "testmetric",
				Fields: []*lp.Field{{Key: "field1", Value: int64(1)}, {Key: "field2", Value: int64(2)}},
				Tags:   []*lp.Tag{{Key: "tag1", Value: "1"}, {Key: "tag2", Value: "2"}},
				TS:     time.Date(2001, 1, 1, 1, 0, 0, 10, time.UTC),
			},
			args: args{
				k: "newField",
				v: 3,
			},
			want: &RowMetric{
				NameStr: "testmetric",
				Fields: []*lp.Field{
					{Key: "field1", Value: int64(1)},
					{Key: "field2", Value: int64(2)},
					{Key: "newField", Value: int64(3)},
				},
				Tags: []*lp.Tag{
					{Key: "tag1", Value: "1"},
					{Key: "tag2", Value: "2"},
				},
				TS: time.Date(2001, 1, 1, 1, 0, 0, 10, time.UTC),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &RowMetric{
				NameStr: tt.fields.Name,
				Tags:    tt.fields.Tags,
				Fields:  tt.fields.Fields,
				TS:      tt.fields.TS,
			}
			m.AddField(tt.args.k, tt.args.v)
			if !cmp.Equal(m, tt.want) {
				t.Errorf("Diff: %s", cmp.Diff(m, tt.want))
			}
		})
	}
}

func TestNewRowMetric(t *testing.T) {
	type args struct {
		fields map[string]interface{}
		name   string
		tags   map[string]string
		ts     time.Time
	}
	tests := []struct {
		name string
		args args
		want *RowMetric
	}{
		{
			name: "testNewRowMetric",
			args: args{
				name:   "testmetric",
				fields: map[string]interface{}{"field1": 1, "field2": 2},
				tags:   map[string]string{"tag1": "1", "tag2": "2"},
				ts:     time.Date(2001, 1, 1, 1, 0, 0, 10, time.UTC),
			},
			want: &RowMetric{
				NameStr: "testmetric",
				Fields:  []*lp.Field{{Key: "field1", Value: int64(1)}, {Key: "field2", Value: int64(2)}},
				Tags:    []*lp.Tag{{Key: "tag1", Value: "1"}, {Key: "tag2", Value: "2"}},
				TS:      time.Date(2001, 1, 1, 1, 0, 0, 10, time.UTC),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewRowMetric(tt.args.fields, tt.args.name, tt.args.tags, tt.args.ts); !cmp.Equal(got, tt.want) { //!reflect.DeepEqual(got, tt.want) {
				t.Errorf("Diff: %s", cmp.Diff(got, tt.want))
			}
		})
	}
}
