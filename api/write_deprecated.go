package api

import (
	"context"

	"github.com/influxdata/influxdb-client-go/api/write"
)

// WriteApi is Write client interface with non-blocking methods for writing time series data asynchronously in batches into an InfluxDB server.
// Deprecated: Use WriteAPI instead.
type WriteApi interface {
	// WriteRecord writes asynchronously line protocol record into bucket.
	// WriteRecord adds record into the buffer which is sent on the background when it reaches the batch size.
	// Blocking alternative is available in the WriteApiBlocking interface
	WriteRecord(line string)
	// WritePoint writes asynchronously Point into bucket.
	// WritePoint adds Point into the buffer which is sent on the background when it reaches the batch size.
	// Blocking alternative is available in the WriteApiBlocking interface
	WritePoint(point *write.Point)
	// Flush forces all pending writes from the buffer to be sent
	Flush()
	// Flushes all pending writes and stop async processes. After this the Write client cannot be used
	Close()
	// Errors returns a channel for reading errors which occurs during async writes.
	// Must be called before performing any writes for errors to be collected.
	// The chan is unbuffered and must be drained or the writer will block.
	Errors() <-chan error
}

// WriteApiBlocking offers blocking methods for writing time series data synchronously into an InfluxDB server.
// Deprecated: use WriteAPIBlocking instead.
type WriteApiBlocking interface {
	// WriteRecord writes line protocol record(s) into bucket.
	// WriteRecord writes without implicit batching. Batch is created from given number of records
	// Non-blocking alternative is available in the WriteApi interface
	WriteRecord(ctx context.Context, line ...string) error
	// WritePoint data point into bucket.
	// WritePoint writes without implicit batching. Batch is created from given number of points
	// Non-blocking alternative is available in the WriteApi interface
	WritePoint(ctx context.Context, point ...*write.Point) error
}
