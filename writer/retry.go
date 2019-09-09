package writer

import (
	"time"

	"github.com/influxdata/influxdb-client-go"
)

const defaultMaxAttempts = 5

// RetryWriter is a metrics writers which decorates other
// metrics writer implementations and automatically retries
// attempts to write metrics under certain error conditions
type RetryWriter struct {
	MetricsWriter

	sleep func(time.Duration)

	maxAttempts int
}

// NewRetryWriter returns a configured *RetryWriter which decorates
// the supplied MetricsWriter
func NewRetryWriter(w MetricsWriter, opts ...RetryOption) *RetryWriter {
	r := &RetryWriter{
		MetricsWriter: w,
		sleep:         time.Sleep,
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
			break
		}

		ierr, ok := err.(*influxdb.Error)
		if !ok {
			break
		}

		switch ierr.Code {
		// retriable errors
		case influxdb.EUnavailable, influxdb.ETooManyRequests:
			if ierr.RetryAfter != nil {
				// given retry-after is configured attempt to sleep
				// for retry-after seconds
				r.sleep(time.Duration(*ierr.RetryAfter) * time.Second)
			}
		default:
			break
		}
	}

	return
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
