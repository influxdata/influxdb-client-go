package writer

import (
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var deltaMsgFmt = "delta between flushes exceeds 105ms: %q"

func Test_PointWriter_Write_Batches(t *testing.T) {
	var (
		underlyingWriter = newTestWriter()
		writer           = NewPointWriter(NewBufferedWriter(underlyingWriter), 50*time.Millisecond)
		// 100 rows of batch size 100 expected
		expected = createNTestRowMetrics(t, 100, 100)
	)

	// write 10000 metrics in various batch sizes
	for _, batchSize := range permuteCounts(t, 10000) {
		n, err := writer.Write(createTestRowMetrics(t, batchSize)...)
		require.NoError(t, err)
		require.Equal(t, batchSize, n)

		time.Sleep(10 * time.Millisecond)
	}

	time.Sleep(100 * time.Millisecond)

	// close writer ensuring no more flushes occur
	require.NoError(t, writer.Close())

	// check batches written to underlying writer are 100 batches of 100 metrics
	require.Equal(t, expected, underlyingWriter.writes)
}

func Test_PointWriter_Write(t *testing.T) {
	var (
		underlyingWriter = newTestWriter()
		buffered         = NewBufferedWriterSize(underlyingWriter, 10)
		writer           = NewPointWriter(buffered, 100*time.Millisecond)
	)

	// write between 1 and 4 metrics every 30ms ideally meaning
	// some writes will flush for because of buffer size
	// and some for periodic flush
	for i := 0; i < 10; i++ {
		var (
			count  = rand.Intn(4) + 1
			n, err = writer.Write(createTestRowMetrics(t, count)...)
		)

		require.Nil(t, err)
		require.Equal(t, count, n)

		time.Sleep(30 * time.Millisecond)
	}

	// close writer to ensure scheduling has stopped
	writer.Close()

	// ensure time between each write does not exceed roughly 100ms
	// as per the flush interval of the point writer
	for i := 1; i < len(underlyingWriter.when); i++ {
		var (
			first, next = underlyingWriter.when[i-1], underlyingWriter.when[i]
			delta       = next.Sub(first)
		)

		// ensure writes are roughly 100 milliseconds apart
		assert.Truef(t, delta <= 105*time.Millisecond, deltaMsgFmt, delta)
	}
}
