// Copyright 2020-2021 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package write

import (
	"context"
	"errors"
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
	b1 := NewBatch("1\n", opts.RetryInterval())
	err := srv.HandleWrite(ctx, b1)
	assert.NotNil(t, err)
	assert.Equal(t, uint(1), b1.retryDelay)
	assert.Equal(t, 1, srv.retryQueue.list.Len())

	//wait retry delay + little more
	<-time.After(time.Millisecond*time.Duration(b1.retryDelay) + time.Microsecond*5)
	b2 := NewBatch("2\n", opts.RetryInterval())
	err = srv.HandleWrite(ctx, b2)
	assert.NotNil(t, err)
	assert.Equal(t, uint(5), b1.retryDelay)
	assert.Equal(t, 2, srv.retryQueue.list.Len())

	//wait retry delay + little more
	<-time.After(time.Millisecond*time.Duration(b1.retryDelay) + time.Microsecond*5)
	b3 := NewBatch("3\n", opts.RetryInterval())
	err = srv.HandleWrite(ctx, b3)
	assert.NotNil(t, err)
	assert.Equal(t, uint(25), b1.retryDelay)
	assert.Equal(t, 3, srv.retryQueue.list.Len())

	// let write pass and it will clear queue
	<-time.After(time.Millisecond*time.Duration(b1.retryDelay) + time.Microsecond*5)
	hs.SetReplyError(nil)
	err = srv.HandleWrite(ctx, NewBatch("4\n", opts.RetryInterval()))
	assert.Nil(t, err)
	assert.Equal(t, 0, srv.retryQueue.list.Len())
	require.Len(t, hs.Lines(), 4)
	assert.Equal(t, "1", hs.Lines()[0])
	assert.Equal(t, "2", hs.Lines()[1])
	assert.Equal(t, "3", hs.Lines()[2])
	assert.Equal(t, "4", hs.Lines()[3])
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
	b1 := NewBatch("1\n", opts.RetryInterval())
	err := srv.HandleWrite(ctx, b1)
	assert.NotNil(t, err)
	assert.Equal(t, uint(1), b1.retryDelay)
	assert.Equal(t, 1, srv.retryQueue.list.Len())

	<-time.After(time.Millisecond * time.Duration(b1.retryDelay))
	b2 := NewBatch("2\n", opts.RetryInterval())
	err = srv.HandleWrite(ctx, b2)
	assert.NotNil(t, err)
	assert.Equal(t, uint(5), b1.retryDelay)
	assert.Equal(t, 2, srv.retryQueue.list.Len())

	<-time.After(time.Millisecond * time.Duration(b1.retryDelay))
	b3 := NewBatch("3\n", opts.RetryInterval())
	err = srv.HandleWrite(ctx, b3)
	assert.NotNil(t, err)
	assert.Equal(t, uint(25), b1.retryDelay)
	assert.Equal(t, 3, srv.retryQueue.list.Len())

	// Write early
	<-time.After(time.Millisecond * time.Duration(b1.retryDelay) / 2)
	b4 := NewBatch("4\n", opts.RetryInterval())
	err = srv.HandleWrite(ctx, b4)
	assert.NoError(t, err)
	assert.Equal(t, uint(1), b2.retryDelay)
	assert.Equal(t, 3, srv.retryQueue.list.Len())

	// Overwrite
	<-time.After(time.Millisecond * time.Duration(b1.retryDelay) / 2)
	b5 := NewBatch("5\n", opts.RetryInterval())
	err = srv.HandleWrite(ctx, b5)
	assert.Error(t, err)
	assert.Equal(t, uint(5), b2.retryDelay)
	assert.Equal(t, 3, srv.retryQueue.list.Len())

	// let write pass and it will clear queue
	<-time.After(time.Millisecond * time.Duration(b1.retryDelay))
	hs.SetReplyError(nil)
	err = srv.HandleWrite(ctx, NewBatch("6\n", opts.RetryInterval()))
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
	opts := write.DefaultOptions().SetRetryInterval(1).SetMaxRetryInterval(10)
	ctx := context.Background()
	srv := NewService("my-org", "my-bucket", hs, opts)
	hs.SetReplyError(&http.Error{
		StatusCode: 503,
	})
	b1 := NewBatch("1\n", opts.RetryInterval())
	err := srv.HandleWrite(ctx, b1)
	assert.NotNil(t, err)
	assert.Equal(t, uint(1), b1.retryDelay)
	assert.Equal(t, 1, srv.retryQueue.list.Len())

	<-time.After(time.Millisecond * time.Duration(b1.retryDelay))
	b2 := NewBatch("2\n", opts.RetryInterval())
	err = srv.HandleWrite(ctx, b2)
	assert.NotNil(t, err)
	assert.Equal(t, uint(5), b1.retryDelay)
	assert.Equal(t, 2, srv.retryQueue.list.Len())

	<-time.After(time.Millisecond * time.Duration(b1.retryDelay))
	b3 := NewBatch("3\n", opts.RetryInterval())
	err = srv.HandleWrite(ctx, b3)
	assert.NotNil(t, err)
	assert.Equal(t, uint(10), b1.retryDelay)
	assert.Equal(t, 3, srv.retryQueue.list.Len())
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

	b1 := NewBatch("1\n", opts.RetryInterval())
	err := srv.HandleWrite(ctx, b1)
	assert.NotNil(t, err)
	assert.Equal(t, uint(1), b1.retryDelay)
	assert.Equal(t, 1, srv.retryQueue.list.Len())

	<-time.After(time.Millisecond * time.Duration(b1.retryDelay))
	b2 := NewBatch("2\n", opts.RetryInterval())
	err = srv.HandleWrite(ctx, b2)
	assert.NotNil(t, err)
	assert.Equal(t, uint(5), b1.retryDelay)
	assert.Equal(t, 2, srv.retryQueue.list.Len())

	<-time.After(time.Millisecond * time.Duration(b1.retryDelay))
	b3 := NewBatch("3\n", opts.RetryInterval())
	err = srv.HandleWrite(ctx, b3)
	assert.NotNil(t, err)
	assert.Equal(t, uint(25), b1.retryDelay)
	assert.Equal(t, 3, srv.retryQueue.list.Len())

	// let write pass and it will clear queue
	<-time.After(time.Millisecond * time.Duration(b1.retryDelay))
	hs.SetReplyError(nil)
	err = srv.HandleWrite(ctx, NewBatch("4\n", opts.RetryInterval()))
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

	b1 := NewBatch("1\n", opts.RetryInterval())
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
		err = srv.HandleWrite(ctx, NewBatch(strings.Join(lines, "\n"), opts.RetryInterval()))
		wg.Done()
	}()
	cancel()
	wg.Wait()
	require.Equal(t, context.Canceled, err)
	assert.Len(t, hs.Lines(), 0)
}
