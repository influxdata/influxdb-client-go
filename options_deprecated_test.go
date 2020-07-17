package influxdb2_test

import (
	"context"
	"crypto/tls"
	influxdb2 "github.com/influxdata/influxdb-client-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestDefaultOptionsDeprecated(t *testing.T) {
	opts := influxdb2.DefaultOptions()
	assert.Equal(t, uint(5000), opts.BatchSize())
	assert.Equal(t, false, opts.UseGZip())
	assert.Equal(t, uint(1000), opts.FlushInterval())
	assert.Equal(t, time.Nanosecond, opts.Precision())
	assert.Equal(t, uint(10000), opts.RetryBufferLimit())
	assert.Equal(t, uint(1000), opts.RetryInterval())
	assert.Equal(t, uint(3), opts.MaxRetries())
	assert.Equal(t, (*tls.Config)(nil), opts.TlsConfig())
	assert.Equal(t, uint(20), opts.HttpRequestTimeout())
	assert.Equal(t, uint(0), opts.LogLevel())
}

func TestSettingsOptionsDeprecated(t *testing.T) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}
	opts := influxdb2.DefaultOptions().
		SetBatchSize(5).
		SetUseGZip(true).
		SetFlushInterval(5000).
		SetPrecision(time.Millisecond).
		SetRetryBufferLimit(5).
		SetRetryInterval(5000).
		SetMaxRetries(7).
		SetTlsConfig(tlsConfig).
		SetHttpRequestTimeout(50).
		SetLogLevel(3).
		AddDefaultTag("t", "a")
	assert.Equal(t, uint(5), opts.BatchSize())
	assert.Equal(t, true, opts.UseGZip())
	assert.Equal(t, uint(5000), opts.FlushInterval())
	assert.Equal(t, time.Millisecond, opts.Precision())
	assert.Equal(t, uint(5), opts.RetryBufferLimit())
	assert.Equal(t, uint(5000), opts.RetryInterval())
	assert.Equal(t, uint(7), opts.MaxRetries())
	assert.Equal(t, tlsConfig, opts.TlsConfig())
	assert.Equal(t, uint(50), opts.HttpRequestTimeout())
	assert.Equal(t, uint(3), opts.LogLevel())
	assert.Len(t, opts.WriteOptions().DefaultTags(), 1)
}

func TestTimeoutDeprecated(t *testing.T) {
	response := `,result,table,_start,_stop,_time,_value,_field,_measurement,a,b,
		,,0,2020-02-17T22:19:49.747562847Z,2020-02-18T22:19:49.747562847Z,2020-02-18T10:34:08.135814545Z,1.4,f,test,1,adsfasdf
		,,0,2020-02-17T22:19:49.747562847Z,2020-02-18T22:19:49.747562847Z,2020-02-18T22:08:44.850214724Z,6.6,f,test,1,adsfasdf
		`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		if r.Method == http.MethodPost {
			w.Header().Set("Content-Type", "text/csv")
			w.WriteHeader(http.StatusOK)
			time.Sleep(2 * time.Second)
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
	client := influxdb2.NewClientWithOptions(server.URL, "a", influxdb2.DefaultOptions().SetHttpRequestTimeout(1))
	queryAPI := client.QueryApi("org")

	_, err := queryAPI.QueryRaw(context.Background(), "flux", nil)
	require.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "Client.Timeout exceeded"))

	client = influxdb2.NewClientWithOptions(server.URL, "a", influxdb2.DefaultOptions().SetHttpRequestTimeout(5))
	queryAPI = client.QueryApi("org")

	result, err := queryAPI.QueryRaw(context.Background(), "flux", nil)
	require.Nil(t, err)
	assert.Equal(t, response, result)
}
