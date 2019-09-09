package writer

import (
	"time"

	"github.com/influxdata/influxdb-client-go"
)

const defaultMaxAttempts = 5

// BackoffFunc is a function which when called with an
// attempt number returns a duration which should be
// waited for until a subsequent attempt is made
type BackoffFunc func(attempt int) time.Duration

// RetryWriter is a metrics writers which decorates other
// metrics writer implementations and automatically retries
// attempts to write metrics under certain error conditions
type RetryWriter struct {
	MetricsWriter

	sleep   func(time.Duration)
	backoff BackoffFunc

	maxAttempts int
}

// NewRetryWriter returns a configured *RetryWriter which decorates
// the supplied MetricsWriter
func NewRetryWriter(w MetricsWriter, opts ...RetryOption) *RetryWriter {
	r := &RetryWriter{
		MetricsWriter: w,
		sleep:         time.Sleep,
		backoff:       func(int) time.Duration { return 0 },
		maxAttempts:   defaultMaxAttempts,
	}

	for _, opt := range opts {
		opt(r)
	}

	return r
}

// Write delegates to underlying MetricsWriter and then
// automatically retries when errors occur
func (r *RetryWriter) Write(m ...influxdb.Metric) (n int, err error) {
	for i := 0; i < r.maxAttempts; i++ {
		n, err = r.MetricsWriter.Write(m...)
		if err == nil {
			return
		}

		ierr, ok := err.(*influxdb.Error)
		if !ok {
			return
		}

		switch ierr.Code {
		// retriable errors
		case influxdb.EUnavailable, influxdb.ETooManyRequests:
			if ierr.RetryAfter != nil {
				// given retry-after is configured attempt to sleep
				// for retry-after seconds
				r.sleep(time.Duration(*ierr.RetryAfter) * time.Second)
				continue
			}

			// given a backoff duration > 0
			if duration := r.backoff(i + 1); duration > 0 {
				// call sleep with backoff duration
				r.sleep(duration)
			}
		default:
			return
		}
	}

	return
}

// LinearBackoff returns a BackoffFunc which when called
// returns attempt * scale.
// e.g.
// LinearBackoff(time.Second)(5) returns 5 seconds
func LinearBackoff(scale time.Duration) BackoffFunc {
	return func(attempt int) time.Duration {
		return time.Duration(attempt) * scale
	}
}

// RetryOption is a functional option for the RetryWriters type
type RetryOption func(*RetryWriter)

// WithMaxAttempts sets the maximum number of attempts for a
// Write operation attempt
func WithMaxAttempts(maxAttempts int) RetryOption {
	return func(r *RetryWriter) {
		r.maxAttempts = maxAttempts
	}
}

// WithBackoff sets of the BackoffFunc on the RetryWriter
func WithBackoff(fn BackoffFunc) RetryOption {
	return func(r *RetryWriter) {
		r.backoff = fn
	}
}
