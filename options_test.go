// Copyright 2020-2021 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package influxdb2_test

import (
	"context"
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultOptions(t *testing.T) {
	opts := influxdb2.DefaultOptions()
	assert.Equal(t, uint(5000), opts.BatchSize())
	assert.Equal(t, false, opts.UseGZip())
	assert.Equal(t, uint(1000), opts.FlushInterval())
	assert.Equal(t, time.Nanosecond, opts.Precision())
	assert.Equal(t, uint(50000), opts.RetryBufferLimit())
	assert.Equal(t, uint(5000), opts.RetryInterval())
	assert.Equal(t, uint(3), opts.MaxRetries())
	assert.Equal(t, uint(300000), opts.MaxRetryInterval())
	assert.Equal(t, (*tls.Config)(nil), opts.TLSConfig())
	assert.Equal(t, uint(20), opts.HTTPRequestTimeout())
	assert.Equal(t, uint(0), opts.LogLevel())
}

func TestSettingsOptions(t *testing.T) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}
	opts := influxdb2.DefaultOptions().
		SetBatchSize(5).
		SetUseGZip(true).
		SetFlushInterval(5000).
		SetPrecision(time.Millisecond).
		SetRetryBufferLimit(5).
		SetRetryInterval(1000).
		SetMaxRetryInterval(10000).
		SetMaxRetries(7).
		SetTLSConfig(tlsConfig).
		SetHTTPRequestTimeout(50).
		SetLogLevel(3).
		AddDefaultTag("t", "a")
	assert.Equal(t, uint(5), opts.BatchSize())
	assert.Equal(t, true, opts.UseGZip())
	assert.Equal(t, uint(5000), opts.FlushInterval())
	assert.Equal(t, time.Millisecond, opts.Precision())
	assert.Equal(t, uint(5), opts.RetryBufferLimit())
	assert.Equal(t, uint(1000), opts.RetryInterval())
	assert.Equal(t, uint(10000), opts.MaxRetryInterval())
	assert.Equal(t, uint(7), opts.MaxRetries())
	assert.Equal(t, tlsConfig, opts.TLSConfig())
	assert.Equal(t, uint(50), opts.HTTPRequestTimeout())
	if client := opts.HTTPClient(); assert.NotNil(t, client) {
		assert.Equal(t, 50*time.Second, client.Timeout)
		assert.Equal(t, tlsConfig, client.Transport.(*http.Transport).TLSClientConfig)
	}
	assert.Equal(t, uint(3), opts.LogLevel())
	assert.Len(t, opts.WriteOptions().DefaultTags(), 1)

	client := &http.Client{
		Transport: &http.Transport{},
	}
	opts.SetHTTPClient(client)
	assert.Equal(t, client, opts.HTTPClient())
}

func TestTimeout(t *testing.T) {
	response := `,result,table,_start,_stop,_time,_value,_field,_measurement,a,b,
		,,0,2020-02-17T22:19:49.747562847Z,2020-02-18T22:19:49.747562847Z,2020-02-18T10:34:08.135814545Z,1.4,f,test,1,adsfasdf
		,,0,2020-02-17T22:19:49.747562847Z,2020-02-18T22:19:49.747562847Z,2020-02-18T22:08:44.850214724Z,6.6,f,test,1,adsfasdf
		`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-time.After(100 * time.Millisecond)
		if r.Method == http.MethodPost {
			w.Header().Set("Content-Type", "text/csv")
			w.WriteHeader(http.StatusOK)
			<-time.After(2 * time.Second)
			_, err := w.Write([]byte(response))
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(err.Error()))
			}
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()
	client := influxdb2.NewClientWithOptions(server.URL, "a", influxdb2.DefaultOptions().SetHTTPRequestTimeout(1))
	queryAPI := client.QueryAPI("org")

	_, err := queryAPI.QueryRaw(context.Background(), "flux", nil)
	require.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "Client.Timeout exceeded"))

	client = influxdb2.NewClientWithOptions(server.URL, "a", influxdb2.DefaultOptions().SetHTTPRequestTimeout(5))
	queryAPI = client.QueryAPI("org")

	result, err := queryAPI.QueryRaw(context.Background(), "flux", nil)
	require.Nil(t, err)
	assert.Equal(t, response, result)

}
