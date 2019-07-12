// Package writer contains useful types for buffering, batching and periodically syncing
// writes onto a provided metric writing client.
//
// The following example demonstrate the usage of a *writer.PointWriter. This is designed to
// buffer calls to Write metrics and flush them in configurable batch sizes (see WithBufferSize).
// It is also designed to periodically flush the buffer if a configurable duration ellapses between
// calls to Write. This is useful to ensure metrics are flushed to the client during a pause in their
// production.
//
// Example Usage
//
//  import (
//      "github.com/influxdata/influxdb-client-go"
//      "github.com/influxdata/influxdb-client-go/writer"
//  )
//
//  func main() {
//      cli, _ := influxdb.New("http://localhost:9999", "some-token")
//
//      wr := writer.New(cli, influxdb.Organisation("influx"), influxdb.Bucket("default"), writer.WithBufferSize(10))
//
//      wr.Write(influxdb.NewRowMetric(
//          map[string]interface{}{
//			    "value": 16,
//		    },
//		    "temp_celsius",
//		    map[string]string{
//			    "room": "living_room",
//		    },
//		    time.Now(),
//      ),
//      influxdb.NewRowMetric(
//          map[string]interface{}{
//			    "value": 17,
//		    },
//		    "temp_celsius",
//		    map[string]string{
//			    "room": "living_room",
//		    },
//		    time.Now(),
//      ))
//
//      wr.Close()
//  }
//
// writer.New(...) return a PointerWriter which is composed of multiple other types available in this
// package.
//
// It first wraps the provided client in a *BucketWriter which takes care of ensuring all written metrics
// are called on the underyling client with a specific organisation and bucket. This is not safe for
// concurrent use.
//
// It then wraps this writer in a *BufferedWriter and configures its buffer size accordingly. This type
// implements the buffering of metrics and exposes a flush method. Once the buffer size is exceed flush
// is called automatically. However, Flush() can be called manually on this type. This is also not safe
// for concurrent use.
//
// Finally, it wraps the buffered writer in a *PointsWriter which takes care of ensuring Flush is called
// automatically when it hasn't been called for a configured duration. This final type is safe for concurrent use.
package writer
