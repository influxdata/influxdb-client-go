package writer

import (
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_BufferedWriter(t *testing.T) {
	var (
		// writer which asserts calls are made for org and bucket
		underlyingWriter = newTestWriter()
		writer           = NewBufferedWriter(underlyingWriter)
		// 100 rows of batch size 100 expected
		expected = createNTestRowMetrics(t, 100, 100)
	)

	// write 10000 metrics in various batch sizes
	for _, batchSize := range permuteCounts(t, 10000) {
		n, err := writer.Write(createTestRowMetrics(t, batchSize)...)
		require.NoError(t, err)
		require.Equal(t, batchSize, n)
	}

	// flush any remaining buffer to underlying writer
	require.Zero(t, writer.Available())
	require.Equal(t, 100, writer.Buffered())
	require.NoError(t, writer.Flush())

	// check batches written to underlying writer are 100 batches of 100 metrics
	require.Equal(t, expected, underlyingWriter.writes)
	require.Zero(t, writer.Buffered())
	require.Equal(t, 100, writer.Available())
}

func Test_BufferedWriter_LargeWriteEmptyBuffer(t *testing.T) {
	var (
		// writer which asserts calls are made for org and bucket
		underlyingWriter = newTestWriter()
		writer           = NewBufferedWriterSize(underlyingWriter, 100)
		expected         = createNTestRowMetrics(t, 1, 500)
	)

	n, err := writer.Write(createTestRowMetrics(t, 500)...)
	require.Nil(t, err)
	require.Equal(t, 500, n)

	// expect one large batch of 500 metrics to skip buffer and
	// go straight to underyling writer
	require.Equal(t, expected, underlyingWriter.writes)
}

func TestBufferedWriter_ShortWrite(t *testing.T) {
	var (
		// writer which reports 9 bytes written
		underlyingWriter = &metricsWriter{n: 9}
		writer           = NewBufferedWriterSize(underlyingWriter, 10)
		metrics          = createTestRowMetrics(t, 11)
	)

	// warm the buffer otherwise we skip it and go
	// straight to client
	n, err := writer.Write(metrics[0])
	require.Nil(t, err)
	require.Equal(t, 1, n)

	// attempt write where underyling write only writes 6 metrics
	n, err = writer.Write(metrics[1:]...)
	require.Equal(t, io.ErrShortWrite, err)
	require.Equal(t, 9, n)
}
