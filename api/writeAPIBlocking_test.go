// Copyright 2020-2021 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package api

import (
	"context"
	"net"
	"net/http"
	"strings"
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
	opts := write.DefaultOptions().SetBatchSize(5)
	writeAPI := NewWriteAPIBlocking("my-org", "my-bucket", service, opts)
	points := test.GenPoints(10)
	err := writeAPI.WritePoint(context.Background(), points...)
	require.Nil(t, err)
	require.Len(t, service.Lines(), 10)
	for i, p := range points {
		line := write.PointToLineProtocol(p, opts.Precision())
		//cut off last \n char
		line = line[:len(line)-1]
		assert.Equal(t, service.Lines()[i], line)
	}
}

func TestWriteRecord(t *testing.T) {
	service := test.NewTestService(t, "http://localhost:8888")
	writeAPI := NewWriteAPIBlocking("my-org", "my-bucket", service, write.DefaultOptions().SetBatchSize(5))
	lines := test.GenRecords(10)
	for _, line := range lines {
		err := writeAPI.WriteRecord(context.Background(), line)
		require.Nil(t, err)
	}
	require.Len(t, service.Lines(), 10)
	require.Equal(t, 10, service.Requests())
	for i, l := range lines {
		assert.Equal(t, l, service.Lines()[i])
	}
	service.Close()

	err := writeAPI.WriteRecord(context.Background(), lines...)
	require.Nil(t, err)
	require.Equal(t, 1, service.Requests())
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

func TestWriteRecordBatch(t *testing.T) {
	service := test.NewTestService(t, "http://localhost:8888")
	writeAPI := NewWriteAPIBlocking("my-org", "my-bucket", service, write.DefaultOptions().SetBatchSize(5))
	lines := test.GenRecords(10)
	batch := strings.Join(lines, "\n")
	err := writeAPI.WriteRecord(context.Background(), batch)
	require.Nil(t, err)
	require.Len(t, service.Lines(), 10)
	for i, l := range lines {
		assert.Equal(t, l, service.Lines()[i])
	}
	service.Close()
}

func TestWriteParallel(t *testing.T) {
	service := test.NewTestService(t, "http://localhost:8888")
	writeAPI := NewWriteAPIBlocking("my-org", "my-bucket", service, write.DefaultOptions().SetBatchSize(5))
	lines := test.GenRecords(1000)

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

func TestWriteErrors(t *testing.T) {
	service := http2.NewService("http://locl:866", "", http2.DefaultOptions().SetHTTPClient(&http.Client{
		Timeout: 100 * time.Millisecond,
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout: 100 * time.Millisecond,
			}).DialContext,
		},
	}))
	writeAPI := NewWriteAPIBlocking("my-org", "my-bucket", service, write.DefaultOptions().SetBatchSize(5))
	points := test.GenPoints(10)
	errors := 0
	for _, p := range points {
		err := writeAPI.WritePoint(context.Background(), p)
		if assert.Error(t, err) {
			errors++
		}
	}
	require.Equal(t, 10, errors)

}

func TestWriteBatchIng(t *testing.T) {
	service := test.NewTestService(t, "http://localhost:8888")
	writeAPI := NewWriteAPIBlockingWithBatching("my-org", "my-bucket", service, write.DefaultOptions().SetBatchSize(5))
	lines := test.GenRecords(10)
	for i, line := range lines {
		err := writeAPI.WriteRecord(context.Background(), line)
		require.Nil(t, err)
		if i == 4 || i == 9 {
			assert.Equal(t, 1, service.Requests())
			require.Len(t, service.Lines(), 5)

			service.Close()
		}
	}
}
