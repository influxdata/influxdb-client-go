// Copyright 2020 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package api

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/influxdata/influxdb-client-go/api/write"
	ihttp "github.com/influxdata/influxdb-client-go/internal/http"
	"github.com/influxdata/influxdb-client-go/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testHTTPService struct {
	serverURL      string
	authorization  string
	lines          []string
	t              *testing.T
	wasGzip        bool
	requestHandler func(c *testHTTPService, url string, body io.Reader) error
	replyError     *ihttp.Error
	lock           sync.Mutex
}

func (t *testHTTPService) ServerURL() string {
	return t.serverURL
}

func (t *testHTTPService) ServerAPIURL() string {
	return t.serverURL
}

func (t *testHTTPService) Authorization() string {
	return t.authorization
}

func (t *testHTTPService) HTTPClient() *http.Client {
	return nil
}

func (t *testHTTPService) Close() {
	t.lock.Lock()
	if len(t.lines) > 0 {
		t.lines = t.lines[:0]
	}
	t.wasGzip = false
	t.replyError = nil
	t.requestHandler = nil
	t.lock.Unlock()
}

func (t *testHTTPService) ReplyError() *ihttp.Error {
	t.lock.Lock()
	defer t.lock.Unlock()
	return t.replyError
}

func (t *testHTTPService) SetAuthorization(_ string) {

}
func (t *testHTTPService) GetRequest(_ context.Context, _ string, _ ihttp.RequestCallback, _ ihttp.ResponseCallback) *ihttp.Error {
	return nil
}
func (t *testHTTPService) DoHTTPRequest(_ *http.Request, _ ihttp.RequestCallback, _ ihttp.ResponseCallback) *ihttp.Error {
	return nil
}

func (t *testHTTPService) DoHTTPRequestWithResponse(_ *http.Request, _ ihttp.RequestCallback) (*http.Response, error) {
	return nil, nil
}

func (t *testHTTPService) PostRequest(_ context.Context, url string, body io.Reader, requestCallback ihttp.RequestCallback, _ ihttp.ResponseCallback) *ihttp.Error {
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return ihttp.NewError(err)
	}
	if requestCallback != nil {
		requestCallback(req)
	}
	if req.Header.Get("Content-Encoding") == "gzip" {
		body, _ = gzip.NewReader(body)
		t.wasGzip = true
	}
	assert.Equal(t.t, fmt.Sprintf("%swrite?bucket=my-bucket&org=my-org&precision=ns", t.serverURL), url)

	if t.ReplyError() != nil {
		return t.ReplyError()
	}
	if t.requestHandler != nil {
		err = t.requestHandler(t, url, body)
	} else {
		err = t.decodeLines(body)
	}

	if err != nil {
		return ihttp.NewError(err)
	} else {
		return nil
	}
}

func (t *testHTTPService) decodeLines(body io.Reader) error {
	bytes, err := ioutil.ReadAll(body)
	if err != nil {
		return err
	}
	lines := strings.Split(string(bytes), "\n")
	lines = lines[:len(lines)-1]
	t.lock.Lock()
	t.lines = append(t.lines, lines...)
	t.lock.Unlock()
	return nil
}

func (t *testHTTPService) Lines() []string {
	t.lock.Lock()
	defer t.lock.Unlock()
	return t.lines
}

func newTestService(t *testing.T, serverURL string) *testHTTPService {
	return &testHTTPService{
		t:         t,
		serverURL: serverURL + "/api/v2/",
	}
}

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
	service := newTestService(t, "http://localhost:8888")
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
	service := newTestService(t, "http://localhost:8888")
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
	service := newTestService(t, "http://localhost:8888")
	log.Log.SetLogLevel(log.DebugLevel)
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
	assert.True(t, service.wasGzip)

	service.Close()
	writeAPI.writeOptions.SetUseGZip(false)
	for _, p := range points {
		writeAPI.WritePoint(p)
	}
	writeAPI.waitForFlushing()
	assert.Len(t, service.Lines(), 5)
	assert.False(t, service.wasGzip)

	writeAPI.Close()
}
func TestFlushInterval(t *testing.T) {
	service := newTestService(t, "http://localhost:8888")
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
	service := newTestService(t, "http://localhost:8888")
	log.Log.SetLogLevel(log.DebugLevel)
	writeAPI := NewWriteAPI("my-org", "my-bucket", service, write.DefaultOptions().SetBatchSize(5).SetRetryInterval(10000))
	points := genPoints(15)
	for i := 0; i < 5; i++ {
		writeAPI.WritePoint(points[i])
	}
	writeAPI.waitForFlushing()
	require.Len(t, service.Lines(), 5)
	service.Close()
	service.replyError = &ihttp.Error{
		StatusCode: 429,
		RetryAfter: 5,
	}
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
	service := newTestService(t, "http://localhost:8888")
	log.Log.SetLogLevel(log.DebugLevel)
	service.replyError = &ihttp.Error{
		StatusCode: 400,
		Code:       "write",
		Message:    "error",
	}
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
