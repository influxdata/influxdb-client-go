// Copyright 2020 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package influxdb2

import (
	"compress/gzip"
	"context"
	"fmt"
	"github.com/influxdata/influxdb-client-go/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"
)

type testClient struct {
	lines          []string
	options        *Options
	t              *testing.T
	wasGzip        bool
	requestHandler func(c *testClient, url string, body io.Reader) error
	replyError     *Error
	lock           sync.Mutex
}

func (t *testClient) WriteApiBlocking(string, string) WriteApiBlocking {
	return nil
}

func (t *testClient) WriteApi(string, string) WriteApi {
	return nil
}

func (t *testClient) Close() {
	t.lock.Lock()
	if len(t.lines) > 0 {
		t.lines = t.lines[:0]
	}
	t.wasGzip = false
	t.replyError = nil
	t.requestHandler = nil
	t.lock.Unlock()
}

func (t *testClient) QueryApi(string) QueryApi {
	return nil
}

func (t *testClient) ReplyError() *Error {
	t.lock.Lock()
	defer t.lock.Unlock()
	return t.replyError
}

func (t *testClient) postRequest(_ context.Context, url string, body io.Reader, requestCallback RequestCallback, _ ResponseCallback) *Error {
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return NewError(err)
	}
	if requestCallback != nil {
		requestCallback(req)
	}
	if req.Header.Get("Content-Encoding") == "gzip" {
		body, _ = gzip.NewReader(body)
		t.wasGzip = true
	}
	assert.Equal(t.t, url, fmt.Sprintf("%s/api/v2/write?bucket=my-bucket&org=my-org&precision=ns", t.ServerUrl()))

	if t.ReplyError() != nil {
		return t.ReplyError()
	}
	if t.requestHandler != nil {
		err = t.requestHandler(t, url, body)
	} else {
		err = t.decodeLines(body)
	}

	if err != nil {
		return NewError(err)
	} else {
		return nil
	}
}

func (t *testClient) decodeLines(body io.Reader) error {
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

func (t *testClient) Lines() []string {
	t.lock.Lock()
	defer t.lock.Unlock()
	return t.lines
}

func (t *testClient) Options() *Options {
	return t.options
}

func (t *testClient) ServerUrl() string {
	return "http://locahost:8900"
}

func (t *testClient) Setup(context.Context, string, string, string, string, int) (*domain.OnboardingResponse, error) {
	return nil, nil
}
func (t *testClient) Ready(context.Context) (bool, error) {
	return true, nil
}

func genPoints(num int) []*Point {
	points := make([]*Point, num)
	rand.Seed(321)

	t := time.Now()
	for i := 0; i < len(points); i++ {
		points[i] = NewPoint(
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

func TestWriteApiImpl_Write(t *testing.T) {
	client := &testClient{
		options: DefaultOptions(),
		t:       t,
	}
	client.options.SetBatchSize(5)
	writeApi := newWriteApiImpl("my-org", "my-bucket", client)
	points := genPoints(10)
	for _, p := range points {
		writeApi.WritePoint(p)
	}
	writeApi.Close()
	require.Len(t, client.Lines(), 10)
	for i, p := range points {
		line := p.ToLineProtocol(client.options.Precision())
		//cut off last \n char
		line = line[:len(line)-1]
		assert.Equal(t, client.Lines()[i], line)
	}
}

func TestGzipWithFlushing(t *testing.T) {
	client := &testClient{
		options: DefaultOptions(),
		t:       t,
	}
	client.options.SetBatchSize(5).SetUseGZip(true)
	writeApi := newWriteApiImpl("my-org", "my-bucket", client)
	points := genPoints(5)
	for _, p := range points {
		writeApi.WritePoint(p)
	}
	time.Sleep(time.Millisecond * 10)
	require.Len(t, client.Lines(), 5)
	assert.True(t, client.wasGzip)

	client.Close()
	client.options.SetUseGZip(false)
	for _, p := range points {
		writeApi.WritePoint(p)
	}
	time.Sleep(time.Millisecond * 10)
	require.Len(t, client.Lines(), 5)
	assert.False(t, client.wasGzip)

	writeApi.Close()
}
func TestFlushInterval(t *testing.T) {
	client := &testClient{
		options: DefaultOptions(),
		t:       t,
	}
	client.options.SetBatchSize(10).SetFlushInterval(500)
	writeApi := newWriteApiImpl("my-org", "my-bucket", client)
	points := genPoints(5)
	for _, p := range points {
		writeApi.WritePoint(p)
	}
	require.Len(t, client.Lines(), 0)
	time.Sleep(time.Millisecond * 600)
	require.Len(t, client.Lines(), 5)
	writeApi.Close()

	client.Close()
	client.options.SetFlushInterval(2000)
	writeApi = newWriteApiImpl("my-org", "my-bucket", client)
	for _, p := range points {
		writeApi.WritePoint(p)
	}
	require.Len(t, client.Lines(), 0)
	time.Sleep(time.Millisecond * 2100)
	require.Len(t, client.Lines(), 5)

	writeApi.Close()
}

func TestRetry(t *testing.T) {
	client := &testClient{
		options: DefaultOptions(),
		t:       t,
	}
	client.options.SetLogLevel(3).
		SetBatchSize(5).
		SetRetryInterval(10000)
	writeApi := newWriteApiImpl("my-org", "my-bucket", client)
	points := genPoints(15)
	for i := 0; i < 5; i++ {
		writeApi.WritePoint(points[i])
	}
	writeApi.waitForFlushing()
	require.Len(t, client.Lines(), 5)
	client.Close()
	client.replyError = &Error{
		StatusCode: 429,
		RetryAfter: 5,
	}
	for i := 0; i < 5; i++ {
		writeApi.WritePoint(points[i])
	}
	writeApi.waitForFlushing()
	require.Len(t, client.Lines(), 0)
	client.Close()
	for i := 5; i < 10; i++ {
		writeApi.WritePoint(points[i])
	}
	writeApi.waitForFlushing()
	require.Len(t, client.Lines(), 0)
	time.Sleep(5*time.Second + 50*time.Millisecond)
	for i := 10; i < 15; i++ {
		writeApi.WritePoint(points[i])
	}
	writeApi.waitForFlushing()
	require.Len(t, client.Lines(), 15)
	assert.True(t, strings.HasPrefix(client.Lines()[7], "test,hostname=host_7"))
	assert.True(t, strings.HasPrefix(client.Lines()[14], "test,hostname=host_14"))
	writeApi.Close()
}

func TestWriteError(t *testing.T) {
	client := &testClient{
		options: DefaultOptions(),
		t:       t,
	}
	client.options.SetLogLevel(3).SetBatchSize(5)
	client.replyError = &Error{
		StatusCode: 400,
		Code:       "write",
		Message:    "error",
	}
	writeApi := newWriteApiImpl("my-org", "my-bucket", client)
	errCh := writeApi.Errors()
	var recErr error
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		recErr = <-errCh
		wg.Done()
	}()
	points := genPoints(15)
	for i := 0; i < 5; i++ {
		writeApi.WritePoint(points[i])
	}
	writeApi.waitForFlushing()
	wg.Wait()
	require.NotNil(t, recErr)

	client.Close()
}
