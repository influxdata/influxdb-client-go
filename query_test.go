package influxdb

import (
	"bytes"
	"encoding/csv"
	"io/ioutil"
	"testing"
	"time"

	cmp "github.com/google/go-cmp/cmp"
)

func mustParseTime(s string) time.Time {
	t, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		panic(err)
	}
	return t
}

func TestQueryCSVResult_Unmarshal(t *testing.T) {
	type fields struct {
		//ReadCloser  io.ReadCloser
		//csvReader   *csv.Reader
		Row         []string
		ColNames    []string
		dataTypes   []string
		group       []bool
		defaultVals []string
		Err         error
	}
	type NotAnInt int
	tests := []struct {
		name   string
		fields fields

		expected interface{}
		arg      interface{}
		wantErr  bool
	}{{
		name: "struct{Fred string}",
		arg: &struct {
			Fred                      ***string   `flux:"_measurement"`
			Stop                      time.Time   `flux:"_stop"`
			StopPtr                   *time.Time  `flux:"_stop"`
			FredPtr                   *string     `flux:"_stop"`
			NotAnInt                  NotAnInt    `flux:"_value"`
			ShouldBeNil               *int        `flux:"ktest2"`
			NonExistantField          int         `flux:"fakefield"`
			NonExistantInterfaceField interface{} `flux:"fakefield2"`
		}{},
		expected: &struct {
			Fred                      ***string   `flux:"_measurement"`
			Stop                      time.Time   `flux:"_stop"`
			StopPtr                   *time.Time  `flux:"_stop"`
			FredPtr                   *string     `flux:"_stop"`
			NotAnInt                  NotAnInt    `flux:"_value"`
			ShouldBeNil               *int        `flux:"ktest2"`
			NonExistantField          int         `flux:"fakefield"`
			NonExistantInterfaceField interface{} `flux:"fakefield2"`
		}{
			Fred:        func() ***string { a := "tes0"; b := &a; c := &b; return &c }(),
			Stop:        mustParseTime("2019-06-05T21:21:04.1818586Z"),
			StopPtr:     func() *time.Time { x := mustParseTime("2019-06-05T21:21:04.1818586Z"); return &x }(),
			FredPtr:     func() *string { x := "2019-06-05T21:21:04.1818586Z"; return &x }(),
			NotAnInt:    5,
			ShouldBeNil: nil,
		},
		wantErr: false,
	}, {
		name:    "map[string]string",
		arg:     make(map[string]string),
		wantErr: false,
		expected: map[string]string{
			"_field":         "",
			"_measurement":   "tes0",
			"_start":         "2019-04-25T05:21:04.1818586Z",
			"_stop":          "2019-06-05T21:21:04.1818586Z",
			"_time":          "2019-06-05T21:20:34.142165Z",
			"_value":         "5",
			"ktest1":         "k-test1",
			"ktest2":         "",
			"ktest2,k-test3": "",
			"ktest3":         "k-test3",
			"result":         "_result",
			"table":          "0"},
	}, {
		name:    "map[string]string",
		arg:     make(map[string]interface{}),
		wantErr: false,
		expected: map[string]interface{}{
			"_field":         "",
			"_measurement":   "tes0",
			"_start":         mustParseTime("2019-04-25T05:21:04.1818586Z"),
			"_stop":          mustParseTime("2019-06-05T21:21:04.1818586Z"),
			"_time":          mustParseTime("2019-06-05T21:20:34.142165Z"),
			"_value":         int64(5),
			"ktest1":         "k-test1",
			"ktest2":         "",
			"ktest2,k-test3": "",
			"ktest3":         "k-test3",
			"result":         "_result",
			"table":          int64(0)},
	},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := bytes.NewBufferString(`#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string,string,string,string,string
#group,false,false,true,true,false,false,false,false,false,false,false,false,false,false
#default,_result,,,,,,,,,,,,,
,result,table,_start,_stop,_time,_value,_measurement,ktest1,ktest2,_field,"ktest2,k-test3",_field,ktest3,_field
,,0,2019-04-25T05:21:04.1818586Z,2019-06-05T21:21:04.1818586Z,2019-06-05T21:20:34.142164001Z,5,tes0,k-test1,,ftest1,,,k-test3,
,,0,2019-04-25T05:21:04.1818586Z,2019-06-05T21:21:04.1818586Z,2019-06-05T21:20:34.142165Z,5,tes0,k-test1,,ftest1,,,k-test3,
`)
			q := &QueryCSVResult{
				ReadCloser:  ioutil.NopCloser(buf),
				csvReader:   csv.NewReader(ioutil.NopCloser(buf)),
				Row:         tt.fields.Row,
				ColNames:    tt.fields.ColNames,
				dataTypes:   tt.fields.dataTypes,
				group:       tt.fields.group,
				defaultVals: tt.fields.defaultVals,
				Err:         tt.fields.Err,
			}
			q.Next()
			q.Next()
			if err := q.Unmarshal(tt.arg); (err != nil) != tt.wantErr {
				t.Fatalf("QueryCSVResult.Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !cmp.Equal(tt.arg, tt.expected) {
				t.Fatalf("expected %v got %v, diff: %s", tt.expected, tt.arg, cmp.Diff(tt.expected, tt.arg))
			}
		})
	}
}
