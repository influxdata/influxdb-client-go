// Copyright 2020-2021 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package api

import (
	"context"
	"sync"
	"testing"
	"time"

	http2 "github.com/influxdata/influxdb-client-go/v2/api/http"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
	"github.com/influxdata/influxdb-client-go/v2/internal/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWritePoint(t *testing.T) {
	service := test.NewTestService(t, "http://localhost:8888")
	writeAPI := NewWriteAPIBlocking("my-org", "my-bucket", service, write.DefaultOptions().SetBatchSize(5))
	points := genPoints(10)
	err := writeAPI.WritePoint(context.Background(), points...)
	require.Nil(t, err)
	require.Len(t, service.Lines(), 10)
	for i, p := range points {
		line := write.PointToLineProtocol(p, writeAPI.writeOptions.Precision())
		//cut off last \n char
		line = line[:len(line)-1]
		assert.Equal(t, service.Lines()[i], line)
	}
}

func TestWriteRecord(t *testing.T) {
	service := test.NewTestService(t, "http://localhost:8888")
	writeAPI := NewWriteAPIBlocking("my-org", "my-bucket", service, write.DefaultOptions().SetBatchSize(5))
	lines := genRecords(10)
	err := writeAPI.WriteRecord(context.Background(), lines...)
	require.Nil(t, err)
	require.Len(t, service.Lines(), 10)
	for i, l := range lines {
		assert.Equal(t, l, service.Lines()[i])
	}
	service.Close()

	err = writeAPI.WriteRecord(context.Background())
	require.Nil(t, err)
	require.Len(t, service.Lines(), 0)

	service.SetReplyError(&http2.Error{Code: "invalid", Message: "data"})
	err = writeAPI.WriteRecord(context.Background(), lines...)
	require.NotNil(t, err)
	require.Equal(t, "invalid: data", err.Error())
}

func TestWriteContextCancel(t *testing.T) {
	service := test.NewTestService(t, "http://localhost:8888")
	writeAPI := NewWriteAPIBlocking("my-org", "my-bucket", service, write.DefaultOptions().SetBatchSize(5))
	lines := genRecords(10)
	ctx, cancel := context.WithCancel(context.Background())
	var err error
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		<-time.After(time.Second)
		err = writeAPI.WriteRecord(ctx, lines...)
		wg.Done()
	}()
	cancel()
	wg.Wait()
	require.Equal(t, context.Canceled, err)
	assert.Len(t, service.Lines(), 0)
}

func TestWriteParallel(t *testing.T) {
	service := test.NewTestService(t, "http://localhost:8888")
	writeAPI := NewWriteAPIBlocking("my-org", "my-bucket", service, write.DefaultOptions().SetBatchSize(5))
	lines := genRecords(1000)

	chanLine := make(chan string)
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			for l := range chanLine {
				err := writeAPI.WriteRecord(context.Background(), l)
				assert.Nil(t, err)
			}
			wg.Done()
		}()
	}
	for _, l := range lines {
		chanLine <- l
	}
	close(chanLine)
	wg.Wait()
	assert.Len(t, service.Lines(), len(lines))

	service.Close()
}
