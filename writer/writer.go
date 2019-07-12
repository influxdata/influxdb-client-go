package writer

import (
	"context"

	"github.com/influxdata/influxdb-client-go"
)

// BucketMetricWriter is a type which Metrics can be written to a particular bucket
// in a particular organisation
type BucketMetricWriter interface {
	Write(context.Context, influxdb.Organisation, influxdb.Bucket, ...influxdb.Metric) (int, error)
}

// New constructs a point writer with an underlying buffer from the provided BucketMetricWriter
// The writer will flushed metrics to the underlying BucketMetricWriter when the buffer is full
// or the configured flush interval ellapses without a flush occuring
func New(writer BucketMetricWriter, org influxdb.Organisation, bkt influxdb.Bucket, opts ...Option) *PointWriter {
	var (
		config   = Options(opts).Config()
		bucket   = NewBucketWriter(writer, org, bkt)
		buffered = NewBufferedWriterSize(bucket, config.size)
	)

	return NewPointWriter(buffered, config.flushInterval)
}

// BucketWriter writes metrics to a particular bucket
// within a particular organisation
type BucketWriter struct {
	w BucketMetricWriter

	ctxt context.Context

	org    influxdb.Organisation
	bucket influxdb.Bucket
}

// NewBucketWriter allocates, configures and returned a new BucketWriter for writing
// metrics to a specific organisations bucket
func NewBucketWriter(w BucketMetricWriter, org influxdb.Organisation, bkt influxdb.Bucket) *BucketWriter {
	return &BucketWriter{w, context.Background(), org, bkt}
}

// Write writes the provided metrics to the underlying metrics writer
// using the org and bucket configured on the bucket writer
func (b *BucketWriter) Write(m ...influxdb.Metric) (int, error) {
	return b.w.Write(b.ctxt, b.org, b.bucket, m...)
}
