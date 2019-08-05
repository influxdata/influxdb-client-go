package writer

import (
	"io"

	"github.com/influxdata/influxdb-client-go"
)

const defaultBufferSize = 100

// MetricsWriter is a type which metrics can be written to
type MetricsWriter interface {
	Write(...influxdb.Metric) (int, error)
}

// BufferedWriter is a buffered implementation of the MetricsWriter interface
// It is unashamedly derived from the bufio pkg https://golang.org/pkg/bufio
// Metrics are buffered up until the buffer size is met and then flushed to
// an underlying MetricsWriter
// The writer can also be flushed manually by calling Flush
// BufferedWriter is not safe to be called concurrently and therefore concurrency
// should be managed by the caller
type BufferedWriter struct {
	wr  MetricsWriter
	buf []influxdb.Metric
	n   int
	err error
}

// NewBufferedWriter returns a new *BufferedWriter with the default
// buffer size
func NewBufferedWriter(w MetricsWriter) *BufferedWriter {
	return NewBufferedWriterSize(w, defaultBufferSize)
}

// NewBufferedWriterSize returns a new *BufferedWriter with a buffer
// allocated with the provided size
func NewBufferedWriterSize(w MetricsWriter, size int) *BufferedWriter {
	if size <= 0 {
		size = defaultBufferSize
	}

	return &BufferedWriter{
		wr:  w,
		buf: make([]influxdb.Metric, size),
	}
}

// Available returns how many bytes are unused in the buffer.
func (b *BufferedWriter) Available() int { return len(b.buf) - b.n }

// Buffered returns the number of bytes that have been written into the current buffer.
func (b *BufferedWriter) Buffered() int { return b.n }

// Write writes the provided metrics to the underlying buffer if there is available
// capacity. Otherwise it flushes the buffer and attempts to assign the remain metrics to
// the buffer. This process repeats until all the metrics are either flushed or in the buffer
func (b *BufferedWriter) Write(m ...influxdb.Metric) (nn int, err error) {
	for len(m) > b.Available() && b.err == nil {
		var n int
		if b.Buffered() == 0 {
			// Large write, empty buffer.
			// Write directly from m to avoid copy.
			n, b.err = b.wr.Write(m...)
		} else {
			n = copy(b.buf[b.n:], m)
			b.n += n
			b.Flush()
		}

		nn += n
		m = m[n:]
	}

	if b.err != nil {
		return nn, b.err
	}

	n := copy(b.buf[b.n:], m)
	b.n += n
	nn += n
	return nn, nil
}

// Flush writes any buffered data to the underlying MetricsWriter
func (b *BufferedWriter) Flush() error {
	if b.err != nil {
		return b.err
	}

	if b.n == 0 {
		return nil
	}

	n, err := b.wr.Write(b.buf[0:b.n]...)
	if n < b.n && err == nil {
		err = io.ErrShortWrite
	}

	if err != nil {
		if n > 0 && n < b.n {
			copy(b.buf[0:b.n-n], b.buf[n:b.n])
		}
		b.n -= n
		b.err = err
		return err
	}

	b.n = 0

	return nil
}
