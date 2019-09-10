package writer

import (
	"testing"
	"time"

	"github.com/influxdata/influxdb-client-go"
	"github.com/stretchr/testify/require"
)

func Test_New(t *testing.T) {
	var (
		client  = &influxdb.Client{}
		bkt     = "default"
		org     = "influx"
		options = Options{WithBufferSize(12), WithFlushInterval(5 * time.Minute), WithRetries()}
		wr      = New(client, bkt, org, options...)
	)

	require.Equal(t, 5*time.Minute, wr.flushInterval)
	require.Len(t, wr.w.(*BufferedWriter).buf, 12)

	require.Nil(t, wr.Close())
}

func Test_BucketWriter(t *testing.T) {
	var (
		spy    = &bucketWriter{}
		bucket = "default"
		org    = "influx"
		wr     = NewBucketWriter(spy, bucket, org)

		expected = []bucketWriteCall{
			{org, bucket, createTestRowMetrics(t, 4)},
			{org, bucket, createTestRowMetrics(t, 8)},
			{org, bucket, createTestRowMetrics(t, 12)},
			{org, bucket, createTestRowMetrics(t, 16)},
		}
	)

	for _, count := range []int{4, 8, 12, 16} {
		n, err := wr.Write(createTestRowMetrics(t, count)...)
		require.Nil(t, err)
		require.Equal(t, count, n)
	}

	// ensure underlying "client" is called as expected
	require.Equal(t, expected, spy.calls)
}
