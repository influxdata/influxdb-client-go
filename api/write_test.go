// Copyright 2020 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package api

import (
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/influxdata/influxdb-client-go/api/write"
	ihttp "github.com/influxdata/influxdb-client-go/internal/http"
	"github.com/influxdata/influxdb-client-go/internal/log"
	"github.com/influxdata/influxdb-client-go/internal/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func genPoints(num int) []*write.Point {
	points := make([]*write.Point, num)
	rand.Seed(321)

	t := time.Now()
	for i := 0; i < len(points); i++ {
		points[i] = write.NewPoint(
			"test",
			map[string]string{
				"id":       fmt.Sprintf("rack_%v", i%10),
				"vendor":   "AWS",
				"hostname": fmt.Sprintf("host_%v", i%100),
			},
			map[string]interface{}{
				"temperature": rand.Float64() * 80.0,
				"disk_free":   rand.Float64() * 1000.0,
				"disk_total":  (i/10 + 1) * 1000000,
				"mem_total":   (i/100 + 1) * 10000000,
				"mem_free":    rand.Uint64(),
			},
			t)
		if i%10 == 0 {
			t = t.Add(time.Second)
		}
	}
	return points
}

func genRecords(num int) []string {
	lines := make([]string, num)
	rand.Seed(321)

	t := time.Now()
	for i := 0; i < len(lines); i++ {
		lines[i] = fmt.Sprintf("test,id=rack_%v,vendor=AWS,hostname=host_%v temperature=%v,disk_free=%v,disk_total=%vi,mem_total=%vi,mem_free=%vu %v",
			i%10, i%100, rand.Float64()*80.0, rand.Float64()*1000.0, (i/10+1)*1000000, (i/100+1)*10000000, rand.Uint64(), t.UnixNano())
		if i%10 == 0 {
			t = t.Add(time.Second)
		}
	}
	return lines
}

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
	points := genPoints(10)
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
	log.Log.SetDebugLevel(4)
	writeAPI := NewWriteAPI("my-org", "my-bucket", service, write.DefaultOptions().SetBatchSize(5).SetUseGZip(true))
	points := genPoints(5)
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
	writeAPI := NewWriteAPI("my-org", "my-bucket", service, write.DefaultOptions().SetBatchSize(10).SetFlushInterval(500))
	points := genPoints(5)
	for _, p := range points {
		writeAPI.WritePoint(p)
	}
	require.Len(t, service.Lines(), 0)
	time.Sleep(time.Millisecond * 600)
	require.Len(t, service.Lines(), 5)
	writeAPI.Close()

	service.Close()
	writeAPI = NewWriteAPI("my-org", "my-bucket", service, writeAPI.writeOptions.SetFlushInterval(2000))
	for _, p := range points {
		writeAPI.WritePoint(p)
	}
	require.Len(t, service.Lines(), 0)
	time.Sleep(time.Millisecond * 2100)
	require.Len(t, service.Lines(), 5)

	writeAPI.Close()
}

func TestRetry(t *testing.T) {
	service := test.NewTestService(t, "http://localhost:8888")
	log.Log.SetDebugLevel(5)
	writeAPI := NewWriteAPI("my-org", "my-bucket", service, write.DefaultOptions().SetBatchSize(5).SetRetryInterval(10000))
	points := genPoints(15)
	for i := 0; i < 5; i++ {
		writeAPI.WritePoint(points[i])
	}
	writeAPI.waitForFlushing()
	require.Len(t, service.Lines(), 5)
	service.Close()
	service.SetReplyError(&ihttp.Error{
		StatusCode: 429,
		RetryAfter: 5,
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
	time.Sleep(5*time.Second + 50*time.Millisecond)
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
	log.Log.SetDebugLevel(3)
	service.SetReplyError(&ihttp.Error{
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
	points := genPoints(15)
	for i := 0; i < 5; i++ {
		writeAPI.WritePoint(points[i])
	}
	writeAPI.waitForFlushing()
	wg.Wait()
	require.NotNil(t, recErr)
	writeAPI.Close()
}
