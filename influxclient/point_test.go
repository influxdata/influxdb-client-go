package influxclient

import (
	"fmt"
	"math"
	"reflect"
	"testing"
	"time"

	"github.com/influxdata/line-protocol/v2/lineprotocol"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type ia int

type st struct {
	d float64
	b bool
}

func (s st) String() string {
	return fmt.Sprintf("%.2f d %v", s.d, s.b)
}

func TestConvert(t *testing.T) {
	obj := []struct {
		val       interface{}
		targetVal interface{}
	}{
		{int(-5), int64(-5)},
		{int8(5), int64(5)},
		{int16(-51), int64(-51)},
		{int32(5), int64(5)},
		{int64(555), int64(555)},
		{uint(5), uint64(5)},
		{uint8(55), uint64(55)},
		{uint16(51), uint64(51)},
		{uint32(555), uint64(555)},
		{uint64(5555), uint64(5555)},
		{"a", "a"},
		{true, true},
		{float32(1.2), float64(1.2)},
		{float64(2.2), float64(2.2)},
		{ia(4), "4"},
		{[]string{"a", "b"}, "[a b]"},
		{map[int]string{1: "a", 2: "b"}, "map[1:a 2:b]"},
		{struct {
			a int
			s string
		}{0, "x"}, "{0 x}"},
		{st{1.22, true}, "1.22 d true"},
		{[]byte("test"), "test"},
		{time.Date(2022, 12, 13, 14, 15, 16, 0, time.UTC), "2022-12-13T14:15:16Z"},
		{12*time.Hour + 11*time.Minute + 10*time.Second, "12h11m10s"},
	}
	for _, tv := range obj {
		t.Run(reflect.TypeOf(tv.val).String(), func(t *testing.T) {
			v := convertField(tv.val)
			assert.Equal(t, reflect.TypeOf(tv.targetVal), reflect.TypeOf(v))
			if f, ok := tv.targetVal.(float64); ok {
				val := reflect.ValueOf(tv.val)
				ft := reflect.TypeOf(float64(0))
				assert.True(t, val.Type().ConvertibleTo(ft))
				valf := val.Convert(ft)
				assert.True(t, math.Abs(f-valf.Float()) < 1e-6)
			} else {
				assert.EqualValues(t, tv.targetVal, v)
			}
		})
	}
}

func TestPoint(t *testing.T) {
	p := NewPoint(
		"test",
		map[string]string{
			"id":        "10ad=",
			"ven=dor":   "AWS",
			`host"name`: `ho\st "a"`,
			`x\" x`:     "a b",
		},
		map[string]interface{}{
			"float64":  80.1234567,
			"float32":  float32(80.0),
			"int":      -1234567890,
			"int8":     int8(-34),
			"int16":    int16(-3456),
			"int32":    int32(-34567),
			"int64":    int64(-1234567890),
			"uint":     uint(12345677890),
			"uint8":    uint8(34),
			"uint16":   uint16(3456),
			"uint32":   uint32(34578),
			"uint 64":  uint64(41234567890),
			"bo\\ol":   false,
			`"string"`: `six, "seven", eight`,
			"stri=ng":  `six=seven\, eight`,
			"time":     time.Date(2020, time.March, 20, 10, 30, 23, 123456789, time.UTC),
			"duration": 4*time.Hour + 24*time.Minute + 3*time.Second,
		},
		time.Unix(60, 70))
	// Test duplicate tag and duplicate field
	p.AddTag("ven=dor", "GCP").AddField("uint32", uint32(345780))

	line, err := p.MarshalBinary(lineprotocol.Nanosecond)
	require.NoError(t, err)
	assert.EqualValues(t, `test,host"name=ho\st\ "a",id=10ad\=,ven\=dor=GCP,x\"\ x=a\ b "string"="six, \"seven\", eight",bo\ol=false,duration="4h24m3s",float32=80,float64=80.1234567,int=-1234567890i,int16=-3456i,int32=-34567i,int64=-1234567890i,int8=-34i,stri\=ng="six=seven\\, eight",time="2020-03-20T10:30:23.123456789Z",uint=12345677890u,uint\ 64=41234567890u,uint16=3456u,uint32=345780u,uint8=34u 60000000070`+"\n", string(line))
}
