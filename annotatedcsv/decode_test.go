package annotatedcsv

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDecodeStruct(t *testing.T) {
	csvTable := `#datatype,string,unsignedLong,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339Nano,duration,string,long,base64Binary,boolean
#group,false,false,true,true,false,false,true,true,true,true
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,took,_field,index,note,b
,,0,2020-02-17T22:19:49.747562847Z,2020-02-18T22:19:49.747562847Z,2020-02-18T10:34:08.135814545Z,32m,f,-1,ZGF0YWluYmFzZTY0,true
,,0,2020-02-17T22:19:49.747562847Z,2020-02-18T22:19:49.747562847Z,2020-02-18T22:08:44.850214724Z,1h23m4s,f,1,eHh4eHhjY2NjY2NkZGRkZA==,false

#datatype,long,double,dateTime,string
#default,,,,
,index,score,time,name
,0,3.3,2021-02-18T10:34:08.135814545Z,Thomas
,1,5.1,2021-02-18T22:08:44.850214724Z,John

`
	reader := strings.NewReader(csvTable)
	res := NewReader(reader)
	require.True(t, res.NextSection())
	require.NoError(t, res.Err())
	require.True(t, res.NextRow())
	require.NoError(t, res.Err())

	s := &struct {
		Table uint          `csv:"table"`
		Start time.Time     `csv:"_start"`
		Stop  time.Time     `csv:"_stop"`
		Time  time.Time     `csv:"_time"`
		Took  time.Duration `csv:"took"`
		Field string        `csv:"_field"`
		Index int           `csv:"index"`
		Note  []byte        `csv:"note"`
		Tag   bool          `csv:"b"`
	}{}
	err := res.Decode(s)
	require.NoError(t, err)
	assert.Equal(t, &struct {
		Table uint          `csv:"table"`
		Start time.Time     `csv:"_start"`
		Stop  time.Time     `csv:"_stop"`
		Time  time.Time     `csv:"_time"`
		Took  time.Duration `csv:"took"`
		Field string        `csv:"_field"`
		Index int           `csv:"index"`
		Note  []byte        `csv:"note"`
		Tag   bool          `csv:"b"`
	}{
		Table: 0,
		Start: mustParseTime("2020-02-17T22:19:49.747562847Z"),
		Stop:  mustParseTime("2020-02-18T22:19:49.747562847Z"),
		Time:  mustParseTime("2020-02-18T10:34:08.135814545Z"),
		Took:  time.Minute * 32,
		Field: "f",
		Index: -1,
		Note:  []byte("datainbase64"),
		Tag:   true,
	}, s)

	s2 := &struct {
		Table interface{} `csv:"table"`
		Start interface{} `csv:"_start"`
		Stop  interface{} `csv:"_stop"`
		Time  interface{} `csv:"_time"`
		Took  interface{} `csv:"took"`
		Field interface{} `csv:"_field"`
		Index interface{} `csv:"index"`
		Note  interface{} `csv:"note"`
		Tag   interface{} `csv:"b"`
	}{}
	err = res.Decode(s2)
	require.NoError(t, err)
	assert.Equal(t, &struct {
		Table interface{} `csv:"table"`
		Start interface{} `csv:"_start"`
		Stop  interface{} `csv:"_stop"`
		Time  interface{} `csv:"_time"`
		Took  interface{} `csv:"took"`
		Field interface{} `csv:"_field"`
		Index interface{} `csv:"index"`
		Note  interface{} `csv:"note"`
		Tag   interface{} `csv:"b"`
	}{
		Table: uint64(0),
		Start: mustParseTime("2020-02-17T22:19:49.747562847Z"),
		Stop:  mustParseTime("2020-02-18T22:19:49.747562847Z"),
		Time:  mustParseTime("2020-02-18T10:34:08.135814545Z"),
		Took:  time.Minute * 32,
		Field: "f",
		Index: int64(-1),
		Note:  []byte("datainbase64"),
		Tag:   true,
	}, s2)

	s3 := &struct {
		Table string `csv:"table"`
		Start string `csv:"_start"`
		Stop  string `csv:"_stop"`
		Time  string `csv:"_time"`
		Took  string `csv:"took"`
		Field string `csv:"_field"`
		Index string `csv:"index"`
		Note  string `csv:"note"`
		Tag   string `csv:"b"`
	}{}
	err = res.Decode(s3)
	require.NoError(t, err)
	assert.Equal(t, &struct {
		Table string `csv:"table"`
		Start string `csv:"_start"`
		Stop  string `csv:"_stop"`
		Time  string `csv:"_time"`
		Took  string `csv:"took"`
		Field string `csv:"_field"`
		Index string `csv:"index"`
		Note  string `csv:"note"`
		Tag   string `csv:"b"`
	}{
		Table: "0",
		Start: "2020-02-17T22:19:49.747562847Z",
		Stop:  "2020-02-18T22:19:49.747562847Z",
		Time:  "2020-02-18T10:34:08.135814545Z",
		Took:  "32m",
		Field: "f",
		Index: "-1",
		Note:  "ZGF0YWluYmFzZTY0",
		Tag:   "true",
	}, s3)

	require.True(t, res.NextRow(), res.Err())
	require.NoError(t, res.Err())
	err = res.Decode(s)
	require.NoError(t, err)

	assert.Equal(t, &struct {
		Table uint          `csv:"table"`
		Start time.Time     `csv:"_start"`
		Stop  time.Time     `csv:"_stop"`
		Time  time.Time     `csv:"_time"`
		Took  time.Duration `csv:"took"`
		Field string        `csv:"_field"`
		Index int           `csv:"index"`
		Note  []byte        `csv:"note"`
		Tag   bool          `csv:"b"`
	}{
		Table: 0,
		Start: mustParseTime("2020-02-17T22:19:49.747562847Z"),
		Stop:  mustParseTime("2020-02-18T22:19:49.747562847Z"),
		Time:  mustParseTime("2020-02-18T22:08:44.850214724Z"),
		Took:  time.Hour + 23*time.Minute + 4*time.Second,
		Field: "f",
		Index: 1,
		Note:  []byte("xxxxxccccccddddd"),
		Tag:   false,
	}, s)

	err = res.Decode(s2)
	require.NoError(t, err)

	assert.Equal(t, &struct {
		Table interface{} `csv:"table"`
		Start interface{} `csv:"_start"`
		Stop  interface{} `csv:"_stop"`
		Time  interface{} `csv:"_time"`
		Took  interface{} `csv:"took"`
		Field interface{} `csv:"_field"`
		Index interface{} `csv:"index"`
		Note  interface{} `csv:"note"`
		Tag   interface{} `csv:"b"`
	}{
		Table: uint64(0),
		Start: mustParseTime("2020-02-17T22:19:49.747562847Z"),
		Stop:  mustParseTime("2020-02-18T22:19:49.747562847Z"),
		Time:  mustParseTime("2020-02-18T22:08:44.850214724Z"),
		Took:  time.Hour + 23*time.Minute + 4*time.Second,
		Field: "f",
		Index: int64(1),
		Note:  []byte("xxxxxccccccddddd"),
		Tag:   false,
	}, s2)

	err = res.Decode(s3)
	require.NoError(t, err)

	assert.Equal(t, &struct {
		Table string `csv:"table"`
		Start string `csv:"_start"`
		Stop  string `csv:"_stop"`
		Time  string `csv:"_time"`
		Took  string `csv:"took"`
		Field string `csv:"_field"`
		Index string `csv:"index"`
		Note  string `csv:"note"`
		Tag   string `csv:"b"`
	}{
		Table: "0",
		Start: "2020-02-17T22:19:49.747562847Z",
		Stop:  "2020-02-18T22:19:49.747562847Z",
		Time:  "2020-02-18T22:08:44.850214724Z",
		Took:  "1h23m4s",
		Field: "f",
		Index: "1",
		Note:  "eHh4eHhjY2NjY2NkZGRkZA==",
		Tag:   "false",
	}, s3)

	require.True(t, res.NextSection())
	require.NoError(t, res.Err())
	require.True(t, res.NextRow())
	require.NoError(t, res.Err())

	sn := &struct {
		Index int64     `csv:"index"`
		Time  time.Time `csv:"time"`
		Name  string    `csv:"name"`
		Score float64   `csv:"score"`
		Sum   float64
	}{}

	err = res.Decode(sn)
	require.NoError(t, err)

	assert.Equal(t, &struct {
		Index int64     `csv:"index"`
		Time  time.Time `csv:"time"`
		Name  string    `csv:"name"`
		Score float64   `csv:"score"`
		Sum   float64
	}{
		Index: 0,
		Time:  mustParseTime("2021-02-18T10:34:08.135814545Z"),
		Score: 3.3,
		Name:  "Thomas",
		Sum:   0,
	}, sn)

	sn2 := &struct {
		Index interface{} `csv:"index"`
		Time  interface{} `csv:"time"`
		Name  interface{} `csv:"name"`
		Score interface{} `csv:"score"`
		Sum   interface{}
	}{}

	err = res.Decode(sn2)
	require.NoError(t, err)

	assert.Equal(t, &struct {
		Index interface{} `csv:"index"`
		Time  interface{} `csv:"time"`
		Name  interface{} `csv:"name"`
		Score interface{} `csv:"score"`
		Sum   interface{}
	}{
		Index: int64(0),
		Time:  mustParseTime("2021-02-18T10:34:08.135814545Z"),
		Score: 3.3,
		Name:  "Thomas",
		Sum:   nil,
	}, sn2)

	sn3 := &struct {
		Index string `csv:"index"`
		Time  string `csv:"time"`
		Name  string `csv:"name"`
		Score string `csv:"score"`
		Sum   string
	}{}

	err = res.Decode(sn3)
	require.NoError(t, err)

	assert.Equal(t, &struct {
		Index string `csv:"index"`
		Time  string `csv:"time"`
		Name  string `csv:"name"`
		Score string `csv:"score"`
		Sum   string
	}{
		Index: "0",
		Time:  "2021-02-18T10:34:08.135814545Z",
		Score: "3.3",
		Name:  "Thomas",
		Sum:   "",
	}, sn3)

	require.True(t, res.NextRow())
	require.NoError(t, res.Err())
	err = res.Decode(sn)
	require.NoError(t, err)

	assert.Equal(t, &struct {
		Index int64     `csv:"index"`
		Time  time.Time `csv:"time"`
		Name  string    `csv:"name"`
		Score float64   `csv:"score"`
		Sum   float64
	}{
		Index: 1,
		Time:  mustParseTime("2021-02-18T22:08:44.850214724Z"),
		Score: 5.1,
		Name:  "John",
		Sum:   0,
	}, sn)

	err = res.Decode(sn2)
	require.NoError(t, err)
	assert.Equal(t, &struct {
		Index interface{} `csv:"index"`
		Time  interface{} `csv:"time"`
		Name  interface{} `csv:"name"`
		Score interface{} `csv:"score"`
		Sum   interface{}
	}{
		Index: int64(1),
		Time:  mustParseTime("2021-02-18T22:08:44.850214724Z"),
		Score: 5.1,
		Name:  "John",
		Sum:   nil,
	}, sn2)

	err = res.Decode(sn3)
	require.NoError(t, err)

	assert.Equal(t, &struct {
		Index string `csv:"index"`
		Time  string `csv:"time"`
		Name  string `csv:"name"`
		Score string `csv:"score"`
		Sum   string
	}{
		Index: "1",
		Time:  "2021-02-18T22:08:44.850214724Z",
		Score: "5.1",
		Name:  "John",
		Sum:   "",
	}, sn3)

}

func TestDecodeStructSkipField(t *testing.T) {
	csvTable := `#datatype,long,double,dateTime:RFC3339Nano,string
#default,,,,
,Index,Score,Time,Name
,0,3.3,2021-02-18T10:34:08.135814545Z,Thomas
,1,5.1,2021-02-18T22:08:44.850214724Z,John

`
	reader := strings.NewReader(csvTable)
	res := NewReader(reader)
	require.True(t, res.NextSection())
	require.NoError(t, res.Err())
	require.True(t, res.NextRow())
	require.NoError(t, res.Err())

	// Test decode with skipping field
	s := &struct {
		Index int64
		Time  time.Time
		Name  string
		Score int `csv:"-"`
	}{}

	err := res.Decode(s)
	require.NoError(t, err)

	assert.Equal(t, &struct {
		Index int64
		Time  time.Time
		Name  string
		Score int `csv:"-"`
	}{
		Index: 0,
		Time:  mustParseTime("2021-02-18T10:34:08.135814545Z"),
		Score: 0,
		Name:  "Thomas",
	}, s)
}

func TestDecodeStructNoTag(t *testing.T) {
	csvTable := `#datatype,long,double,dateTime:RFC3339Nano,string
#default,,,,
,Index,Score,Time,Name
,0,3.3,2021-02-18T10:34:08.135814545Z,Thomas
,1,5.1,2021-02-18T22:08:44.850214724Z,John

`
	reader := strings.NewReader(csvTable)
	res := NewReader(reader)
	require.True(t, res.NextSection())
	require.NoError(t, res.Err())
	require.True(t, res.NextRow())
	require.NoError(t, res.Err())

	// Test decode in struct no tag
	s := &struct {
		Index int64
		Time  time.Time
		Name  string
		Score float64
	}{}

	err := res.Decode(s)
	require.NoError(t, err)

	assert.Equal(t, &struct {
		Index int64
		Time  time.Time
		Name  string
		Score float64
	}{
		Index: 0,
		Time:  mustParseTime("2021-02-18T10:34:08.135814545Z"),
		Score: 3.3,
		Name:  "Thomas",
	}, s)
}

func TestDecodeStructNoMatchedFields(t *testing.T) {
	csvTable := `#datatype,long,double,dateTime:RFC3339Nano,string
#default,,,,
,index,score,time,name
,0,3.3,2021-02-18T10:34:08.135814545Z,Thomas
,1,5.1,2021-02-18T22:08:44.850214724Z,John

`
	reader := strings.NewReader(csvTable)
	res := NewReader(reader)
	require.True(t, res.NextSection())
	require.NoError(t, res.Err())
	require.True(t, res.NextRow())
	require.NoError(t, res.Err())

	s := &struct {
		Index int64
		Time  time.Time
		Name  string
		Score float64
	}{}

	err := res.Decode(s)
	require.NoError(t, err)

	assert.Equal(t, &struct {
		Index int64
		Time  time.Time
		Name  string
		Score float64
	}{
		Index: 0,
		Time:  time.Time{},
		Score: 0.0,
		Name:  "",
	}, s)

	// Test decode in struct no matching tag
	sn := &struct {
		Index int64     `csv:"Index"`
		Time  time.Time `csv:"Time"`
		Name  string    `csv:"Name"`
		Score float64   `csv:"Score"`
		Sum   float64
	}{}

	err = res.Decode(sn)
	require.NoError(t, err)

	assert.Equal(t, &struct {
		Index int64     `csv:"Index"`
		Time  time.Time `csv:"Time"`
		Name  string    `csv:"Name"`
		Score float64   `csv:"Score"`
		Sum   float64
	}{
		Index: 0,
		Time:  time.Time{},
		Score: 0.0,
		Name:  "",
	}, sn)
}
func TestDecodeStructTagAttribute(t *testing.T) {
	reader := strings.NewReader(createSimpleCSV("score", "long", "1"))
	res := NewReader(reader)

	require.True(t, res.NextSection())
	require.NoError(t, res.Err())
	require.True(t, res.NextRow())
	require.NoError(t, res.Err())

	// Test decode with tag attribute
	s := &struct {
		Score int `csv:"score,omitempty"`
	}{}

	err := res.Decode(s)
	require.Error(t, err)
	require.Equal(t, "tag attributes are not supported", err.Error())

	// Test decode with empty tag attribute
	s2 := &struct {
		Score int `csv:"score,"`
	}{}

	err = res.Decode(s2)
	require.NoError(t, err)
	require.Equal(t, 1, s2.Score)
}

func TestDecodeInvalidType(t *testing.T) {
	reader := strings.NewReader(createSimpleCSV("test", "long", "1"))
	res := NewReader(reader)

	require.True(t, res.NextSection(), res.Err())
	require.True(t, res.NextRow(), res.Err())

	err := res.Decode(map[string]int{})
	require.Error(t, err)
	assert.Equal(t, "cannot decode into non-pointer type map[string]int", err.Error())

	s := struct {
		Index int64     `csv:"Index"`
		Time  time.Time `csv:"Time"`
		Name  string    `csv:"Name"`
		Score float64   `csv:"Score"`
		Sum   float64
	}{}

	err = res.Decode(s)
	require.Error(t, err)
	assert.True(t, strings.HasPrefix(err.Error(), "cannot decode into non-pointer type struct { Index int64 \"csv:"))

	sp := &s
	sp = nil
	err = res.Decode(sp)
	require.Error(t, err)
	assert.True(t, strings.HasPrefix(err.Error(), "decode into nil *struct { Index"))

	s1 := struct {
		Test fmt.Stringer `csv:"test"`
	}{}
	err = res.Decode(&s1)
	require.Error(t, err)
	assert.Equal(t, "cannot convert from column type long to fmt.Stringer", err.Error())

	var r []float64
	err = res.Decode(r)
	require.Error(t, err)
	assert.Equal(t, "cannot decode into non-pointer type []float64", err.Error())

	err = res.Decode(&r)
	require.Error(t, err)
	assert.Equal(t, "cannot decode into *[]float64", err.Error())

	var f float64
	err = res.Decode(&f)
	require.Error(t, err)
	assert.Equal(t, "cannot decode into *float64", err.Error())

	var a []fmt.Stringer
	err = res.Decode(&a)
	require.Error(t, err)
	assert.Equal(t, "cannot decode into *[]fmt.Stringer", err.Error())

	var p *[]string
	err = res.Decode(p)
	require.Error(t, err)
	assert.Equal(t, "decode into nil *[]string", err.Error())

}

func TestDecodeSlice(t *testing.T) {
	csvTable := `#datatype,string,unsignedLong,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339Nano,duration,string,long,base64Binary,boolean
#group,false,false,true,true,false,false,true,true,true,true
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,took,_field,index,note,b
,,0,2020-02-17T22:19:49.747562847Z,2020-02-18T22:19:49.747562847Z,2020-02-18T10:34:08.135814545Z,32m,f,-1,ZGF0YWluYmFzZTY0,true
,,0,2020-02-17T22:19:49.747562847Z,2020-02-18T22:19:49.747562847Z,2020-02-18T22:08:44.850214724Z,1h23m4s,f,1,eHh4eHhjY2NjY2NkZGRkZA==,false

#datatype,long,double,dateTime,string
#default,,,,
,index,score,time,name
,0,3.3,2021-02-18T10:34:08.135814545Z,Thomas
,1,5.1,2021-02-18T22:08:44.850214724Z,John

`
	reader := strings.NewReader(csvTable)
	res := NewReader(reader)
	require.True(t, res.NextSection())
	require.NoError(t, res.Err())
	require.True(t, res.NextRow())
	require.NoError(t, res.Err())

	var r []interface{}
	require.NoError(t, res.Decode(&r))
	er := []interface{}{
		"_result",
		uint64(0),
		mustParseTime("2020-02-17T22:19:49.747562847Z"),
		mustParseTime("2020-02-18T22:19:49.747562847Z"),
		mustParseTime("2020-02-18T10:34:08.135814545Z"),
		time.Minute * 32,
		"f",
		int64(-1),
		[]byte("datainbase64"),
		true,
	}
	require.Equal(t, er, r)

	var rs []string
	require.NoError(t, res.Decode(&rs))
	er2 := []string{
		"_result",
		"0",
		"2020-02-17T22:19:49.747562847Z",
		"2020-02-18T22:19:49.747562847Z",
		"2020-02-18T10:34:08.135814545Z",
		"32m",
		"f",
		"-1",
		"ZGF0YWluYmFzZTY0",
		"true",
	}
	require.Equal(t, er2, rs)
}

// createSimpleCSV creates annotated CSV string with one column of given params
func createSimpleCSV(fieldName, fieldType, value string) string {
	return fmt.Sprintf(`#datatype,%s
#default,
,%s
,%s`,
		fieldType, fieldName, value)
}

func TestConversionErrors(t *testing.T) {
	var singleRowDecodeTests = []struct {
		testName    string
		csv         string
		value       interface{}
		expectError string
	}{
		{
			testName: "Invalid base64binary",
			csv:      createSimpleCSV("bytes", "base64Binary", `"#"`),
			value: struct {
				B []byte `csv:"bytes"`
			}{},
			expectError: `cannot convert value "#" in column of type "base64Binary" to Go type []uint8 at line 4:1: illegal base64 data at input byte 0`,
		},
		{
			testName: "Invalid duration",
			csv:      createSimpleCSV("dur", "duration", `"#"`),
			value: struct {
				D time.Duration `csv:"dur"`
			}{},
			expectError: `cannot convert value "#" in column of type "duration" to Go type time.Duration at line 4:1: time: invalid duration "#"`,
		},
		{
			testName: "Invalid time",
			csv:      createSimpleCSV("time", "dateTime:RFC3339", `"#"`),
			value: struct {
				T time.Time `csv:"time"`
			}{},
			expectError: `cannot convert value "#" in column of type "dateTime:RFC3339" to Go type time.Time at line 4:1: parsing time "#" as "2006-01-02T15:04:05.999999999Z07:00": cannot parse "#" as "2006"`,
		},
		{
			testName: "Invalid int",
			csv:      createSimpleCSV("index", "long", `"#"`),
			value: struct {
				I int8 `csv:"index"`
			}{},
			expectError: `cannot convert value "#" in column of type "long" to Go type int8 at line 4:1: strconv.ParseInt: parsing "#": invalid syntax`,
		},
		{
			testName: "Overflow int",
			csv:      createSimpleCSV("index", "long", `1600`),
			value: struct {
				I int8 `csv:"index"`
			}{},
			expectError: `cannot convert value "1600" in column of type "long" to Go type int8 at line 4:1: value 1600 overflows type int8`,
		},
		{
			testName: "Invalid uint",
			csv:      createSimpleCSV("index", "unsignedLong", `"#"`),
			value: struct {
				U uint8 `csv:"index"`
			}{},
			expectError: `cannot convert value "#" in column of type "unsignedLong" to Go type uint8 at line 4:1: strconv.ParseUint: parsing "#": invalid syntax`,
		},
		{
			testName: "Overflow uint",
			csv:      createSimpleCSV("index", "unsignedLong", `1600`),
			value: struct {
				U uint8 `csv:"index"`
			}{},
			expectError: `cannot convert value "1600" in column of type "unsignedLong" to Go type uint8 at line 4:1: value 1600 overflows type uint8`,
		},
		{
			testName: "Invalid bool",
			csv:      createSimpleCSV("valid", "boolean", `"#"`),
			value: struct {
				B bool `csv:"valid"`
			}{},
			expectError: `cannot convert value "#" in column of type "boolean" to Go type bool at line 4:1: invalid bool value: "#"`,
		},
		{
			testName: "Invalid float",
			csv:      createSimpleCSV("mem", "double", `"#"`),
			value: struct {
				F float32 `csv:"mem"`
			}{},
			expectError: `cannot convert value "#" in column of type "double" to Go type float32 at line 4:1: strconv.ParseFloat: parsing "#": invalid syntax`,
		},
		{
			testName: "Overflow float",
			csv:      createSimpleCSV("mem", "double", `"1e64"`),
			value: struct {
				F float32 `csv:"mem"`
			}{},
			expectError: `cannot convert value "1e64" in column of type "double" to Go type float32 at line 4:1: value 1e64 overflows type float32`,
		},
		{
			testName: "Invalid column type",
			csv:      createSimpleCSV("note", "strung", `"text"`),
			value: struct {
				S string `csv:"note"`
			}{"text"},
			expectError: "",
		},
	}

	for _, test := range singleRowDecodeTests {
		t.Run(test.testName, func(t *testing.T) {
			decVal := reflect.New(reflect.TypeOf(test.value))
			r := NewReader(strings.NewReader(test.csv))
			require.True(t, r.NextSection())
			require.True(t, r.NextRow())
			err := r.Decode(decVal.Interface())
			if test.expectError != "" {
				require.Error(t, err)
				require.Equal(t, test.expectError, err.Error())
				return
			}
			require.NoError(t, err)
			require.Equal(t, test.value, decVal.Elem().Interface())
		})
	}

}

func TestConversionsMapping(t *testing.T) {
	assert.NotNil(t, conversions[conv{stringCol, fieldKind(reflect.String)}])
	assert.Nil(t, conversions[conv{stringCol, fieldKind(reflect.Bool)}])
	assert.Nil(t, conversions[conv{stringCol, fieldKind(reflect.Int)}])
	assert.Nil(t, conversions[conv{stringCol, fieldKind(reflect.Int8)}])
	assert.Nil(t, conversions[conv{stringCol, fieldKind(reflect.Int16)}])
	assert.Nil(t, conversions[conv{stringCol, fieldKind(reflect.Int32)}])
	assert.Nil(t, conversions[conv{stringCol, fieldKind(reflect.Int64)}])
	assert.Nil(t, conversions[conv{stringCol, fieldKind(reflect.Uint)}])
	assert.Nil(t, conversions[conv{stringCol, fieldKind(reflect.Uint8)}])
	assert.Nil(t, conversions[conv{stringCol, fieldKind(reflect.Uint16)}])
	assert.Nil(t, conversions[conv{stringCol, fieldKind(reflect.Uint32)}])
	assert.Nil(t, conversions[conv{stringCol, fieldKind(reflect.Uint64)}])
	assert.Nil(t, conversions[conv{stringCol, fieldKind(reflect.Float32)}])
	assert.Nil(t, conversions[conv{stringCol, fieldKind(reflect.Float64)}])
	assert.Nil(t, conversions[conv{stringCol, durationKind}])
	assert.Nil(t, conversions[conv{stringCol, timeKind}])
	assert.Nil(t, conversions[conv{stringCol, bytesKind}])

	assert.NotNil(t, conversions[conv{boolCol, fieldKind(reflect.String)}])
	assert.NotNil(t, conversions[conv{boolCol, fieldKind(reflect.Bool)}])
	assert.Nil(t, conversions[conv{boolCol, fieldKind(reflect.Int)}])
	assert.Nil(t, conversions[conv{boolCol, fieldKind(reflect.Int8)}])
	assert.Nil(t, conversions[conv{boolCol, fieldKind(reflect.Int16)}])
	assert.Nil(t, conversions[conv{boolCol, fieldKind(reflect.Int32)}])
	assert.Nil(t, conversions[conv{boolCol, fieldKind(reflect.Int64)}])
	assert.Nil(t, conversions[conv{boolCol, fieldKind(reflect.Uint)}])
	assert.Nil(t, conversions[conv{boolCol, fieldKind(reflect.Uint8)}])
	assert.Nil(t, conversions[conv{boolCol, fieldKind(reflect.Uint16)}])
	assert.Nil(t, conversions[conv{boolCol, fieldKind(reflect.Uint32)}])
	assert.Nil(t, conversions[conv{boolCol, fieldKind(reflect.Uint64)}])
	assert.Nil(t, conversions[conv{boolCol, fieldKind(reflect.Float32)}])
	assert.Nil(t, conversions[conv{boolCol, fieldKind(reflect.Float64)}])
	assert.Nil(t, conversions[conv{boolCol, durationKind}])
	assert.Nil(t, conversions[conv{boolCol, timeKind}])
	assert.Nil(t, conversions[conv{boolCol, bytesKind}])

	assert.NotNil(t, conversions[conv{durationCol, fieldKind(reflect.String)}])
	assert.Nil(t, conversions[conv{durationCol, fieldKind(reflect.Bool)}])
	assert.Nil(t, conversions[conv{durationCol, fieldKind(reflect.Int)}])
	assert.Nil(t, conversions[conv{durationCol, fieldKind(reflect.Int8)}])
	assert.Nil(t, conversions[conv{durationCol, fieldKind(reflect.Int16)}])
	assert.Nil(t, conversions[conv{durationCol, fieldKind(reflect.Int32)}])
	assert.Nil(t, conversions[conv{durationCol, fieldKind(reflect.Int64)}])
	assert.Nil(t, conversions[conv{durationCol, fieldKind(reflect.Uint)}])
	assert.Nil(t, conversions[conv{durationCol, fieldKind(reflect.Uint8)}])
	assert.Nil(t, conversions[conv{durationCol, fieldKind(reflect.Uint16)}])
	assert.Nil(t, conversions[conv{durationCol, fieldKind(reflect.Uint32)}])
	assert.Nil(t, conversions[conv{durationCol, fieldKind(reflect.Uint64)}])
	assert.Nil(t, conversions[conv{durationCol, fieldKind(reflect.Float32)}])
	assert.Nil(t, conversions[conv{durationCol, fieldKind(reflect.Float64)}])
	assert.NotNil(t, conversions[conv{durationCol, durationKind}])
	assert.Nil(t, conversions[conv{durationCol, timeKind}])
	assert.Nil(t, conversions[conv{durationCol, bytesKind}])

	assert.NotNil(t, conversions[conv{longCol, fieldKind(reflect.String)}])
	assert.Nil(t, conversions[conv{longCol, fieldKind(reflect.Bool)}])
	assert.NotNil(t, conversions[conv{longCol, fieldKind(reflect.Int)}])
	assert.NotNil(t, conversions[conv{longCol, fieldKind(reflect.Int8)}])
	assert.NotNil(t, conversions[conv{longCol, fieldKind(reflect.Int16)}])
	assert.NotNil(t, conversions[conv{longCol, fieldKind(reflect.Int32)}])
	assert.NotNil(t, conversions[conv{longCol, fieldKind(reflect.Int64)}])
	assert.Nil(t, conversions[conv{longCol, fieldKind(reflect.Uint)}])
	assert.Nil(t, conversions[conv{longCol, fieldKind(reflect.Uint8)}])
	assert.Nil(t, conversions[conv{longCol, fieldKind(reflect.Uint16)}])
	assert.Nil(t, conversions[conv{longCol, fieldKind(reflect.Uint32)}])
	assert.Nil(t, conversions[conv{longCol, fieldKind(reflect.Uint64)}])
	assert.NotNil(t, conversions[conv{longCol, fieldKind(reflect.Float32)}])
	assert.NotNil(t, conversions[conv{longCol, fieldKind(reflect.Float64)}])
	assert.Nil(t, conversions[conv{longCol, durationKind}])
	assert.Nil(t, conversions[conv{longCol, timeKind}])
	assert.Nil(t, conversions[conv{longCol, bytesKind}])

	assert.NotNil(t, conversions[conv{unsignedLongCol, fieldKind(reflect.String)}])
	assert.Nil(t, conversions[conv{unsignedLongCol, fieldKind(reflect.Bool)}])
	assert.NotNil(t, conversions[conv{unsignedLongCol, fieldKind(reflect.Int)}])
	assert.NotNil(t, conversions[conv{unsignedLongCol, fieldKind(reflect.Int8)}])
	assert.NotNil(t, conversions[conv{unsignedLongCol, fieldKind(reflect.Int16)}])
	assert.NotNil(t, conversions[conv{unsignedLongCol, fieldKind(reflect.Int32)}])
	assert.NotNil(t, conversions[conv{unsignedLongCol, fieldKind(reflect.Int64)}])
	assert.NotNil(t, conversions[conv{unsignedLongCol, fieldKind(reflect.Uint)}])
	assert.NotNil(t, conversions[conv{unsignedLongCol, fieldKind(reflect.Uint8)}])
	assert.NotNil(t, conversions[conv{unsignedLongCol, fieldKind(reflect.Uint16)}])
	assert.NotNil(t, conversions[conv{unsignedLongCol, fieldKind(reflect.Uint32)}])
	assert.NotNil(t, conversions[conv{unsignedLongCol, fieldKind(reflect.Uint64)}])
	assert.NotNil(t, conversions[conv{unsignedLongCol, fieldKind(reflect.Float32)}])
	assert.NotNil(t, conversions[conv{unsignedLongCol, fieldKind(reflect.Float64)}])
	assert.Nil(t, conversions[conv{unsignedLongCol, durationKind}])
	assert.Nil(t, conversions[conv{unsignedLongCol, timeKind}])
	assert.Nil(t, conversions[conv{unsignedLongCol, bytesKind}])

	assert.NotNil(t, conversions[conv{doubleCol, fieldKind(reflect.String)}])
	assert.Nil(t, conversions[conv{doubleCol, fieldKind(reflect.Bool)}])
	assert.Nil(t, conversions[conv{doubleCol, fieldKind(reflect.Int)}])
	assert.Nil(t, conversions[conv{doubleCol, fieldKind(reflect.Int8)}])
	assert.Nil(t, conversions[conv{doubleCol, fieldKind(reflect.Int16)}])
	assert.Nil(t, conversions[conv{doubleCol, fieldKind(reflect.Int32)}])
	assert.Nil(t, conversions[conv{doubleCol, fieldKind(reflect.Int64)}])
	assert.Nil(t, conversions[conv{doubleCol, fieldKind(reflect.Uint)}])
	assert.Nil(t, conversions[conv{doubleCol, fieldKind(reflect.Uint8)}])
	assert.Nil(t, conversions[conv{doubleCol, fieldKind(reflect.Uint16)}])
	assert.Nil(t, conversions[conv{doubleCol, fieldKind(reflect.Uint32)}])
	assert.Nil(t, conversions[conv{doubleCol, fieldKind(reflect.Uint64)}])
	assert.NotNil(t, conversions[conv{doubleCol, fieldKind(reflect.Float32)}])
	assert.NotNil(t, conversions[conv{doubleCol, fieldKind(reflect.Float64)}])
	assert.Nil(t, conversions[conv{doubleCol, durationKind}])
	assert.Nil(t, conversions[conv{doubleCol, timeKind}])
	assert.Nil(t, conversions[conv{doubleCol, bytesKind}])

	assert.NotNil(t, conversions[conv{base64BinaryCol, fieldKind(reflect.String)}])
	assert.Nil(t, conversions[conv{base64BinaryCol, fieldKind(reflect.Bool)}])
	assert.Nil(t, conversions[conv{base64BinaryCol, fieldKind(reflect.Int)}])
	assert.Nil(t, conversions[conv{base64BinaryCol, fieldKind(reflect.Int8)}])
	assert.Nil(t, conversions[conv{base64BinaryCol, fieldKind(reflect.Int16)}])
	assert.Nil(t, conversions[conv{base64BinaryCol, fieldKind(reflect.Int32)}])
	assert.Nil(t, conversions[conv{base64BinaryCol, fieldKind(reflect.Int64)}])
	assert.Nil(t, conversions[conv{base64BinaryCol, fieldKind(reflect.Uint)}])
	assert.Nil(t, conversions[conv{base64BinaryCol, fieldKind(reflect.Uint8)}])
	assert.Nil(t, conversions[conv{base64BinaryCol, fieldKind(reflect.Uint16)}])
	assert.Nil(t, conversions[conv{base64BinaryCol, fieldKind(reflect.Uint32)}])
	assert.Nil(t, conversions[conv{base64BinaryCol, fieldKind(reflect.Uint64)}])
	assert.Nil(t, conversions[conv{base64BinaryCol, fieldKind(reflect.Float32)}])
	assert.Nil(t, conversions[conv{base64BinaryCol, fieldKind(reflect.Float64)}])
	assert.Nil(t, conversions[conv{base64BinaryCol, durationKind}])
	assert.Nil(t, conversions[conv{base64BinaryCol, timeKind}])
	assert.NotNil(t, conversions[conv{base64BinaryCol, bytesKind}])

	assert.NotNil(t, conversions[conv{rfc3339Col, fieldKind(reflect.String)}])
	assert.Nil(t, conversions[conv{rfc3339Col, fieldKind(reflect.Bool)}])
	assert.Nil(t, conversions[conv{rfc3339Col, fieldKind(reflect.Int)}])
	assert.Nil(t, conversions[conv{rfc3339Col, fieldKind(reflect.Int8)}])
	assert.Nil(t, conversions[conv{rfc3339Col, fieldKind(reflect.Int16)}])
	assert.Nil(t, conversions[conv{rfc3339Col, fieldKind(reflect.Int32)}])
	assert.Nil(t, conversions[conv{rfc3339Col, fieldKind(reflect.Int64)}])
	assert.Nil(t, conversions[conv{rfc3339Col, fieldKind(reflect.Uint)}])
	assert.Nil(t, conversions[conv{rfc3339Col, fieldKind(reflect.Uint8)}])
	assert.Nil(t, conversions[conv{rfc3339Col, fieldKind(reflect.Uint16)}])
	assert.Nil(t, conversions[conv{rfc3339Col, fieldKind(reflect.Uint32)}])
	assert.Nil(t, conversions[conv{rfc3339Col, fieldKind(reflect.Uint64)}])
	assert.Nil(t, conversions[conv{rfc3339Col, fieldKind(reflect.Float32)}])
	assert.Nil(t, conversions[conv{rfc3339Col, fieldKind(reflect.Float64)}])
	assert.Nil(t, conversions[conv{rfc3339Col, durationKind}])
	assert.NotNil(t, conversions[conv{rfc3339Col, timeKind}])
	assert.Nil(t, conversions[conv{rfc3339Col, bytesKind}])

	assert.NotNil(t, conversions[conv{rfc3339Col, fieldKind(reflect.String)}])
	assert.Nil(t, conversions[conv{rfc3339Col, fieldKind(reflect.Bool)}])
	assert.Nil(t, conversions[conv{rfc3339Col, fieldKind(reflect.Int)}])
	assert.Nil(t, conversions[conv{rfc3339Col, fieldKind(reflect.Int8)}])
	assert.Nil(t, conversions[conv{rfc3339Col, fieldKind(reflect.Int16)}])
	assert.Nil(t, conversions[conv{rfc3339Col, fieldKind(reflect.Int32)}])
	assert.Nil(t, conversions[conv{rfc3339Col, fieldKind(reflect.Int64)}])
	assert.Nil(t, conversions[conv{rfc3339Col, fieldKind(reflect.Uint)}])
	assert.Nil(t, conversions[conv{rfc3339Col, fieldKind(reflect.Uint8)}])
	assert.Nil(t, conversions[conv{rfc3339Col, fieldKind(reflect.Uint16)}])
	assert.Nil(t, conversions[conv{rfc3339Col, fieldKind(reflect.Uint32)}])
	assert.Nil(t, conversions[conv{rfc3339Col, fieldKind(reflect.Uint64)}])
	assert.Nil(t, conversions[conv{rfc3339Col, fieldKind(reflect.Float32)}])
	assert.Nil(t, conversions[conv{rfc3339Col, fieldKind(reflect.Float64)}])
	assert.Nil(t, conversions[conv{rfc3339Col, durationKind}])
	assert.NotNil(t, conversions[conv{rfc3339Col, timeKind}])
	assert.Nil(t, conversions[conv{rfc3339Col, bytesKind}])
}

// mustParseTime parses s as an RFC3339 timestamp and panics on error.
func mustParseTime(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return t
}
