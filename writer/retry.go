package writer

import "github.com/influxdata/influxdb-client-go"

const defaultMaxAttempts = 5

// RetryWriter is a metrics writers which decorates other
// metrics writer implementations and automatically retries
// attempts to write metrics under certain error conditions
type RetryWriter struct {
	MetricsWriter

	maxAttempts int
}

// NewRetryWriter returns a configured *RetryWriter which decorates
// the supplied MetricsWriter
func NewRetryWriter(w MetricsWriter, opts ...RetryOption) *RetryWriter {
	r := &RetryWriter{MetricsWriter: w, maxAttempts: defaultMaxAttempts}

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

		eerr, ok := err.(influxdb.Error)
		if !ok {
			break
		}

		switch eerr.Code {
		case influxdb.EUnavailable, influxdb.ETooManyRequests:
			// retriable errors
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
