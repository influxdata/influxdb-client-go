package influxdb2

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestTimeout(t *testing.T) {
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
	client := NewClientWithOptions(server.URL, "a", DefaultOptions().SetHttpRequestTimeout(1))
	queryApi := client.QueryApi("org")

	_, err := queryApi.QueryRaw(context.Background(), "flux", nil)
	require.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "Client.Timeout exceeded"))

	client = NewClientWithOptions(server.URL, "a", DefaultOptions().SetHttpRequestTimeout(5))
	queryApi = client.QueryApi("org")

	result, err := queryApi.QueryRaw(context.Background(), "flux", nil)
	require.Nil(t, err)
	assert.Equal(t, response, result)

}
