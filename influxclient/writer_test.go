package influxclient

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var mu sync.Mutex

func TestWrite(t *testing.T) {
	output := make([]byte, 0, 10)
	var w *PointsWriter
	fill := func() {
		for i := 0; i < 10; i++ {
			s := []byte(fmt.Sprintf("a%d\n", i))
			w.Write(s)
		}
	}
	check := func(t *testing.T) {
		mu.Lock()
		l := len(output)
		assert.True(t, l >= 15, "Len %d", l)
		mu.Unlock()
	}
	tests := []struct {
		name    string
		paramsF func() WriteParams
		testF   func(t *testing.T)
	}{
		{
			"Flush when batchsize reached",
			func() WriteParams {
				params := DefaultWriteParams
				params.BatchSize = 5
				return params
			},
			func(t *testing.T) {
				fill()
				waitForCondition(t, 100, func() bool {
					mu.Lock()
					defer mu.Unlock()
					return len(output) >= 10
				})
				check(t)
			},
		},
		{
			"Flush in interval",
			func() WriteParams {
				params := DefaultWriteParams
				params.FlushInterval = 100
				return params
			},
			func(t *testing.T) {
				fill()
				<-time.After(110 * time.Millisecond)
				waitForCondition(t, 100, func() bool {
					mu.Lock()
					defer mu.Unlock()
					return len(output) >= 10
				})
				check(t)
				mu.Lock()
				output = output[:0]
				mu.Unlock()
				fill()
				<-time.After(110 * time.Millisecond)
				waitForCondition(t, 100, func() bool {
					mu.Lock()
					defer mu.Unlock()
					return len(output) >= 10
				})
				check(t)
			},
		},
		{
			"Manual flush",
			func() WriteParams {
				return DefaultWriteParams
			},
			func(t *testing.T) {
				fill()
				w.Flush()
				waitForCondition(t, 100, func() bool {
					mu.Lock()
					defer mu.Unlock()
					return len(output) >= 10
				})
				check(t)
			},
		},
		{
			"Flush  when max bytes reached",
			func() WriteParams {
				params := DefaultWriteParams
				params.MaxBatchBytes = 15
				return params
			},
			func(t *testing.T) {
				fill()
				waitForCondition(t, 100, func() bool {
					mu.Lock()
					defer mu.Unlock()
					return len(output) >= 10
				})
				check(t)
			},
		},
	}
	for _, test := range tests {
		mu.Lock()
		output = output[:0]
		mu.Unlock()
		t.Run(test.name, func(t *testing.T) {
			w = NewPointsWriter(func(ctx context.Context, bucket string, bs []byte) error {
				mu.Lock()
				defer mu.Unlock()
				output = append(output, bs...)
				return nil
			}, "bucket", test.paramsF())

			test.testF(t)

			w.Close()

		})
	}
}

func TestWriteRetriesWithoutErrors(t *testing.T) {
	output := make([]byte, 0, 10)
	requests := 0
	params := DefaultWriteParams
	params.RetryInterval = 100
	params.FlushInterval = 5
	params.RetryJitter = 0
	failures := 0
	success := 0
	params.WriteFailed = func(err error, lines []byte, attempt int, expires time.Time) bool {
		mu.Lock()
		defer mu.Unlock()
		failures++
		return true
	}
	w := NewPointsWriter(func(ctx context.Context, bucket string, bs []byte) error {
		mu.Lock()
		defer mu.Unlock()
		requests++
		if requests%2 == 0 {
			success = success + linesCount(bs)
			output = append(output, bs...)
			return nil
		} else {
			//return &ServerError{StatusCode: 500, Message: "error"}
			return errors.New("error")
		}
	}, "bucket", params)

	w.WritePoints(NewPointWithMeasurement("test").AddField("f", 35))
	waitForCondition(t, 5000, func() bool {
		mu.Lock()
		defer mu.Unlock()
		return success == 1
	})
	assert.EqualValues(t, 1, failures)

	output = output[:0]
	w.WritePoints(&Point{}, //ignored, with warning, will call writeFailed callback
		NewPointWithMeasurement("test"), //ignore, with warning, //ignored, with warning, will call writeFailed callback
		NewPointWithMeasurement("test").AddField("f", 2),
		NewPointWithMeasurement("test").AddField("f", 3),
		NewPointWithMeasurement("test").AddField("f", 4).SetTimestamp(time.Unix(1, 0)),
	)

	waitForCondition(t, 5000, func() bool {
		mu.Lock()
		defer mu.Unlock()
		return success == 4
	})
	assert.EqualValues(t, 4, failures)
	s1 := struct {
		Measurement string `lp:"measurement"`
		F           int    `lp:"field,f"`
	}{
		"air",
		5,
	}
	s2 := struct { //will generate warning
		Measurement string `lp:"measurement"`
	}{
		"air",
	}
	w.WriteData(s1, s2)
	waitForCondition(t, 5000, func() bool {
		mu.Lock()
		defer mu.Unlock()
		return success == 5
	})
	assert.EqualValues(t, 6, failures)

}

func TestWriteRetriesExpiration(t *testing.T) {
	var buff strings.Builder
	log.SetOutput(&buff)
	defer fmt.Println(buff.String())
	params := DefaultWriteParams
	params.RetryInterval = 100
	params.MaxRetryTime = 100
	params.RetryJitter = 0
	params.FlushInterval = 5
	failures := 0
	requests := 0
	success := 0
	params.WriteFailed = func(err error, lines []byte, attempt int, expires time.Time) bool {
		mu.Lock()
		defer mu.Unlock()
		failures++
		return true
	}
	w := NewPointsWriter(func(ctx context.Context, bucket string, bs []byte) error {
		mu.Lock()
		defer mu.Unlock()
		requests++
		if requests%2 == 0 {
			success = success + linesCount(bs)
			return nil
		} else {
			//return &ServerError{StatusCode: 500, Message: "error"}
			return errors.New("error")
		}
	}, "bucket", params)
	w.Write([]byte("test f=35"))
	waitForCondition(t, 5000, func() bool {
		mu.Lock()
		defer mu.Unlock()
		return failures == 2 //one per failed write, second for the expired time
	})
	assert.EqualValues(t, 0, success)
}

func TestIgnoreErrors(t *testing.T) {
	i := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		i++
		w.WriteHeader(http.StatusInternalServerError)
		switch i {
		case 1:
			_, _ = w.Write([]byte(`{"error":" "write failed: hinted handoff queue not empty"`))
		case 2:
			_, _ = w.Write([]byte(`{"code":"internal error", "message":"partial write: field type conflict"}`))
		case 3:
			_, _ = w.Write([]byte(`{"code":"internal error", "message":"partial write: points beyond retention policy"}`))
		case 4:
			_, _ = w.Write([]byte(`{"code":"internal error", "message":"unable to parse 'cpu value': invalid field format"}`))
		case 5:
			_, _ = w.Write([]byte(`{"code":"internal error", "message":"gateway error"}`))
		}
	}))
	defer server.Close()

	cl, err := New(Params{ServerURL: server.URL})
	require.NoError(t, err)

	writer := cl.PointsWriter("bucket")

	b := &batch{
		lines:             []byte("a"),
		remainingAttempts: 0,
		expires:           time.Time{},
	}
	err = writer.writeBatch(b, 0)
	assert.NoError(t, err)
	err = writer.writeBatch(b, 0)
	assert.NoError(t, err)
	err = writer.writeBatch(b, 1)
	assert.NoError(t, err)
	err = writer.writeBatch(b, 2)
	assert.NoError(t, err)
	err = writer.writeBatch(b, 3)
	assert.Error(t, err)
}
