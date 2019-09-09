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
func NewRetryWriter(w MetricsWriter) *RetryWriter {
	return &RetryWriter{MetricsWriter: w, maxAttempts: defaultMaxAttempts}
}

// Write delegates to underlying MetricsWriter and then
// automatically retries when errors occur
func (r *RetryWriter) Write(m ...influxdb.Metric) (int, error) {
	return len(m), influxdb.ErrUnimplemented
}
