package influxclient

import (
	"github.com/influxdata/line-protocol/v2/lineprotocol"
	"time"
)

// RetryParams configures retry behavior used by PointsWriter
type RetryParams struct {
	// Default retry interval in ms, if not sent by server. Default 5,000.
	RetryInterval int
	// Maximum count of retry attempts of failed writes, default 5.
	MaxRetries int
	// Maximum number of points to keep for retry. Should be multiple of BatchSize. Default 50,000.
	RetryBufferLimit int
	// The maximum delay between each retry attempt in milliseconds, default 125,000.
	MaxRetryInterval int
	// The maximum total retry timeout in milliseconds, default 180,000.
	MaxRetryTime int
	// The base for the exponential retry delay
	ExponentialBase int
	// Max random value added to the retry delay in milliseconds, default 200
	RetryJitter int
}

var DefaultRetryParams = RetryParams{
	RetryInterval:    5_000,
	MaxRetries:       5,
	RetryBufferLimit: 50_0000,
	MaxRetryInterval: 125_000,
	ExponentialBase:  2,
	MaxRetryTime:     180_000,
}

// WriteParams holds configuration properties for write
type WriteParams struct {
	RetryParams
	// Maximum number of points sent to server in single request, used by PointsWriter. Default 5000
	BatchSize int
	// Maximum size of batch in bytes, used by PointsWriter. Default 50_000_000.
	MaxBatchBytes int
	// Interval, in ms, used by PointsWriter, in which is buffer flushed if it has not been already written (by reaching batch size) . Default 1000ms
	FlushInterval int
	// Precision to use in writes for timestamp.
	// Default lineprotocol.Nanosecond
	Precision lineprotocol.Precision
	// Tags added to each point during writing. If a point already has a tag with the same key, it is left unchanged.
	DefaultTags map[string]string
	// Write body larger than the threshold is gzipped. 0 to don't gzip at all
	GzipThreshold int
	// WriteFailed is called to inform about an error occurred during writing procedure.
	// It can be called when point encoding fails, sending batch over network fails or batch expires during retrying.
	// Params:
	//   error - write error.
	//   lines - failed batch of lines. nil in case of error occur before sending, e.g. in case of conversion error
	//   attempt - count of already failed attempts to write the lines (1 ... maxRetries+1). 0 if error occur before sending, e.g. in case of conversion error
	//   expires - expiration time for the lines to be retried in millis since epoch. 0 if error occur before sending, e.g. in case of conversion error
	// Returns true to continue using default retry mechanism (applies only in case of write of an error)
	WriteFailed func(err error, lines []byte, attempt int, expires time.Time) bool
	// WriteRetrySkipped is informed about lines that were removed from the retry buffer
	// to keep the size of the retry buffer under the configured limit (maxBufferLines).
	WriteRetrySkipped OnRemoveCallback
}

// DefaultWriteParams specifies default write param
var DefaultWriteParams = WriteParams{
	RetryParams:   DefaultRetryParams,
	BatchSize:     5_000,
	MaxBatchBytes: 50_000_000,
	FlushInterval: 60_000,
	Precision:     lineprotocol.Nanosecond,
	GzipThreshold: 1_000,
}
