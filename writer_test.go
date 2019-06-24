package influxdb_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	influxdb "github.com/influxdata/influxdb-client-go"
)

func TestWriterStartupAndShutdown(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	cl, err := influxdb.New(server.Client(), influxdb.WithToken("foo"), influxdb.WithAddress(server.URL))
	if err != nil {
		server.Close()
		t.Fatal(err)
	}
	w := cl.NewBufferingWriter("my-bucket", "my-org", 10*time.Second, 1024*100, func(err error) {
		t.Error(err)
	})
	wg := sync.WaitGroup{}
	w.Start()
	for i := 0; i < 20; i++ {
		wg.Add(1)
		time.Sleep(1)
		go func() {
			w.Start()
			wg.Done()
		}()
	}
	for i := 0; i < 20; i++ {
		wg.Add(1)
		time.Sleep(1)
		go func() {
			w.Stop()
			wg.Done()
		}()
	}
	wg.Wait()
}

func TestAutoFlush(t *testing.T) {
	q := uint64(0)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		res := atomic.AddUint64(&q, 1)
		if res > 3 {
			t.Errorf("size based flush happened too often, expected 3 but got %d", res)
		}
	}))
	cl, err := influxdb.New(server.Client(), influxdb.WithToken("foo"), influxdb.WithAddress(server.URL))
	if err != nil {
		t.Error(e2e)
	}
	w := cl.NewBufferingWriter("my-bucket", "my-org", 0, 100*1024, func(err error) {
		t.Error(err)
	})
	w.Start()
	ts := time.Time{}
	for i := 0; i < 3000; i++ {
		ts = ts.Add(1)
		_, err = w.Write([]byte("TestWriterE2E"),
			ts,
			[][]byte{[]byte("test1"), []byte("test2")},
			[][]byte{[]byte("here"), []byte("alsohere")},
			[][]byte{[]byte("val1"), []byte("val2")},
			[]interface{}{1, 99})
		if err != nil {
			t.Error(err)
		}
	}
	w.Flush(context.Background())
	tries := atomic.LoadUint64(&q)
	w.Stop()
	if tries < 3 {
		t.Errorf("size based flush happened too infrequently expected 3 got %d", tries)
	}
}
