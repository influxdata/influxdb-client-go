// Copyright 2020 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package influxdb2

import (
	"context"
	"github.com/influxdata/influxdb-client-go/internal/http"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sync"
	"testing"
	"time"
)

func TestWritePoint(t *testing.T) {
	client := &client{serverUrl: "http://locahost:4444", options: DefaultOptions()}
	service := &testHttpService{
		t:         t,
		options:   client.Options(),
		serverUrl: client.ServerUrl(),
	}
	client.options.SetBatchSize(5)
	writeApi := newWriteApiBlockingImpl("my-org", "my-bucket", service, client)
	points := genPoints(10)
	err := writeApi.WritePoint(context.Background(), points...)
	require.Nil(t, err)
	require.Len(t, service.lines, 10)
	for i, p := range points {
		line := p.ToLineProtocol(client.options.Precision())
		//cut off last \n char
		line = line[:len(line)-1]
		assert.Equal(t, service.lines[i], line)
	}
}

func TestWriteRecord(t *testing.T) {
	client := &client{serverUrl: "http://locahost:4444", options: DefaultOptions()}
	service := &testHttpService{
		t:         t,
		options:   client.Options(),
		serverUrl: client.ServerUrl(),
	}
	client.options.SetBatchSize(5)
	writeApi := newWriteApiBlockingImpl("my-org", "my-bucket", service, client)
	lines := genRecords(10)
	err := writeApi.WriteRecord(context.Background(), lines...)
	require.Nil(t, err)
	require.Len(t, service.lines, 10)
	for i, l := range lines {
		assert.Equal(t, l, service.lines[i])
	}
	service.Close()

	err = writeApi.WriteRecord(context.Background())
	require.Nil(t, err)
	require.Len(t, service.lines, 0)

	service.replyError = &http.Error{Code: "invalid", Message: "data"}
	err = writeApi.WriteRecord(context.Background(), lines...)
	require.NotNil(t, err)
	require.Equal(t, "invalid: data", err.Error())
}

func TestWriteContextCancel(t *testing.T) {
	client := &client{serverUrl: "http://locahost:4444", options: DefaultOptions()}
	service := &testHttpService{
		t:         t,
		options:   client.Options(),
		serverUrl: client.ServerUrl(),
	}
	client.options.SetBatchSize(5)
	writeApi := newWriteApiBlockingImpl("my-org", "my-bucket", service, client)
	lines := genRecords(10)
	ctx, cancel := context.WithCancel(context.Background())
	var err error
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		time.Sleep(time.Second)
		err = writeApi.WriteRecord(ctx, lines...)
		wg.Done()
	}()
	cancel()
	wg.Wait()
	require.Equal(t, context.Canceled, err)
	assert.Len(t, service.lines, 0)
}
