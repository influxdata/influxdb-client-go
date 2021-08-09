// Copyright 2020-2021 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package write

import (
	"context"
	"errors"
	"fmt"
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

func TestPrecisionToString(t *testing.T) {
	assert.Equal(t, "ns", precisionToString(time.Nanosecond))
	assert.Equal(t, "us", precisionToString(time.Microsecond))
	assert.Equal(t, "ms", precisionToString(time.Millisecond))
	assert.Equal(t, "s", precisionToString(time.Second))
	assert.Equal(t, "ns", precisionToString(time.Hour))
	assert.Equal(t, "ns", precisionToString(time.Microsecond*20))
}

func TestAddDefaultTags(t *testing.T) {
	hs := test.NewTestService(t, "http://localhost:8888")
	opts := write.DefaultOptions()
	assert.Len(t, opts.DefaultTags(), 0)

	opts.AddDefaultTag("dt1", "val1")
	opts.AddDefaultTag("zdt", "val2")
	srv := NewService("org", "buc", hs, opts)

	p := write.NewPointWithMeasurement("test")
	p.AddTag("id", "101")

	p.AddField("float32", float32(80.0))

	s, err := srv.EncodePoints(p)
	require.Nil(t, err)
	assert.Equal(t, "test,dt1=val1,id=101,zdt=val2 float32=80\n", s)
	assert.Len(t, p.TagList(), 1)

	p = write.NewPointWithMeasurement("x")
	p.AddTag("xt", "1")
	p.AddField("i", 1)

	s, err = srv.EncodePoints(p)
	require.Nil(t, err)
	assert.Equal(t, "x,dt1=val1,xt=1,zdt=val2 i=1i\n", s)
	assert.Len(t, p.TagList(), 1)

	p = write.NewPointWithMeasurement("d")
	p.AddTag("id", "1")
	// do not overwrite point tag
	p.AddTag("zdt", "val10")
	p.AddField("i", -1)

	s, err = srv.EncodePoints(p)
	require.Nil(t, err)
	assert.Equal(t, "d,dt1=val1,id=1,zdt=val10 i=-1i\n", s)

	assert.Len(t, p.TagList(), 2)
}

func TestRetryStrategy(t *testing.T) {
	log.Log.SetLogLevel(log.DebugLevel)
	hs := test.NewTestService(t, "http://localhost:8086")
	opts := write.DefaultOptions().SetRetryInterval(1)
	ctx := context.Background()
	srv := NewService("my-org", "my-bucket", hs, opts)
	hs.SetReplyError(&http.Error{
		StatusCode: 429,
	})
	b1 := NewBatch("1\n", opts.RetryInterval(), opts.MaxRetryTime())
	err := srv.HandleWrite(ctx, b1)
	assert.NotNil(t, err)
	assert.EqualValues(t, 1, b1.retryDelay)
	assert.Equal(t, 1, srv.retryQueue.list.Len())

	//wait retry delay + little more
	<-time.After(time.Millisecond*time.Duration(b1.retryDelay) + time.Microsecond*5)
	b2 := NewBatch("2\n", opts.RetryInterval(), opts.MaxRetryTime())
	err = srv.HandleWrite(ctx, b2)
	assert.NotNil(t, err)
	assertBetween(t, b1.retryDelay, 2, 4)
	assert.Equal(t, 2, srv.retryQueue.list.Len())

	//wait retry delay + little more
	<-time.After(time.Millisecond*time.Duration(b1.retryDelay) + time.Microsecond*5)
	b3 := NewBatch("3\n", opts.RetryInterval(), opts.MaxRetryTime())
	err = srv.HandleWrite(ctx, b3)
	assert.NotNil(t, err)
	assertBetween(t, b1.retryDelay, 4, 8)
	assert.Equal(t, 3, srv.retryQueue.list.Len())

	//wait retry delay + little more
	<-time.After(time.Millisecond*time.Duration(b1.retryDelay) + time.Microsecond*5)
	b4 := NewBatch("4\n", opts.RetryInterval(), opts.MaxRetryTime())
	err = srv.HandleWrite(ctx, b4)
	assert.NotNil(t, err)
	assertBetween(t, b1.retryDelay, 8, 16)
	assert.Equal(t, 4, srv.retryQueue.list.Len())

	// let write pass and it will clear queue
	<-time.After(time.Millisecond*time.Duration(b1.retryDelay) + time.Microsecond*5)
	hs.SetReplyError(nil)
	err = srv.HandleWrite(ctx, NewBatch("5\n", opts.RetryInterval(), opts.MaxRetryTime()))
	assert.Nil(t, err)
	assert.Equal(t, 0, srv.retryQueue.list.Len())
	require.Len(t, hs.Lines(), 5)
	assert.Equal(t, "1", hs.Lines()[0])
	assert.Equal(t, "2", hs.Lines()[1])
	assert.Equal(t, "3", hs.Lines()[2])
	assert.Equal(t, "4", hs.Lines()[3])
	assert.Equal(t, "5", hs.Lines()[4])
}

func TestBufferOverwrite(t *testing.T) {
	log.Log.SetLogLevel(log.DebugLevel)
	hs := test.NewTestService(t, "http://localhost:8086")
	// Buffer limit 15000, bach ii 5000 => buffer for 3 batches
	opts := write.DefaultOptions().SetRetryInterval(1).SetRetryBufferLimit(15000)
	ctx := context.Background()
	srv := NewService("my-org", "my-bucket", hs, opts)
	hs.SetReplyError(&http.Error{
		StatusCode: 429,
	})
	b1 := NewBatch("1\n", opts.RetryInterval(), opts.MaxRetryTime())
	err := srv.HandleWrite(ctx, b1)
	assert.NotNil(t, err)
	assert.Equal(t, uint(1), b1.retryDelay)
	assert.Equal(t, 1, srv.retryQueue.list.Len())

	<-time.After(time.Millisecond * time.Duration(b1.retryDelay))
	b2 := NewBatch("2\n", opts.RetryInterval(), opts.MaxRetryTime())
	err = srv.HandleWrite(ctx, b2)
	assert.NotNil(t, err)
	assertBetween(t, b1.retryDelay, 2, 4)
	assert.Equal(t, 2, srv.retryQueue.list.Len())

	<-time.After(time.Millisecond * time.Duration(b1.retryDelay))
	b3 := NewBatch("3\n", opts.RetryInterval(), opts.MaxRetryTime())
	err = srv.HandleWrite(ctx, b3)
	assert.NotNil(t, err)
	assertBetween(t, b1.retryDelay, 4, 8)
	assert.Equal(t, 3, srv.retryQueue.list.Len())

	// Write early
	<-time.After(time.Millisecond)
	b4 := NewBatch("4\n", opts.RetryInterval(), opts.MaxRetryTime())
	err = srv.HandleWrite(ctx, b4)
	assert.NoError(t, err)
	assert.Equal(t, uint(1), b2.retryDelay)
	assert.Equal(t, 3, srv.retryQueue.list.Len())

	// Overwrite
	<-time.After(time.Millisecond * time.Duration(b1.retryDelay) / 2)
	b5 := NewBatch("5\n", opts.RetryInterval(), opts.MaxRetryTime())
	err = srv.HandleWrite(ctx, b5)
	assert.Error(t, err)
	assertBetween(t, b2.retryDelay, 2, 4)
	assert.Equal(t, 3, srv.retryQueue.list.Len())

	// let write pass and it will clear queue
	<-time.After(time.Millisecond * time.Duration(b1.retryDelay))
	hs.SetReplyError(nil)
	err = srv.HandleWrite(ctx, NewBatch("6\n", opts.RetryInterval(), opts.MaxRetryTime()))
	assert.Nil(t, err)
	assert.Equal(t, 0, srv.retryQueue.list.Len())
	require.Len(t, hs.Lines(), 4)
	assert.Equal(t, "3", hs.Lines()[0])
	assert.Equal(t, "4", hs.Lines()[1])
	assert.Equal(t, "5", hs.Lines()[2])
	assert.Equal(t, "6", hs.Lines()[3])
}

func TestMaxRetryInterval(t *testing.T) {
	log.Log.SetLogLevel(log.DebugLevel)
	hs := test.NewTestService(t, "http://localhost:8086")
	//
	opts := write.DefaultOptions().SetRetryInterval(1).SetMaxRetryInterval(4)
	ctx := context.Background()
	srv := NewService("my-org", "my-bucket", hs, opts)
	hs.SetReplyError(&http.Error{
		StatusCode: 503,
	})
	b1 := NewBatch("1\n", opts.RetryInterval(), opts.MaxRetryTime())
	err := srv.HandleWrite(ctx, b1)
	assert.NotNil(t, err)
	assert.Equal(t, uint(1), b1.retryDelay)
	assert.Equal(t, 1, srv.retryQueue.list.Len())

	<-time.After(time.Millisecond * time.Duration(b1.retryDelay))
	b2 := NewBatch("2\n", opts.RetryInterval(), opts.MaxRetryTime())
	err = srv.HandleWrite(ctx, b2)
	assert.NotNil(t, err)
	assertBetween(t, b1.retryDelay, 2, 4)
	assert.Equal(t, 2, srv.retryQueue.list.Len())

	<-time.After(time.Millisecond * time.Duration(b1.retryDelay))
	b3 := NewBatch("3\n", opts.RetryInterval(), opts.MaxRetryTime())
	err = srv.HandleWrite(ctx, b3)
	assert.NotNil(t, err)
	assert.EqualValues(t, 4, b1.retryDelay)
	assert.Equal(t, 3, srv.retryQueue.list.Len())
}

func TestMaxRetries(t *testing.T) {
	log.Log.SetLogLevel(log.DebugLevel)
	hs := test.NewTestService(t, "http://localhost:8086")
	opts := write.DefaultOptions().SetRetryInterval(1)
	ctx := context.Background()
	srv := NewService("my-org", "my-bucket", hs, opts)
	hs.SetReplyError(&http.Error{
		StatusCode: 429,
	})
	b1 := NewBatch("1\n", opts.RetryInterval(), opts.MaxRetryTime())
	err := srv.HandleWrite(ctx, b1)
	assert.NotNil(t, err)
	assert.EqualValues(t, 1, b1.retryDelay)
	assert.Equal(t, 1, srv.retryQueue.list.Len())

	for i, e := uint(1), uint(2); i <= opts.MaxRetries(); i++ {
		//wait retry delay + little more
		<-time.After(time.Millisecond*time.Duration(b1.retryDelay) + time.Microsecond*5)
		b := NewBatch(fmt.Sprintf("%d\n", i+1), opts.RetryInterval(), opts.MaxRetryTime())
		err = srv.HandleWrite(ctx, b)
		assert.NotNil(t, err)
		assertBetween(t, b1.retryDelay, e, e*2)
		exp := min(i+1, opts.MaxRetries())
		assert.EqualValues(t, exp, srv.retryQueue.list.Len())
		e *= 2
	}
	assert.True(t, b1.evicted)

	// let write pass and it will clear queue
	<-time.After(time.Millisecond*time.Duration(b1.retryDelay) + time.Microsecond*5)
	hs.SetReplyError(nil)
	err = srv.HandleWrite(ctx, NewBatch(fmt.Sprintf("%d\n", opts.MaxRetries()+2), opts.RetryInterval(), opts.MaxRetryTime()))
	assert.Nil(t, err)
	assert.Equal(t, 0, srv.retryQueue.list.Len())
	require.Len(t, hs.Lines(), int(opts.MaxRetries()+1))
	for i := uint(2); i <= opts.MaxRetries()+2; i++ {
		assert.Equal(t, fmt.Sprintf("%d", i), hs.Lines()[i-2])
	}
}

func TestMaxRetryTime(t *testing.T) {
	log.Log.SetLogLevel(log.DebugLevel)
	hs := test.NewTestService(t, "http://localhost:8086")
	opts := write.DefaultOptions().SetRetryInterval(1).SetMaxRetryTime(5)
	ctx := context.Background()
	srv := NewService("my-org", "my-bucket", hs, opts)
	hs.SetReplyError(&http.Error{
		StatusCode: 429,
	})
	b1 := NewBatch("1\n", opts.RetryInterval(), opts.MaxRetryTime())
	err := srv.HandleWrite(ctx, b1)
	assert.NotNil(t, err)
	assert.EqualValues(t, 1, b1.retryDelay)
	assert.Equal(t, 1, srv.retryQueue.list.Len())
	<-time.After(5 * time.Millisecond)

	b := NewBatch("2\n", opts.RetryInterval(), opts.MaxRetryTime())
	err = srv.HandleWrite(ctx, b)
	require.NotNil(t, err)
	assert.Equal(t, "write failed (attempts 1): max retry time exceeded", err.Error())
	assert.Equal(t, 1, srv.retryQueue.list.Len())
	// let write pass and it will clear queue
	hs.SetReplyError(nil)

	err = srv.HandleWrite(ctx, NewBatch("3\n", opts.RetryInterval(), opts.MaxRetryTime()))
	assert.Nil(t, err)
	assert.Equal(t, 0, srv.retryQueue.list.Len())
	require.Len(t, hs.Lines(), 2)
	assert.Equal(t, "2", hs.Lines()[0])
	assert.Equal(t, "3", hs.Lines()[1])
}

func TestRetryOnConnectionError(t *testing.T) {
	log.Log.SetLogLevel(log.DebugLevel)
	hs := test.NewTestService(t, "http://localhost:8086")
	//
	opts := write.DefaultOptions().SetRetryInterval(1).SetRetryBufferLimit(15000)
	ctx := context.Background()
	srv := NewService("my-org", "my-bucket", hs, opts)

	hs.SetReplyError(&http.Error{
		Err: errors.New("connection refused"),
	})

	b1 := NewBatch("1\n", opts.RetryInterval(), opts.MaxRetryTime())
	err := srv.HandleWrite(ctx, b1)
	assert.NotNil(t, err)
	assert.EqualValues(t, 1, b1.retryDelay)
	assert.Equal(t, 1, srv.retryQueue.list.Len())

	<-time.After(time.Millisecond * time.Duration(b1.retryDelay))
	b2 := NewBatch("2\n", opts.RetryInterval(), opts.MaxRetryTime())
	err = srv.HandleWrite(ctx, b2)
	assert.NotNil(t, err)
	assertBetween(t, b1.retryDelay, 2, 4)
	assert.Equal(t, 2, srv.retryQueue.list.Len())

	<-time.After(time.Millisecond * time.Duration(b1.retryDelay))
	b3 := NewBatch("3\n", opts.RetryInterval(), opts.MaxRetryTime())
	err = srv.HandleWrite(ctx, b3)
	assert.NotNil(t, err)
	assertBetween(t, b1.retryDelay, 4, 8)
	assert.Equal(t, 3, srv.retryQueue.list.Len())

	// let write pass and it will clear queue
	<-time.After(time.Millisecond * time.Duration(b1.retryDelay))
	hs.SetReplyError(nil)
	err = srv.HandleWrite(ctx, NewBatch("4\n", opts.RetryInterval(), opts.MaxRetryTime()))
	assert.Nil(t, err)
	assert.Equal(t, 0, srv.retryQueue.list.Len())
	require.Len(t, hs.Lines(), 4)
	assert.Equal(t, "1", hs.Lines()[0])
	assert.Equal(t, "2", hs.Lines()[1])
	assert.Equal(t, "3", hs.Lines()[2])
	assert.Equal(t, "4", hs.Lines()[3])

}

func TestNoRetryIfMaxRetriesIsZero(t *testing.T) {
	log.Log.SetLogLevel(log.DebugLevel)
	hs := test.NewTestService(t, "http://localhost:8086")
	//
	opts := write.DefaultOptions().SetMaxRetries(0)
	ctx := context.Background()
	srv := NewService("my-org", "my-bucket", hs, opts)

	hs.SetReplyError(&http.Error{
		Err: errors.New("connection refused"),
	})

	b1 := NewBatch("1\n", opts.RetryInterval(), opts.MaxRetryTime())
	err := srv.HandleWrite(ctx, b1)
	assert.NotNil(t, err)
	assert.Equal(t, 0, srv.retryQueue.list.Len())
}

func TestWriteContextCancel(t *testing.T) {
	hs := test.NewTestService(t, "http://localhost:8888")
	opts := write.DefaultOptions()
	srv := NewService("my-org", "my-bucket", hs, opts)
	lines := test.GenRecords(10)
	ctx, cancel := context.WithCancel(context.Background())
	var err error
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		<-time.After(10 * time.Millisecond)
		err = srv.HandleWrite(ctx, NewBatch(strings.Join(lines, "\n"), opts.RetryInterval(), opts.MaxRetryTime()))
		wg.Done()
	}()
	cancel()
	wg.Wait()
	require.Equal(t, context.Canceled, err)
	assert.Len(t, hs.Lines(), 0)
}

func TestPow(t *testing.T) {
	assert.EqualValues(t, 1, pow(10, 0))
	assert.EqualValues(t, 10, pow(10, 1))
	assert.EqualValues(t, 4, pow(2, 2))
	assert.EqualValues(t, 1, pow(1, 2))
	assert.EqualValues(t, 125, pow(5, 3))
}

func assertBetween(t *testing.T, val, min, max uint) {
	t.Helper()
	assert.True(t, val >= min && val <= max, fmt.Sprintf("%d is outside <%d;%d>", val, min, max))
}

func TestComputeRetryDelay(t *testing.T) {
	hs := test.NewTestService(t, "http://localhost:8888")
	opts := write.DefaultOptions()
	srv := NewService("my-org", "my-bucket", hs, opts)
	assertBetween(t, srv.computeRetryDelay(0), 5_000, 10_000)
	assertBetween(t, srv.computeRetryDelay(1), 10_000, 20_000)
	assertBetween(t, srv.computeRetryDelay(2), 20_000, 40_000)
	assertBetween(t, srv.computeRetryDelay(3), 40_000, 80_000)
	assertBetween(t, srv.computeRetryDelay(4), 80_000, 125_000)
	assert.EqualValues(t, 125_000, srv.computeRetryDelay(5))
}
