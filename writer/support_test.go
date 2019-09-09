package writer

import (
	"context"
	"math"
	"math/rand"
	"testing"
	"time"

	"github.com/influxdata/influxdb-client-go"
)

type bucketWriter struct {
	calls []bucketWriteCall
}

type bucketWriteCall struct {
	bkt  string
	org  string
	data []influxdb.Metric
}

func (b *bucketWriter) Write(_ context.Context, bucket, org string, m ...influxdb.Metric) (int, error) {
	b.calls = append(b.calls, bucketWriteCall{bucket, org, m})
	return len(m), nil
}

type metricsWriter struct {
	// length override
	n int
	// metrics written
	when   []time.Time
	writes [][]influxdb.Metric
	// error response
	called int
	errs   []error
}

func newTestWriter(errs ...error) *metricsWriter {
	return &metricsWriter{n: -1, errs: errs}
}

func (w *metricsWriter) Write(m ...influxdb.Metric) (n int, err error) {
	defer func() { w.called++ }()

	w.when = append(w.when, time.Now())
	w.writes = append(w.writes, m)

	n = len(m)
	if w.n > -1 {
		// override length response
		n = w.n
	}

	if w.called < len(w.errs) {
		n = 0
		err = w.errs[w.called]
	}

	return
}

// permuteCounts returns a set of pseudo-random batch size counts
// which sum to the provided total
// E.g. for a sum total of 100 (permuteCounts(t, 100))
// this function may produce the following [5 12 8 10 14 11 9 9 8 14]
// The sum of these values is == 100 and there are √100 buckets
func permuteCounts(t *testing.T, total int) (buckets []int) {
	t.Helper()

	var accum int

	buckets = make([]int, int(math.Sqrt(float64(total))))
	for i := 0; i < len(buckets); i++ {
		size := total / len(buckets)
		if accum+size > total {
			size = total - accum
		}

		buckets[i] = size

		accum += size

		// shuffle some counts from previous bucket forward to current bucket
		if i > 0 {
			var (
				min   = math.Min(float64(buckets[i]), float64(buckets[i-1]))
				delta = rand.Intn(int(min / 2))
			)

			buckets[i-1], buckets[i] = buckets[i-1]-delta, buckets[i]+delta
		}
	}

	return
}

func createNTestRowMetrics(t *testing.T, rows, count int) (metrics [][]influxdb.Metric) {
	metrics = make([][]influxdb.Metric, 0, rows)

	for i := 0; i < rows; i++ {
		metrics = append(metrics, createTestRowMetrics(t, count))
	}

	return
}

func createTestRowMetrics(t *testing.T, count int) (metrics []influxdb.Metric) {
	t.Helper()

	metrics = make([]influxdb.Metric, 0, count)
	for i := 0; i < count; i++ {
		metrics = append(metrics, influxdb.NewRowMetric(
			map[string]interface{}{
				"some_field": "some_value",
			},
			"some_measurement",
			map[string]string{
				"some_tag": "some_value",
			},
			time.Date(2019, time.January, 1, 0, 0, 0, 0, time.UTC),
		))
	}

	return
}
