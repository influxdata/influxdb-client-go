package influxclient

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/influxdata/line-protocol/v2/lineprotocol"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncode(t *testing.T) {

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
			error: `encoding error: invalid tag key ""`,
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
			error: `encoding error: invalid field key ""`,
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

			client, err := New(Params{ServerURL: "http://localhost:8086"})
			require.NoError(t, err)
			b, err := encode(ts.s, client.params.WriteParams)
			if ts.error == "" {
				require.NoError(t, err)
				assert.Equal(t, ts.line, string(b))
			} else {
				require.Error(t, err)
				assert.Equal(t, ts.error, err.Error())
			}
		})
	}
}

func genPoints(t *testing.T, count int) []*Point {
	ps := make([]*Point, count)
	ts := time.Now()
	ts.Truncate(time.Second)
	rand.Seed(321)
	for i := range ps {
		p := NewPointWithMeasurement("host")
		p.AddTag("rack", fmt.Sprintf("rack_%2d", i%10))
		p.AddTag("name", fmt.Sprintf("machine_%2d", i))
		p.AddField("temperature", rand.Float64()*80.0)
		p.AddField("disk_free", rand.Float64()*1000.0)
		p.AddField("disk_total", (i/10+1)*1000000)
		p.AddField("mem_total", (i/100+1)*10000000)
		p.AddField("mem_free", rand.Uint64())
		p.Timestamp = ts
		ps[i] = p
		ts = ts.Add(time.Millisecond)
	}
	return ps
}
func points2bytes(t *testing.T, points []*Point) []byte {
	var bytes []byte
	for _, p := range points {
		bs, err := p.MarshalBinary(lineprotocol.Millisecond, nil)
		require.NoError(t, err)
		bytes = append(bytes, bs...)
	}
	return bytes
}

func returnHTTPError(w http.ResponseWriter, code int, message string) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(code)
	_, _ = w.Write([]byte(fmt.Sprintf(`{"code":"invalid", "message":"%s"}`, message)))
}

// compArrays compares arrays
func compArrays(b1 []byte, b2 []byte) int {
	if len(b1) != len(b2) {
		return -1
	}
	for i := 0; i < len(b1); i++ {
		if b1[i] != b2[i] {
			return i
		}
	}
	return 0
}

func TestWriteCorrectUrl(t *testing.T) {
	correctPath := "/path/api/v2/write?bucket=my-bucket&org=my-org&precision=ms"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.EqualValues(t, correctPath, r.URL.String())
		w.WriteHeader(204)
	}))

	params := DefaultWriteParams
	params.Precision = lineprotocol.Millisecond
	c, err := New(Params{
		ServerURL:    ts.URL + "/path/",
		Organization: "my-org",
		WriteParams:  params,
	})
	require.NoError(t, err)
	err = c.Write(context.Background(), "my-bucket", []byte("a f=1"))
	assert.NoError(t, err)
	correctPath = "/path/api/v2/write?bucket=my-bucket&consistency=quorum&org=my-org&precision=ms"
	c.params.WriteParams.Consistency = ConsistencyQuorum
	err = c.Write(context.Background(), "my-bucket", []byte("a f=1"))
	assert.NoError(t, err)

}
func TestWritePointsAndBytes(t *testing.T) {
	points := genPoints(t, 5000)
	byts := points2bytes(t, points)
	reqs := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqs++
		buff, err := io.ReadAll(r.Body)
		if err != nil {
			returnHTTPError(w, http.StatusInternalServerError, fmt.Sprintf("error reading body: %v", err))
			return
		}
		if r := compArrays(byts, buff); r != 0 {
			if r == -1 {
				returnHTTPError(w, http.StatusInternalServerError, fmt.Sprintf("error lens are not equal %d vs %d", len(byts), len(buff)))
			} else {
				returnHTTPError(w, http.StatusInternalServerError, fmt.Sprintf("error bytes are not equal %d", r))
			}
			return
		}
		w.WriteHeader(204)
	}))
	defer ts.Close()
	c, err := New(Params{
		ServerURL: ts.URL,
	})
	c.params.WriteParams.Precision = lineprotocol.Millisecond
	c.params.WriteParams.GzipThreshold = 0
	require.NoError(t, err)
	err = c.Write(context.Background(), "b", byts)
	assert.NoError(t, err)
	assert.Equal(t, 1, reqs)

	err = c.WritePoints(context.Background(), "b", points...)
	assert.NoError(t, err)
	assert.Equal(t, 2, reqs)

	// test error
	err = c.Write(context.Background(), "b", []byte("line"))
	require.Error(t, err)
	assert.Equal(t, 3, reqs)
	assert.Equal(t, "invalid: error lens are not equal 911244 vs 4", err.Error())
}

func TestWriteData(t *testing.T) {
	now := time.Now()
	s := struct {
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
	}
	byts := []byte(fmt.Sprintf("air,device_id=10,sensor=SHT31 humidity=55i,temperature=23.5 %d\n", now.UnixNano()))
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buff, err := io.ReadAll(r.Body)
		if err != nil {
			returnHTTPError(w, http.StatusInternalServerError, fmt.Sprintf("error reading body: %v", err))
			return
		}
		if r := compArrays(byts, buff); r != 0 {
			if r == -1 {
				returnHTTPError(w, http.StatusInternalServerError, fmt.Sprintf("error lens are not equal %d vs %d", len(byts), len(buff)))
			} else {
				returnHTTPError(w, http.StatusInternalServerError, fmt.Sprintf("error bytes are not equal %d", r))
			}
			return
		}
		w.WriteHeader(204)
	}))
	defer ts.Close()
	c, err := New(Params{
		ServerURL: ts.URL,
	})
	c.params.WriteParams.GzipThreshold = 0
	require.NoError(t, err)
	err = c.WriteData(context.Background(), "b", s)
	assert.NoError(t, err)

}

func TestGzip(t *testing.T) {
	points := genPoints(t, 1)
	byts := points2bytes(t, points)
	wasGzip := false
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := r.Body
		if r.Header.Get("Content-Encoding") == "gzip" {
			body, _ = gzip.NewReader(body)
			wasGzip = true
		}
		buff, err := io.ReadAll(body)
		if err != nil {
			returnHTTPError(w, http.StatusInternalServerError, fmt.Sprintf("error reading body: %v", err))
			return
		}
		if r := compArrays(byts, buff); r != 0 {
			if r == -1 {
				returnHTTPError(w, http.StatusInternalServerError, fmt.Sprintf("error lens  are not equal %d vs %d", len(byts), len(buff)))
			} else {
				returnHTTPError(w, http.StatusInternalServerError, fmt.Sprintf("error bytes are not equal %d", r))
			}
			return
		}
		w.WriteHeader(204)
	}))
	defer ts.Close()
	c, err := New(Params{
		ServerURL: ts.URL,
	})
	require.NoError(t, err)
	//Test no gzip on small body
	err = c.Write(context.Background(), "b", byts)
	assert.NoError(t, err)
	assert.False(t, wasGzip)
	// Test gzip on larger body
	points = genPoints(t, 100)
	byts = points2bytes(t, points)
	err = c.Write(context.Background(), "b", byts)
	assert.NoError(t, err)
	assert.True(t, wasGzip)
	// Test disable gzipping
	wasGzip = false
	c.params.WriteParams.GzipThreshold = 0
	err = c.Write(context.Background(), "b", byts)
	assert.NoError(t, err)
	assert.False(t, wasGzip)
}
