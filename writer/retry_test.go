package writer

import (
	"errors"
	"testing"
	"time"

	"github.com/influxdata/influxdb-client-go"
	"github.com/stretchr/testify/assert"
)

var (
	errSimple  = errors.New("something went wrong")
	errTooMany = func(retryAfter *int32) error {
		return &influxdb.Error{
			Code:       influxdb.ETooManyRequests,
			Message:    "too many requests",
			RetryAfter: retryAfter,
		}
	}
	one   int32 = 1
	two   int32 = 2
	three int32 = 3
)

func Test_RetryWriter_Write(t *testing.T) {
	for _, test := range []retryWriteCase{
		{
			name:    `one non-influxb.Error type error (mac attempts 3)`,
			options: []RetryOption{WithMaxAttempts(3)},
			metrics: createTestRowMetrics(t, 3),
			errors: []error{
				errSimple,
			},
			count: 0,
			err:   errSimple,
			writes: [][]influxdb.Metric{
				// one write attempt, immediate failure
				createTestRowMetrics(t, 3),
			},
		},
		{
			name:    `two "too many requests" errors (max attempts 3)`,
			options: []RetryOption{WithMaxAttempts(3)},
			metrics: createTestRowMetrics(t, 3),
			errors: []error{
				errTooMany(nil),
				errTooMany(nil),
			},
			count: 3,
			writes: [][]influxdb.Metric{
				// three writes, third succeeds
				createTestRowMetrics(t, 3),
				createTestRowMetrics(t, 3),
				createTestRowMetrics(t, 3),
			},
		},
		{
			name:    `one "unavailable" errors (max attempts 3)`,
			options: []RetryOption{WithMaxAttempts(3)},
			metrics: createTestRowMetrics(t, 3),
			errors: []error{
				&influxdb.Error{Code: influxdb.EUnavailable, Message: "too many requests"},
			},
			count: 3,
			writes: [][]influxdb.Metric{
				// two writes, second succeeds
				createTestRowMetrics(t, 3),
				createTestRowMetrics(t, 3),
			},
		},
		{
			name:    `three "too many requests" errors (max attempts 3)`,
			options: []RetryOption{WithMaxAttempts(3)},
			metrics: createTestRowMetrics(t, 3),
			errors: []error{
				errTooMany(nil),
				errTooMany(nil),
				errTooMany(nil),
			},
			count: 0,
			err:   errTooMany(nil),
			writes: [][]influxdb.Metric{
				// three writes all error
				createTestRowMetrics(t, 3),
				createTestRowMetrics(t, 3),
				createTestRowMetrics(t, 3),
			},
		},
		{
			name:    `three "too many requests" errors (max attempts 3) with retry-after`,
			options: []RetryOption{WithMaxAttempts(3)},
			metrics: createTestRowMetrics(t, 3),
			errors: []error{
				errTooMany(&one),
				errTooMany(&two),
				errTooMany(&three),
			},
			count: 0,
			err:   errTooMany(&three),
			writes: [][]influxdb.Metric{
				// three writes all error
				createTestRowMetrics(t, 3),
				createTestRowMetrics(t, 3),
				createTestRowMetrics(t, 3),
			},
			sleeps: []time.Duration{
				time.Second,
				2 * time.Second,
				3 * time.Second,
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
	sleeps []time.Duration
}

func (test *retryWriteCase) Run(t *testing.T) {
	var (
		writer = newTestWriter(test.errors...)
		retry  = NewRetryWriter(writer, test.options...)
		sleeps []time.Duration
	)

	retry.sleep = func(d time.Duration) {
		sleeps = append(sleeps, d)
	}

	count, err := retry.Write(test.metrics...)
	assert.Equal(t, test.err, err)
	assert.Equal(t, test.count, count)
	assert.Equal(t, test.writes, writer.writes)
	assert.Equal(t, test.sleeps, sleeps)
}
