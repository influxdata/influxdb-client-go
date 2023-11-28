package api

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/influxdata/influxdb-client-go/v2/api/write"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	lp "github.com/influxdata/line-protocol"
)

func TestDataToPoint(t *testing.T) {
	pointToLine := func(point *write.Point) string {
		var buffer bytes.Buffer
		e := lp.NewEncoder(&buffer)
		e.SetFieldTypeSupport(lp.UintSupport)
		e.FailOnFieldErr(true)
		_, err := e.Encode(point)
		if err != nil {
			panic(err)
		}
		return buffer.String()
	}
	now := time.Now()
	tests := []struct {
		name  string
		s     interface{}
		line  string
		error string
	}{{
		name: "test normal structure",
		s: struct {
			Measurement string    `lp:"measurement"`
			Sensor      string    `lp:"tag,sensor"`
			ID          string    `lp:"tag,device_id"`
			Temp        float64   `lp:"field,temperature"`
			Hum         int       `lp:"field,humidity"`
			Time        time.Time `lp:"timestamp"`
			Description string    `lp:"-"`
		}{
			"air",
			"SHT31",
			"10",
			23.5,
			55,
			now,
			"Room temp",
		},
		line: fmt.Sprintf("air,device_id=10,sensor=SHT31 humidity=55i,temperature=23.5 %d\n", now.UnixNano()),
	},
		{
			name: "test pointer to normal structure",
			s: &struct {
				Measurement string    `lp:"measurement"`
				Sensor      string    `lp:"tag,sensor"`
				ID          string    `lp:"tag,device_id"`
				Temp        float64   `lp:"field,temperature"`
				Hum         int       `lp:"field,humidity"`
				Time        time.Time `lp:"timestamp"`
				Description string    `lp:"-"`
			}{
				"air",
				"SHT31",
				"10",
				23.5,
				55,
				now,
				"Room temp",
			},
			line: fmt.Sprintf("air,device_id=10,sensor=SHT31 humidity=55i,temperature=23.5 %d\n", now.UnixNano()),
		}, {
			name: "test no tag, no timestamp",
			s: &struct {
				Measurement string  `lp:"measurement"`
				Temp        float64 `lp:"field,temperature"`
			}{
				"air",
				23.5,
			},
			line: "air temperature=23.5\n",
		},
		{
			name: "test default struct field name",
			s: &struct {
				Measurement string  `lp:"measurement"`
				Sensor      string  `lp:"tag"`
				Temp        float64 `lp:"field"`
			}{
				"air",
				"SHT31",
				23.5,
			},
			line: "air,Sensor=SHT31 Temp=23.5\n",
		},
		{
			name: "test missing struct field tag name",
			s: &struct {
				Measurement string  `lp:"measurement"`
				Sensor      string  `lp:"tag,"`
				Temp        float64 `lp:"field"`
			}{
				"air",
				"SHT31",
				23.5,
			},
			error: `cannot use field 'Sensor': invalid lp tag name ""`,
		},
		{
			name: "test missing struct field field name",
			s: &struct {
				Measurement string  `lp:"measurement"`
				Temp        float64 `lp:"field,"`
			}{
				"air",
				23.5,
			},
			error: `cannot use field 'Temp': invalid lp field name ""`,
		},
		{
			name: "test missing measurement",
			s: &struct {
				Measurement string  `lp:"tag"`
				Sensor      string  `lp:"tag"`
				Temp        float64 `lp:"field"`
			}{
				"air",
				"SHT31",
				23.5,
			},
			error: `no struct field with tag 'measurement'`,
		},
		{
			name: "test no field",
			s: &struct {
				Measurement string  `lp:"measurement"`
				Sensor      string  `lp:"tag"`
				Temp        float64 `lp:"tag"`
			}{
				"air",
				"SHT31",
				23.5,
			},
			error: `no struct field with tag 'field'`,
		},
		{
			name: "test double measurement",
			s: &struct {
				Measurement string  `lp:"measurement"`
				Sensor      string  `lp:"measurement"`
				Temp        float64 `lp:"field,a"`
				Hum         float64 `lp:"field,a"`
			}{
				"air",
				"SHT31",
				23.5,
				43.1,
			},
			error: `multiple measurement fields`,
		},
		{
			name: "test multiple tag attributes",
			s: &struct {
				Measurement string  `lp:"measurement"`
				Sensor      string  `lp:"tag,a,a"`
				Temp        float64 `lp:"field,a"`
				Hum         float64 `lp:"field,a"`
			}{
				"air",
				"SHT31",
				23.5,
				43.1,
			},
			error: `multiple tag attributes are not supported`,
		},
		{
			name: "test wrong timestamp type",
			s: &struct {
				Measurement string  `lp:"measurement"`
				Sensor      string  `lp:"tag,sensor"`
				Temp        float64 `lp:"field,a"`
				Hum         float64 `lp:"timestamp"`
			}{
				"air",
				"SHT31",
				23.5,
				43.1,
			},
			error: `cannot use field 'Hum' as a timestamp`,
		},
		{
			name: "test map",
			s: map[string]interface{}{
				"measurement": "air",
				"sensor":      "SHT31",
				"temp":        23.5,
			},
			error: `cannot use map[string]interface {} as point`,
		},
	}
	for _, ts := range tests {
		t.Run(ts.name, func(t *testing.T) {
			point, err := DataToPoint(ts.s)
			if ts.error == "" {
				require.NoError(t, err)
				assert.Equal(t, ts.line, pointToLine(point))
			} else {
				require.Error(t, err)
				assert.Equal(t, ts.error, err.Error())
			}
		})
	}
}
