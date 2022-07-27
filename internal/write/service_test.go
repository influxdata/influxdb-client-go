// Copyright 2020-2021 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package write

import (
	"context"
	"errors"
	"fmt"
	ilog "log"
	"runtime"
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
	// Set permanent reply error to force writes fail and retry
	hs.SetReplyError(&http.Error{
		StatusCode: 429,
	})
	// This batch will fail and it be added to retry queue
	b1 := NewBatch("1\n", opts.MaxRetryTime())
	err := srv.HandleWrite(ctx, b1)
	assert.NotNil(t, err)
	assert.EqualValues(t, 1, srv.retryDelay)
	assert.Equal(t, 1, srv.retryQueue.list.Len())

	//wait retry delay + little more
	<-time.After(time.Millisecond*time.Duration(srv.retryDelay) + time.Microsecond*5)
	// First batch will be tried to write again and this one will added to retry queue
	b2 := NewBatch("2\n", opts.MaxRetryTime())
	err = srv.HandleWrite(ctx, b2)
	assert.NotNil(t, err)
	assertBetween(t, srv.retryDelay, 2, 4)
	assert.Equal(t, 2, srv.retryQueue.list.Len())

	//wait retry delay + little more
	<-time.After(time.Millisecond*time.Duration(srv.retryDelay) + time.Microsecond*5)
	// First batch will be tried to write again and this one will added to retry queue
	b3 := NewBatch("3\n", opts.MaxRetryTime())
	err = srv.HandleWrite(ctx, b3)
	assert.NotNil(t, err)
	assertBetween(t, srv.retryDelay, 4, 8)
	assert.Equal(t, 3, srv.retryQueue.list.Len())

	//wait retry delay + little more
	<-time.After(time.Millisecond*time.Duration(srv.retryDelay) + time.Microsecond*5)
	// First batch will be tried to write again and this one will added to retry queue
	b4 := NewBatch("4\n", opts.MaxRetryTime())
	err = srv.HandleWrite(ctx, b4)
	assert.NotNil(t, err)
	assertBetween(t, srv.retryDelay, 8, 16)
	assert.Equal(t, 4, srv.retryQueue.list.Len())

	<-time.After(time.Millisecond*time.Duration(srv.retryDelay) + time.Microsecond*5)
	// Clear error and let write pass
	hs.SetReplyError(nil)
	// Batches from retry queue will be sent first
	err = srv.HandleWrite(ctx, NewBatch("5\n", opts.MaxRetryTime()))
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
	ilog.SetFlags(ilog.Ldate | ilog.Lmicroseconds)
	hs := test.NewTestService(t, "http://localhost:8086")
	// sleep takes at least more than 10ms (sometimes 15ms) on Windows https://github.com/golang/go/issues/44343
	baseRetryInterval := uint(1)
	if runtime.GOOS == "windows" {
		baseRetryInterval = 20
	}
	// Buffer limit 15000, bach is 5000 => buffer for 3 batches
	opts := write.DefaultOptions().SetRetryInterval(baseRetryInterval).SetRetryBufferLimit(15000)
	ctx := context.Background()
	srv := NewService("my-org", "my-bucket", hs, opts)
	// Set permanent reply error to force writes fail and retry
	hs.SetReplyError(&http.Error{
		StatusCode: 429,
	})
	// This batch will fail and it will be added to retry queue
	b1 := NewBatch("1\n", opts.MaxRetryTime())
	err := srv.HandleWrite(ctx, b1)
	assert.NotNil(t, err)
	//assert.Equal(t, uint(baseRetryInterval), srv.retryDelay)
	assertBetween(t, srv.retryDelay, baseRetryInterval, baseRetryInterval*2)
	assert.Equal(t, 1, srv.retryQueue.list.Len())

	<-time.After(time.Millisecond * time.Duration(srv.retryDelay))
	b2 := NewBatch("2\n", opts.MaxRetryTime())
	// First batch will be tried to write again and this one will added to retry queue
	err = srv.HandleWrite(ctx, b2)
	assert.NotNil(t, err)
	assertBetween(t, srv.retryDelay, baseRetryInterval*2, baseRetryInterval*4)
	assert.Equal(t, 2, srv.retryQueue.list.Len())

	<-time.After(time.Millisecond * time.Duration(srv.retryDelay))
	b3 := NewBatch("3\n", opts.MaxRetryTime())
	// First batch will be tried to write again and this one will added to retry queue
	err = srv.HandleWrite(ctx, b3)
	assert.NotNil(t, err)
	assertBetween(t, srv.retryDelay, baseRetryInterval*4, baseRetryInterval*8)
	assert.Equal(t, 3, srv.retryQueue.list.Len())

	// Write early and overwrite
	b4 := NewBatch("4\n", opts.MaxRetryTime())
	// No write will occur, because retry delay has not passed yet
	// However new bach will be added to retry queue. Retry queue has limit 3,
	// so first batch will be discarded
	priorRetryDelay := srv.retryDelay
	err = srv.HandleWrite(ctx, b4)
	assert.NoError(t, err)
	assert.Equal(t, priorRetryDelay, srv.retryDelay) // Accumulated retry delay should be retained despite batch discard
	assert.Equal(t, 3, srv.retryQueue.list.Len())

	// Overwrite
	<-time.After(time.Millisecond * time.Duration(srv.retryDelay) / 2)
	b5 := NewBatch("5\n", opts.MaxRetryTime())
	// Second batch will be tried to write again
	// However, write will fail and as new batch is added to retry queue
	// the second batch will be discarded
	err = srv.HandleWrite(ctx, b5)
	assert.Nil(t, err) // No error should be returned, because no write was attempted (still waiting for retryDelay to expire)
	assert.Equal(t, 3, srv.retryQueue.list.Len())

	<-time.After(time.Millisecond * time.Duration(srv.retryDelay))
	// Clear error and let write pass
	hs.SetReplyError(nil)
	// Batches from retry queue will be sent first
	err = srv.HandleWrite(ctx, NewBatch("6\n", opts.MaxRetryTime()))
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
	// MaxRetryInterval only 4ms, will be reached quickly
	opts := write.DefaultOptions().SetRetryInterval(1).SetMaxRetryInterval(4)
	ctx := context.Background()
	srv := NewService("my-org", "my-bucket", hs, opts)
	// Set permanent reply error to force writes fail and retry
	hs.SetReplyError(&http.Error{
		StatusCode: 503,
	})
	// This batch will fail and it be added to retry queue
	b1 := NewBatch("1\n", opts.MaxRetryTime())
	err := srv.HandleWrite(ctx, b1)
	assert.NotNil(t, err)
	assert.Equal(t, uint(1), srv.retryDelay)
	assert.Equal(t, 1, srv.retryQueue.list.Len())

	<-time.After(time.Millisecond * time.Duration(srv.retryDelay))
	b2 := NewBatch("2\n", opts.MaxRetryTime())
	// First batch will be tried to write again and this one will added to retry queue
	err = srv.HandleWrite(ctx, b2)
	assert.NotNil(t, err)
	assertBetween(t, srv.retryDelay, 2, 4)
	assert.Equal(t, 2, srv.retryQueue.list.Len())

	<-time.After(time.Millisecond * time.Duration(srv.retryDelay))
	b3 := NewBatch("3\n", opts.MaxRetryTime())
	// First batch will be tried to write again and this one will added to retry queue
	err = srv.HandleWrite(ctx, b3)
	assert.NotNil(t, err)
	// New computed delay of first batch should be 4-8, is limited to 4
	assert.EqualValues(t, 4, srv.retryDelay)
	assert.Equal(t, 3, srv.retryQueue.list.Len())

	<-time.After(time.Millisecond * time.Duration(srv.retryDelay))
	b4 := NewBatch("4\n", opts.MaxRetryTime())
	// First batch will be tried to write again and this one will added to retry queue
	err = srv.HandleWrite(ctx, b4)
	assert.NotNil(t, err)
	// New computed delay of first batch should be 8-116, is limited to 4
	assert.EqualValues(t, 4, srv.retryDelay)
	assert.Equal(t, 4, srv.retryQueue.list.Len())
}

func min(a, b uint) uint {
	if a > b {
		return b
	}
	return a
}

func TestMaxRetries(t *testing.T) {
	log.Log.SetLogLevel(log.DebugLevel)
	hs := test.NewTestService(t, "http://localhost:8086")
	opts := write.DefaultOptions().SetRetryInterval(1)
	ctx := context.Background()
	srv := NewService("my-org", "my-bucket", hs, opts)
	// Set permanent reply error to force writes fail and retry
	hs.SetReplyError(&http.Error{
		StatusCode: 429,
	})
	// This batch will fail and it be added to retry queue
	b1 := NewBatch("1\n", opts.MaxRetryTime())
	err := srv.HandleWrite(ctx, b1)
	assert.NotNil(t, err)
	assert.EqualValues(t, 1, srv.retryDelay)
	assert.Equal(t, 1, srv.retryQueue.list.Len())
	// Write so many batches as it is maxRetries (5)
	// First batch will be written and it will reach max retry limit
	for i, e := uint(1), uint(2); i <= opts.MaxRetries(); i++ {
		//wait retry delay + little more
		<-time.After(time.Millisecond*time.Duration(srv.retryDelay) + time.Microsecond*5)
		b := NewBatch(fmt.Sprintf("%d\n", i+1), opts.MaxRetryTime())
		err = srv.HandleWrite(ctx, b)
		assert.NotNil(t, err)
		assertBetween(t, srv.retryDelay, e, e*2)
		exp := min(i+1, opts.MaxRetries())
		assert.EqualValues(t, exp, srv.retryQueue.list.Len())
		e *= 2
	}
	//Test if was removed from retry queue
	assert.True(t, b1.Evicted)

	<-time.After(time.Millisecond*time.Duration(srv.retryDelay) + time.Microsecond*5)
	// Clear error and let write pass
	hs.SetReplyError(nil)
	// Batches from retry queue will be sent first
	err = srv.HandleWrite(ctx, NewBatch(fmt.Sprintf("%d\n", opts.MaxRetries()+2), opts.MaxRetryTime()))
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
	// Set maxRetryTime 5ms
	opts := write.DefaultOptions().SetRetryInterval(1).SetMaxRetryTime(5)
	ctx := context.Background()
	srv := NewService("my-org", "my-bucket", hs, opts)
	// Set permanent reply error to force writes fail and retry
	hs.SetReplyError(&http.Error{
		StatusCode: 429,
	})
	// This batch will fail and it be added to retry queue and it will expire 5ms after
	b1 := NewBatch("1\n", opts.MaxRetryTime())
	err := srv.HandleWrite(ctx, b1)
	assert.NotNil(t, err)
	assert.EqualValues(t, 1, srv.retryDelay)
	assert.Equal(t, 1, srv.retryQueue.list.Len())

	// Wait for batch expiration
	<-time.After(5 * time.Millisecond)

	exp := opts.MaxRetryTime()
	// sleep takes at least more than 10ms (sometimes 15ms) on Windows https://github.com/golang/go/issues/44343
	if runtime.GOOS == "windows" {
		exp = 20
	}
	// create new batch for sending
	b := NewBatch("2\n", exp)
	// First batch will  be checked against maxRetryTime and it will expire. New batch will fail and it will added to retry queue
	err = srv.HandleWrite(ctx, b)
	require.NotNil(t, err)
	// 1st Batch expires and writing 2nd trows error
	assert.Equal(t, "write failed (attempts 1): Unexpected status code 429", err.Error())
	assert.Equal(t, 1, srv.retryQueue.list.Len())

	//wait until remaining accumulated retryDelay has passed, because there hasn't been a successful write yet
	<-time.After(time.Until(srv.lastWriteAttempt.Add(time.Millisecond * time.Duration(srv.retryDelay))))
	// Clear error and let write pass
	hs.SetReplyError(nil)
	// A batch from retry queue will be sent first
	err = srv.HandleWrite(ctx, NewBatch("3\n", opts.MaxRetryTime()))
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

	// Set permanent non HTTP  error to force writes fail and retry
	hs.SetReplyError(&http.Error{
		Err: errors.New("connection refused"),
	})

	// This batch will fail and it be added to retry queue
	b1 := NewBatch("1\n", opts.MaxRetryTime())
	err := srv.HandleWrite(ctx, b1)
	assert.NotNil(t, err)
	assert.EqualValues(t, 1, srv.retryDelay)
	assert.Equal(t, 1, srv.retryQueue.list.Len())

	<-time.After(time.Millisecond * time.Duration(srv.retryDelay))

	b2 := NewBatch("2\n", opts.MaxRetryTime())
	// First batch will be tried to write again and this one will added to retry queue
	err = srv.HandleWrite(ctx, b2)
	assert.NotNil(t, err)
	assertBetween(t, srv.retryDelay, 2, 4)
	assert.Equal(t, 2, srv.retryQueue.list.Len())

	<-time.After(time.Millisecond * time.Duration(srv.retryDelay))

	b3 := NewBatch("3\n", opts.MaxRetryTime())
	// First batch will be tried to write again and this one will added to retry queue
	err = srv.HandleWrite(ctx, b3)
	assert.NotNil(t, err)
	assertBetween(t, srv.retryDelay, 4, 8)
	assert.Equal(t, 3, srv.retryQueue.list.Len())

	<-time.After(time.Millisecond * time.Duration(srv.retryDelay))
	// Clear error and let write pass
	hs.SetReplyError(nil)
	// Batches from retry queue will be sent first
	err = srv.HandleWrite(ctx, NewBatch("4\n", opts.MaxRetryTime()))
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

	b1 := NewBatch("1\n", opts.MaxRetryTime())
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
		err = srv.HandleWrite(ctx, NewBatch(strings.Join(lines, "\n"), opts.MaxRetryTime()))
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

func TestErrorCallback(t *testing.T) {
	log.Log.SetLogLevel(log.DebugLevel)
	hs := test.NewTestService(t, "http://localhost:8086")
	//
	opts := write.DefaultOptions().SetRetryInterval(1).SetRetryBufferLimit(15000)
	ctx := context.Background()
	srv := NewService("my-org", "my-bucket", hs, opts)

	hs.SetReplyError(&http.Error{
		Err: errors.New("connection refused"),
	})

	srv.SetBatchErrorCallback(func(batch *Batch, error2 http.Error) bool {
		return batch.RetryAttempts < 2
	})
	b1 := NewBatch("1\n", opts.MaxRetryTime())
	err := srv.HandleWrite(ctx, b1)
	assert.NotNil(t, err)
	assert.Equal(t, 1, srv.retryQueue.list.Len())

	<-time.After(time.Millisecond * time.Duration(srv.retryDelay))
	b := NewBatch("2\n", opts.MaxRetryTime())
	err = srv.HandleWrite(ctx, b)
	assert.NotNil(t, err)
	assert.Equal(t, 2, srv.retryQueue.list.Len())

	<-time.After(time.Millisecond * time.Duration(srv.retryDelay))
	b = NewBatch("3\n", opts.MaxRetryTime())
	err = srv.HandleWrite(ctx, b)
	assert.NotNil(t, err)
	assert.Equal(t, 2, srv.retryQueue.list.Len())

}

func minInt(a, b int) int {
	if a > b {
		return b
	}
	return a
}

func TestRetryIntervalAccumulation(t *testing.T) {
	// log.Log.SetLogLevel(log.DebugLevel)
	log.Log.SetLogLevel(log.InfoLevel)

	// Setup test service with scenario's configuration
	hs := test.NewTestService(t, "http://localhost:8086")
	baseRetryInterval := uint(20)
	if runtime.GOOS == "windows" {
		baseRetryInterval = 30
	}
	opts := write.DefaultOptions().
		SetRetryInterval(baseRetryInterval).
		SetMaxRetryInterval(300).
		SetMaxRetryTime(baseRetryInterval * 5)
	ctx := context.Background()
	srv := NewService("my-org", "my-bucket", hs, opts)
	writeInterval := time.Duration(opts.RetryInterval()) * time.Millisecond

	// Set permanent reply error to force writes fail and retry
	hs.SetReplyError(&http.Error{StatusCode: 429})

	lastInterval := uint(0)
	assert.Equal(t, uint(0), srv.retryAttempts) // Should initialize to zero
	i := 1
	for ; i <= 45; i++ {
		b := NewBatch(fmt.Sprintf("%d\n", i), opts.MaxRetryTime())
		err := srv.HandleWrite(ctx, b)
		assert.Equal(t, minInt(i, 5), srv.retryQueue.list.Len())
		assert.GreaterOrEqual(t, srv.retryDelay, lastInterval)         // Should not decrease while writes failing
		assert.LessOrEqual(t, srv.retryDelay, opts.MaxRetryInterval()) // Should not grow larger than max
		if err != nil {
			if lastInterval == opts.MaxRetryInterval() {
				// Write attempt failed, and interval was already at max, so should stay there
				assert.Equal(t, srv.retryDelay, opts.MaxRetryInterval())
				log.Log.Infof("Retry interval capped at %d ms", srv.retryDelay)
			} else {
				// A write attempt was made and failed, so retry interval should have increased
				assert.Greater(t, srv.retryDelay, lastInterval)
				log.Log.Infof("Retry interval increased to %d ms", srv.retryDelay)
			}
		} else {
			// Write attempt was not made, so retry interval should remain the same
			assert.Equal(t, srv.retryDelay, lastInterval)
			log.Log.Infof("Retry interval still at %d ms", srv.retryDelay)
		}
		lastInterval = srv.retryDelay

		<-time.After(writeInterval)
	}

	// Clear error and let write pass
	hs.SetReplyError(nil)

	// Wait until write queue is ready to retry; in meantime, keep writing and confirming queue state
	retryTimeout := srv.lastWriteAttempt.Add(time.Millisecond * time.Duration(srv.retryDelay))
	log.Log.Infof("Continuing to write for %d ms until flushing write attempt", time.Until(retryTimeout).Milliseconds())
	for ; time.Until(retryTimeout) >= 0; i++ {
		b := NewBatch(fmt.Sprintf("%d\n", i), opts.MaxRetryTime())
		err := srv.HandleWrite(ctx, b)
		assert.Nil(t, err) // There should be no write attempt
		assert.Equal(t, minInt(i, 5), srv.retryQueue.list.Len())
		assert.Equal(t, srv.retryDelay, opts.MaxRetryInterval()) // Should remain the same
		log.Log.Infof("Retry interval still at %d ms", srv.retryDelay)
		<-time.After(writeInterval)
	}

	// Retry interval should now have expired, so this write attempt should succeed and cause retry queue to flush
	b := NewBatch(fmt.Sprintf("%d\n", i), opts.MaxRetryTime())
	err := srv.HandleWrite(ctx, b)
	assert.Nil(t, err)
	assert.Equal(t, 0, srv.retryQueue.list.Len())
	assert.Equal(t, srv.retryAttempts, uint(0)) // Should reset to zero

	// Ensure proper batches got written to server
	require.Len(t, hs.Lines(), 5)
	assert.Equal(t, fmt.Sprintf("%d", i-4), hs.Lines()[0])
	assert.Equal(t, fmt.Sprintf("%d", i-3), hs.Lines()[1])
	assert.Equal(t, fmt.Sprintf("%d", i-2), hs.Lines()[2])
	assert.Equal(t, fmt.Sprintf("%d", i-1), hs.Lines()[3])
	assert.Equal(t, fmt.Sprintf("%d", i-0), hs.Lines()[4])

	// Debug line to capture output of successful test
	// assert.True(t, false)
}

func TestFlush(t *testing.T) {
	log.Log.SetLogLevel(log.DebugLevel)
	hs := test.NewTestService(t, "http://localhost:8086")
	//
	opts := write.DefaultOptions().SetRetryInterval(1)
	ctx := context.Background()
	srv := NewService("my-org", "my-bucket", hs, opts)

	hs.SetReplyError(&http.Error{
		Err: errors.New("connection refused"),
	})

	lines := test.GenRecords(5)
	// Test flush will fail all batches
	for _, line := range lines {
		b := NewBatch(line, 20)
		_ = srv.HandleWrite(ctx, b)
	}
	assert.Equal(t, 5, srv.retryQueue.list.Len())
	srv.Flush()
	assert.Len(t, hs.Lines(), 0)

	// Test flush will find all batches expired
	for _, line := range lines {
		b := NewBatch(line, 5)
		_ = srv.HandleWrite(ctx, b)
	}

	assert.Equal(t, 5, srv.retryQueue.list.Len())
	<-time.After(5 * time.Millisecond)

	hs.SetReplyError(nil)
	// all batches should expire
	srv.Flush()
	assert.Len(t, hs.Lines(), 0)
	assert.Equal(t, 0, srv.retryQueue.list.Len())

	// Test flush will succeed
	hs.SetReplyError(&http.Error{
		Err: errors.New("connection refused"),
	})
	for _, line := range lines {
		b := NewBatch(line, 5)
		_ = srv.HandleWrite(ctx, b)
	}

	assert.Equal(t, 5, srv.retryQueue.list.Len())
	hs.SetReplyError(nil)
	// all batches should expire
	srv.Flush()
	assert.Len(t, hs.Lines(), 5)
	assert.Equal(t, 0, srv.retryQueue.list.Len())
}
