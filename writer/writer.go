package writer

import (
	"context"

	"github.com/influxdata/influxdb-client-go"
)

// BucketMetricWriter is a type which Metrics can be written to a particular bucket
// in a particular organisation
type BucketMetricWriter interface {
	Write(ctx context.Context, bucket string, org string, m ...influxdb.Metric) (int, error)
}

// New constructs a point writer with an underlying buffer from the provided BucketMetricWriter
// The writer will flushed metrics to the underlying BucketMetricWriter when the buffer is full
// or the configured flush interval ellapses without a flush occuring
func New(writer BucketMetricWriter, bkt, org string, opts ...Option) *PointWriter {
	var (
		config   = Options(opts).Config()
		bucket   = NewBucketWriter(writer, bkt, org)
		buffered = NewBufferedWriterSize(bucket, config.size)
	)

	// set bucket write context to provided context
	bucket.ctxt = config.ctxt

	if config.retry {
		// configure automatic retries for transient errors
		retry := NewRetryWriter(bucket, config.retryOptions...)
		buffered = NewBufferedWriterSize(retry, config.size)
	}

	return NewPointWriter(buffered, config.flushInterval)
}

// BucketWriter writes metrics to a particular bucket
// within a particular organisation
type BucketWriter struct {
	w BucketMetricWriter

	ctxt context.Context

	bucket string
	org    string
}

// NewBucketWriter allocates, configures and returned a new BucketWriter for writing
// metrics to a specific organisations bucket
func NewBucketWriter(w BucketMetricWriter, bucket, org string) *BucketWriter {
	return &BucketWriter{w, context.Background(), bucket, org}
}

// Write writes the provided metrics to the underlying metrics writer
// using the org and bucket configured on the bucket writer
func (b *BucketWriter) Write(m ...influxdb.Metric) (int, error) {
	return b.w.Write(b.ctxt, b.bucket, b.org, m...)
}
