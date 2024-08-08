// Copyright 2020-2021 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package api

import (
	"fmt"
	"io"
	"math"
	ihttp "net/http"
	"net/http/httptest"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/influxdata/influxdb-client-go/v2/api/http"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
	"github.com/influxdata/influxdb-client-go/v2/internal/test"
	"github.com/influxdata/influxdb-client-go/v2/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteAPIWriteDefaultTag(t *testing.T) {
	service := test.NewTestService(t, "http://localhost:8888")
	opts := write.DefaultOptions().
		SetBatchSize(1)
	opts.AddDefaultTag("dft", "a")
	writeAPI := NewWriteAPI("my-org", "my-bucket", service, opts)
	point := write.NewPoint("test",
		map[string]string{
			"vendor": "AWS",
		},
		map[string]interface{}{
			"mem_free": 1234567,
		}, time.Unix(60, 60))
	writeAPI.WritePoint(point)
	writeAPI.Close()
	require.Len(t, service.Lines(), 1)
	assert.Equal(t, "test,dft=a,vendor=AWS mem_free=1234567i 60000000060", service.Lines()[0])
}

func TestWriteAPIImpl_Write(t *testing.T) {
	service := test.NewTestService(t, "http://localhost:8888")
	writeAPI := NewWriteAPI("my-org", "my-bucket", service, write.DefaultOptions().SetBatchSize(5))
	points := test.GenPoints(10)
	for _, p := range points {
		writeAPI.WritePoint(p)
	}
	writeAPI.Close()
	require.Len(t, service.Lines(), 10)
	for i, p := range points {
		line := write.PointToLineProtocol(p, writeAPI.writeOptions.Precision())
		//cut off last \n char
		line = line[:len(line)-1]
		assert.Equal(t, service.Lines()[i], line)
	}
}

func TestGzipWithFlushing(t *testing.T) {
	service := test.NewTestService(t, "http://localhost:8888")
	log.Log.SetLogLevel(log.DebugLevel)
	writeAPI := NewWriteAPI("my-org", "my-bucket", service, write.DefaultOptions().SetBatchSize(5).SetUseGZip(true))
	points := test.GenPoints(5)
	for _, p := range points {
		writeAPI.WritePoint(p)
	}
	start := time.Now()
	writeAPI.waitForFlushing()
	end := time.Now()
	fmt.Printf("Flash duration: %dns\n", end.Sub(start).Nanoseconds())
	assert.Len(t, service.Lines(), 5)
	assert.True(t, service.WasGzip())

	service.Close()
	writeAPI.writeOptions.SetUseGZip(false)
	for _, p := range points {
		writeAPI.WritePoint(p)
	}
	writeAPI.waitForFlushing()
	assert.Len(t, service.Lines(), 5)
	assert.False(t, service.WasGzip())

	writeAPI.Close()
}
func TestFlushInterval(t *testing.T) {
	service := test.NewTestService(t, "http://localhost:8888")
	writeAPI := NewWriteAPI("my-org", "my-bucket", service, write.DefaultOptions().SetBatchSize(10).SetFlushInterval(30))
	points := test.GenPoints(5)
	for _, p := range points {
		writeAPI.WritePoint(p)
	}
	require.Len(t, service.Lines(), 0)
	<-time.After(time.Millisecond * 50)
	require.Len(t, service.Lines(), 5)
	writeAPI.Close()

	service.Close()
}

func TestRetry(t *testing.T) {
	service := test.NewTestService(t, "http://localhost:8888")
	log.Log.SetLogLevel(log.DebugLevel)
	writeAPI := NewWriteAPI("my-org", "my-bucket", service, write.DefaultOptions().SetBatchSize(5).SetRetryInterval(10000))
	points := test.GenPoints(15)
	for i := 0; i < 5; i++ {
		writeAPI.WritePoint(points[i])
	}
	writeAPI.waitForFlushing()
	require.Len(t, service.Lines(), 5)
	service.Close()
	service.SetReplyError(&http.Error{
		StatusCode: 429,
		RetryAfter: 1,
	})
	for i := 0; i < 5; i++ {
		writeAPI.WritePoint(points[i])
	}
	writeAPI.waitForFlushing()
	require.Len(t, service.Lines(), 0)
	service.Close()
	for i := 5; i < 10; i++ {
		writeAPI.WritePoint(points[i])
	}
	writeAPI.waitForFlushing()
	require.Len(t, service.Lines(), 0)
	<-time.After(time.Second + 50*time.Millisecond)
	for i := 10; i < 15; i++ {
		writeAPI.WritePoint(points[i])
	}
	writeAPI.waitForFlushing()
	require.Len(t, service.Lines(), 15)
	assert.True(t, strings.HasPrefix(service.Lines()[7], "test,hostname=host_7"))
	assert.True(t, strings.HasPrefix(service.Lines()[14], "test,hostname=host_14"))
	writeAPI.Close()
}

func TestWriteError(t *testing.T) {
	service := test.NewTestService(t, "http://localhost:8888")
	log.Log.SetLogLevel(log.DebugLevel)
	service.SetReplyError(&http.Error{
		StatusCode: 400,
		Code:       "write",
		Message:    "error",
	})
	writeAPI := NewWriteAPI("my-org", "my-bucket", service, write.DefaultOptions().SetBatchSize(5))
	errCh := writeAPI.Errors()
	var recErr error
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		recErr = <-errCh
		wg.Done()
	}()
	points := test.GenPoints(15)
	for i := 0; i < 5; i++ {
		writeAPI.WritePoint(points[i])
	}
	writeAPI.waitForFlushing()
	wg.Wait()
	require.NotNil(t, recErr)
	writeAPI.Close()
}

func TestWriteErrorCallback(t *testing.T) {
	service := test.NewTestService(t, "http://localhost:8888")
	log.Log.SetLogLevel(log.DebugLevel)
	service.SetReplyError(&http.Error{
		StatusCode: 429,
		Code:       "write",
		Message:    "error",
	})
	// sleep takes at least more than 10ms (sometimes 15ms) on Windows https://github.com/golang/go/issues/44343
	retryInterval := uint(1)
	if runtime.GOOS == "windows" {
		retryInterval = 20
	}
	writeAPI := NewWriteAPI("my-org", "my-bucket", service, write.DefaultOptions().SetBatchSize(1).SetRetryInterval(retryInterval))
	writeAPI.SetWriteFailedCallback(func(batch string, error http.Error, retryAttempts uint) bool {
		return retryAttempts < 2
	})
	points := test.GenPoints(10)
	// first batch will be discarded by callback after 3 write attempts, second batch should survive with only one failed attempt
	for i, j := 0, 0; i < 6; i++ {
		writeAPI.WritePoint(points[i])
		writeAPI.waitForFlushing()
		w := int(math.Pow(5, float64(j)) * float64(retryInterval))
		fmt.Printf("Waiting %dms\n", w)
		<-time.After(time.Duration(w) * time.Millisecond)
		j++
		if j == 3 {
			j = 0
		}
	}
	service.SetReplyError(nil)
	writeAPI.SetWriteFailedCallback(func(batch string, error http.Error, retryAttempts uint) bool {
		return true
	})
	for i := 6; i < 10; i++ {
		writeAPI.WritePoint(points[i])
	}
	writeAPI.waitForFlushing()
	assert.Len(t, service.Lines(), 9)

	writeAPI.Close()
}

func TestClosing(t *testing.T) {
	service := test.NewTestService(t, "http://localhost:8888")
	log.Log.SetLogLevel(log.DebugLevel)
	writeAPI := NewWriteAPI("my-org", "my-bucket", service, write.DefaultOptions().SetBatchSize(5).SetRetryInterval(10000))
	points := test.GenPoints(15)
	for i := 0; i < 5; i++ {
		writeAPI.WritePoint(points[i])
	}
	writeAPI.Close()
	require.Len(t, service.Lines(), 5)

	writeAPI = NewWriteAPI("my-org", "my-bucket", service, write.DefaultOptions().SetBatchSize(5).SetRetryInterval(10000))
	service.Close()
	service.SetReplyError(&http.Error{
		StatusCode: 425,
	})
	_ = writeAPI.Errors()
	for i := 0; i < 15; i++ {
		writeAPI.WritePoint(points[i])
	}
	start := time.Now()
	writeAPI.Close()
	diff := time.Since(start)
	fmt.Println("Diff", diff)
	assert.Len(t, service.Lines(), 0)

}

func TestFlushWithRetries(t *testing.T) {
	service := test.NewTestService(t, "http://localhost:8888")
	log.Log.SetLogLevel(log.DebugLevel)
	writeAPI := NewWriteAPI("my-org", "my-bucket", service, write.DefaultOptions().SetRetryInterval(200).SetBatchSize(1))
	points := test.GenPoints(5)
	fails := 0

	var mu sync.Mutex

	service.SetRequestHandler(func(url string, body io.Reader) error {
		mu.Lock()
		defer mu.Unlock()
		// fail 4 times, then succeed on the 5th try - maxRetries default is 5
		if fails >= 4 {
			_ = service.DecodeLines(body)
			return nil
		}
		fails++
		return fmt.Errorf("spurious failure")
	})
	// write will try first batch and others will be put to the retry queue of retry delay caused by first write error
	for i := 0; i < len(points); i++ {
		writeAPI.WritePoint(points[i])
	}
	// Flush will try sending first batch again and then others
	// 1st, 2nd and 3rd will fail, because test service rejects 4 writes
	writeAPI.Flush()
	writeAPI.Close()
	// two remained
	assert.Equal(t, 2, len(service.Lines()))
}

func TestWriteApiErrorHeaders(t *testing.T) {
	calls := 0
	var mu sync.Mutex
	server := httptest.NewServer(ihttp.HandlerFunc(func(w ihttp.ResponseWriter, r *ihttp.Request) {
		mu.Lock()
		defer mu.Unlock()
		calls++
		w.Header().Set("X-Test-Val1", "Not All Correct")
		w.Header().Set("X-Test-Val2", "Atlas LV-3B")
		w.Header().Set("X-Call-Count", strconv.Itoa(calls))
		w.WriteHeader(ihttp.StatusBadRequest)
		_, _ = w.Write([]byte(`{ "code": "bad request", "message": "test header" }`))
	}))
	defer server.Close()
	svc := http.NewService(server.URL, "my-token", http.DefaultOptions())
	writeAPI := NewWriteAPI("my-org", "my-bucket", svc, write.DefaultOptions().SetBatchSize(5))
	defer writeAPI.Close()
	errCh := writeAPI.Errors()
	var wg sync.WaitGroup
	var recErr error
	wg.Add(1)
	go func() {
		for i := 0; i < 3; i++ {
			recErr = <-errCh
			assert.NotNil(t, recErr, "errCh should not run out of values")
			assert.Len(t, recErr.(*http.Error).Header, 6)
			assert.NotEqual(t, "", recErr.(*http.Error).Header.Get("Date"))
			assert.NotEqual(t, "", recErr.(*http.Error).Header.Get("Content-Length"))
			assert.NotEqual(t, "", recErr.(*http.Error).Header.Get("Content-Type"))
			assert.Equal(t, strconv.Itoa(i+1), recErr.(*http.Error).Header.Get("X-Call-Count"))
			assert.Equal(t, "Not All Correct", recErr.(*http.Error).Header.Get("X-Test-Val1"))
			assert.Equal(t, "Atlas LV-3B", recErr.(*http.Error).Header.Get("X-Test-Val2"))
		}
		wg.Done()
	}()
	points := test.GenPoints(15)
	for i := 0; i < 15; i++ {
		writeAPI.WritePoint(points[i])
	}
	writeAPI.waitForFlushing()
	wg.Wait()
	assert.Equal(t, calls, 3)
}
