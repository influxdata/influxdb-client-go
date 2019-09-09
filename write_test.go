package influxdb

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var five = int32(5)

func Test_Client_Write(t *testing.T) {
	for _, test := range []struct {
		name string
		// api response
		body       []byte
		statusCode int
		headers    http.Header
		// expectations
		count int
		err   error
	}{
		{
			name:       "successful write",
			statusCode: http.StatusOK,
			count:      10,
		},
		{
			name:       "rate limited",
			statusCode: http.StatusTooManyRequests,
			headers: http.Header{
				"Retry-After": []string{"5"},
			},
			err: &Error{
				Code:       ETooManyRequests,
				Message:    "exceeded rate limit",
				RetryAfter: &five,
			},
		},
		{
			name:       "unavailable",
			statusCode: http.StatusServiceUnavailable,
			headers:    http.Header{},
			err: &Error{
				Code:    EUnavailable,
				Message: "service temporarily unavailable",
			},
		},
		{
			name:       "json encoded error",
			statusCode: http.StatusInternalServerError,
			body:       []byte(`{"code": "internal error", "op": "doing something", "message": "foo"}`),
			headers: http.Header{
				"Content-Type": []string{"application/json; charset=utf-8"},
			},
			err: &Error{
				Code:    EInternal,
				Op:      "doing something",
				Message: "foo",
			},
		},
		{
			name:       "plain text encoded error",
			statusCode: http.StatusBadRequest,
			body:       []byte(`payload is bad`),
			err: &Error{
				Code:    "400 Bad Request",
				Message: "payload is bad",
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			var (
				metrics = createTestRowMetrics(t, 10)
				server  = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					// path
					assert.Equal(t, "/api/v2/write", r.URL.Path)
					// headers
					assert.Equal(t, "Token token", r.Header.Get("Authorization"))
					assert.Equal(t, "text/plain; charset=utf-8", r.Header.Get("Content-Type"))
					assert.Equal(t, "gzip", r.Header.Get("Content-Encoding"))
					// params
					assert.Equal(t, "bucket", r.URL.Query().Get("bucket"))
					assert.Equal(t, "org", r.URL.Query().Get("org"))

					for k, v := range test.headers {
						w.Header()[k] = v
					}
					w.WriteHeader(test.statusCode)
					w.Write(test.body)
				}))
				client, err = New(server.URL, "token")
			)

			require.Nil(t, err)

			defer func() {
				require.Nil(t, client.Close())
				server.Close()
			}()

			count, err := client.Write(context.TODO(), "bucket", "org", metrics...)
			assert.Equal(t, test.err, err)
			assert.Equal(t, test.count, count)
		})
	}
}

func createTestRowMetrics(t *testing.T, count int) (metrics []Metric) {
	t.Helper()

	metrics = make([]Metric, 0, count)
	for i := 0; i < count; i++ {
		metrics = append(metrics, NewRowMetric(
			map[string]interface{}{
				"some_field": "some_value",
			},
			"some_measurement",
			map[string]string{
				"some_tag": "some_value",
			},
			time.Date(2019, time.January, 1, 0, 0, 0, 0, time.UTC),
		))
	}

	return
}
