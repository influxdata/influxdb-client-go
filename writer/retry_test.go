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
	three int32 = 3
	four  int32 = 4
	five  int32 = 5
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
			name:    `one "internal error" error (max attempts 3)`,
			options: []RetryOption{WithMaxAttempts(3)},
			metrics: createTestRowMetrics(t, 3),
			errors: []error{
				&influxdb.Error{Code: influxdb.EInternal, Message: "something went wrong"},
			},
			count: 0,
			err:   &influxdb.Error{Code: influxdb.EInternal, Message: "something went wrong"},
			writes: [][]influxdb.Metric{
				// one attempted write
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
				errTooMany(&three),
				errTooMany(&three),
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
				3 * time.Second,
				3 * time.Second,
				3 * time.Second,
			},
		},
		{
			name: `three "too many requests" errors (max attempts 3) with backoff`,
			options: []RetryOption{
				WithMaxAttempts(3),
				WithBackoff(LinearBackoff(time.Millisecond)),
			},
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
			sleeps: []time.Duration{
				1 * time.Millisecond,
				2 * time.Millisecond,
				3 * time.Millisecond,
			},
		},
		{
			name: `three "too many requests" errors (max attempts 3) with backoff and retry-after`,
			options: []RetryOption{
				WithMaxAttempts(3),
				WithBackoff(LinearBackoff(time.Millisecond)),
			},
			metrics: createTestRowMetrics(t, 3),
			errors: []error{
				errTooMany(&three),
				errTooMany(&four),
				errTooMany(&five),
			},
			count: 0,
			err:   errTooMany(&five),
			writes: [][]influxdb.Metric{
				// three writes all error
				createTestRowMetrics(t, 3),
				createTestRowMetrics(t, 3),
				createTestRowMetrics(t, 3),
			},
			sleeps: []time.Duration{
				3 * time.Second,
				4 * time.Second,
				5 * time.Second,
			},
		},
		{
			name: `three "too many requests" errors (max attempts 3) with retry-after and retry limit`,
			options: []RetryOption{
				WithMaxAttempts(3),
				WithRetrySleepLimit(4),
			},
			metrics: createTestRowMetrics(t, 3),
			errors: []error{
				errTooMany(&three),
				errTooMany(&three),
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
				3 * time.Second,
				1 * time.Second,
			},
		},
		{
			name: `three "too many requests" errors (max attempts 3) with backoff and retry limit`,
			options: []RetryOption{
				WithMaxAttempts(3),
				WithBackoff(LinearBackoff(time.Second)),
				WithRetrySleepLimit(4),
			},
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
			sleeps: []time.Duration{
				1 * time.Second,
				2 * time.Second,
				1 * time.Second,
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

	now := time.Now()

	retry.sleep = func(d time.Duration) {
		sleeps = append(sleeps, d)
		now = now.Add(d)
	}

	retry.now = func() time.Time {
		return now
	}

	count, err := retry.Write(test.metrics...)
	assert.Equal(t, test.err, err)
	assert.Equal(t, test.count, count)
	assert.Equal(t, test.writes, writer.writes)
	assert.Equal(t, test.sleeps, sleeps)
}
