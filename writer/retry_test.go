package writer

import (
	"testing"

	"github.com/influxdata/influxdb-client-go"
	"github.com/stretchr/testify/assert"
)

func Test_RetryWriter_Write(t *testing.T) {
	for _, test := range []retryWriteCase{
		{
			name:    `two "too many requests" errors (max attempts 3)`,
			options: []RetryOption{WithMaxAttempts(3)},
			metrics: createTestRowMetrics(t, 10),
			errors: []error{
				influxdb.Error{Code: influxdb.ETooManyRequests, Message: "too many requests"},
				influxdb.Error{Code: influxdb.ETooManyRequests, Message: "too many requests"},
			},
			count: 10,
			writes: [][]influxdb.Metric{
				createTestRowMetrics(t, 10),
				createTestRowMetrics(t, 10),
			},
		},
		{
			name:    `two "unavailable" errors (max attempts 3)`,
			options: []RetryOption{WithMaxAttempts(3)},
			metrics: createTestRowMetrics(t, 10),
			errors: []error{
				influxdb.Error{Code: influxdb.EUnavailable, Message: "too many requests"},
				influxdb.Error{Code: influxdb.EUnavailable, Message: "too many requests"},
			},
			count: 10,
			writes: [][]influxdb.Metric{
				createTestRowMetrics(t, 10),
				createTestRowMetrics(t, 10),
			},
		},
		{
			name:    `three "too many requests" errors (max attempts 3)`,
			options: []RetryOption{WithMaxAttempts(3)},
			metrics: createTestRowMetrics(t, 10),
			errors: []error{
				influxdb.Error{Code: influxdb.ETooManyRequests, Message: "too many requests"},
				influxdb.Error{Code: influxdb.ETooManyRequests, Message: "too many requests"},
				influxdb.Error{Code: influxdb.ETooManyRequests, Message: "too many requests"},
			},
			count: 0,
			err:   influxdb.Error{Code: influxdb.ETooManyRequests, Message: "too many requests"},
			writes: [][]influxdb.Metric{
				createTestRowMetrics(t, 10),
				createTestRowMetrics(t, 10),
				createTestRowMetrics(t, 10),
			},
		},
	} {
		t.Run(test.name, test.Run)
	}
}

type retryWriteCase struct {
	name string
	// inputs
	options []RetryOption
	metrics []influxdb.Metric
	// errors returned by write
	errors []error
	// response expectations
	count int
	err   error
	// attempted writes
	writes [][]influxdb.Metric
}

func (test *retryWriteCase) Run(t *testing.T) {
	var (
		writer     = newTestWriter()
		retry      = NewRetryWriter(writer, test.options...)
		count, err = retry.Write(test.metrics...)
	)

	assert.Equal(t, test.err, err)
	assert.Equal(t, test.count, count)
	assert.Equal(t, test.writes, writer.writes)
}
