package influxclient

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQuery(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/csv")
		w.WriteHeader(200)
		w.Write([]byte(`#group,false,false,true,true,false,true,true,true,false,false,false
#default,_result,,,,,,,,,,
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,string,string,string,double,double,double
,result,table,_start,_stop,_time,deviceId,location,sensor,air_hum,air_press,air_temp
,,0,2021-10-19T14:39:57.464357168Z,2021-10-19T14:54:57.464357168Z,2021-10-19T14:40:21.833564544Z,2663346492,saman-home-room-0-1,BME280,48.8,1022.28,22.73
,,0,2021-10-19T14:39:57.464357168Z,2021-10-19T14:54:57.464357168Z,2021-10-19T14:41:29.840881203Z,2663346492,saman-home-room-0-1,BME280,49.2,1022.34,22.7`))
	}))
	defer ts.Close()
	client, err := New(Params{ServerURL: ts.URL})
	require.NoError(t, err)
	res, err := client.Query(context.Background(), "query", nil)
	require.NoError(t, err)
	defer res.Close()
	i := 0
	for res.NextSection() && res.Err() == nil {
		for res.NextRow() {
			i++
		}
	}
	require.NoError(t, res.Err())
	assert.Equal(t, 2, i)

}

func TestQueryError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(400)
		w.Write([]byte(`{"code":"invalid","message":"compilation failed: error at @1:170-1:171: invalid expression @1:167-1:168: |"}`))
	}))
	defer ts.Close()
	client, err := New(Params{ServerURL: ts.URL})
	require.NoError(t, err)
	res, err := client.Query(context.Background(), "query", nil)
	assert.Nil(t, res)
	require.Error(t, err)
	assert.Equal(t, `invalid: compilation failed: error at @1:170-1:171: invalid expression @1:167-1:168: |`, err.Error())
}

func TestQueryParamsTypes(t *testing.T) {
	var i int8 = 1
	var paramsTypeTests = []struct {
		testName    string
		params      interface{}
		expectError string
	}{
		{
			"structWillAllSupportedTypes",
			struct {
				B   bool
				I   int
				I8  int8
				I16 int16
				I32 int32
				I64 int64
				U   uint
				U8  uint8
				U16 uint16
				U32 uint32
				U64 uint64
				F32 float32
				F64 float64
				D   time.Duration
				T   time.Time
			}{},
			"",
		},
		{
			"structWithInvalidFieldEmptyInterface",
			struct {
				F interface{}
			}{},
			"cannot use field 'F' of type 'interface {}' as a query param",
		},
		{
			"structWithFieldAsValidInterfaceValue",
			struct {
				F interface{}
			}{"string"},
			"",
		},
		{
			"structAsPointer",
			&struct {
				S string
			}{"a"},
			"",
		},
		{
			"structWithInvalidFieldAsMap",
			struct {
				M map[string]string
			}{},
			"cannot use field 'M' of type 'map[string]string' as a query param",
		},
		{
			"structWithFieldAsPointer",
			struct {
				P *int8
			}{&i},
			"",
		},
		{
			"mapOfBool",
			map[string]bool{},
			"",
		},
		{
			"mapOfFloat64",
			map[string]float64{},
			"",
		},
		{
			"mapOfString",
			map[string]string{},
			"",
		},
		{
			"mapOfTime",
			map[string]time.Time{},
			"",
		},
		{
			"mapOfInterfaceEmpty",
			map[string]interface{}{},
			"",
		},
		{
			"mapOfInterfaceWithValidValues",
			map[string]interface{}{"s": "s", "t": time.Now()},
			"",
		},
		{
			"mapOfInterfaceWithStructInvalid",
			map[string]interface{}{"s": struct {
				a int
			}{1}},
			"cannot use map value type 'struct { a int }' as a query param",
		},
		{
			"mapOfStructInvalid",
			map[string]struct {
				a int
			}{"a": {1}},
			"cannot use map value type 'struct { a int }' as a query param",
		},
		{
			"mapWithInvalidKey",
			map[int]string{},
			"cannot use map key of type 'int' for query param name",
		},
		{
			"invalidParamsType",
			0,
			"cannot use int as query param",
		},
	}
	for _, test := range paramsTypeTests {
		t.Run(test.testName, func(t *testing.T) {
			err := checkContainerType(test.params, true, "query param")
			if test.expectError != "" {
				require.Error(t, err)
				require.Equal(t, test.expectError, err.Error())
				return
			}
			require.NoError(t, err)
		})
	}
}
