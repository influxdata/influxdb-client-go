package influxdb

import (
	"compress/gzip"
	"context"
	"io/ioutil"
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
		// inputs
		metrics []Metric
		// api response
		body       []byte
		statusCode int
		headers    http.Header
		// expectations
		writtenPoints []byte
		count         int
		err           error
	}{
		{
			name:       "successful write",
			metrics:    createTestRowMetrics(t, 10),
			statusCode: http.StatusOK,
			count:      10,
		},
		{
			name: "successful write of explicit points",
			metrics: []Metric{
				NewRowMetric(
					map[string]interface{}{
						"some_int":  1,
						"some_uint": uint64(1),
					},
					"some_measurement",
					map[string]string{
						"some_tag": "some_value",
					},
					time.Date(2019, time.January, 1, 0, 0, 0, 0, time.UTC),
				),
				NewRowMetric(
					map[string]interface{}{
						"some_int":  2,
						"some_uint": uint64(2),
					},
					"some_measurement",
					map[string]string{
						"some_tag": "some_value",
					},
					time.Date(2019, time.January, 1, 0, 0, 0, 0, time.UTC),
				),
			},
			writtenPoints: []byte(`some_measurement,some_tag=some_value some_int=1i,some_uint=1u 1546300800000000000
some_measurement,some_tag=some_value some_int=2i,some_uint=2u 1546300800000000000
`),
			statusCode: http.StatusOK,
			count:      2,
		},
		{
			name:       "rate limited",
			metrics:    createTestRowMetrics(t, 10),
			statusCode: http.StatusTooManyRequests,
			headers: http.Header{
				"Retry-After": []string{"5"},
			},
			err: &Error{
				StatusCode: 429,
				Code:       ETooManyRequests,
				Message:    "exceeded rate limit",
				RetryAfter: &five,
			},
		},
		{
			name:       "unavailable",
			metrics:    createTestRowMetrics(t, 10),
			statusCode: http.StatusServiceUnavailable,
			headers:    http.Header{},
			err: &Error{
				StatusCode: 503,
				Code:       EUnavailable,
				Message:    "service temporarily unavailable",
			},
		},
		{
			name:       "json encoded error",
			metrics:    createTestRowMetrics(t, 10),
			statusCode: http.StatusInternalServerError,
			body:       []byte(`{"code": "internal error", "op": "doing something", "message": "foo"}`),
			headers: http.Header{
				"Content-Type": []string{"application/json; charset=utf-8"},
			},
			err: &Error{
				StatusCode: 500,
				Code:       EInternal,
				Op:         "doing something",
				Message:    "foo",
			},
		},
		{
			name:       "plain text encoded error",
			metrics:    createTestRowMetrics(t, 10),
			statusCode: http.StatusBadRequest,
			body:       []byte(`payload is bad`),
			err: &Error{
				StatusCode: 400,
				Code:       "400 Bad Request",
				Message:    "payload is bad",
			},
		},
		{
			name:       "size limited",
			metrics:    createTestRowMetrics(t, 10),
			statusCode: http.StatusRequestEntityTooLarge,
			headers:    http.Header{},
			err: &Error{
				StatusCode: 413,
				Code:       ETooLarge,
				Message:    "tried to write too large a batch",
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			var (
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

					if test.writtenPoints != nil {
						// check points written as expected
						reader, _ := gzip.NewReader(r.Body)
						data, _ := ioutil.ReadAll(reader)
						assert.Equal(t, test.writtenPoints, data)
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

			count, err := client.Write(context.TODO(), "bucket", "org", test.metrics...)
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
