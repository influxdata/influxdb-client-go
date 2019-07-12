package influxdb

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	lp "github.com/influxdata/line-protocol"
)

const maxPooledBuffer = 4 << 20 //8 megs

// LPWriter is a type for writing line protocol in a buffered way.
// It allows you to set a flush interval and flush regularly or to call the Flush method to flush its internal buffer.
type LPWriter struct {
	stopTicker    func()
	flushChan     <-chan time.Time
	flushInterval time.Duration
	flushSize     int
	c             *Client
	buf           switchableBuffer
	lock          sync.Mutex
	enc           *lp.Encoder
	bucket, org   string
	tries         uint64
	maxRetries    int
	errOnFieldErr bool
	stop          chan struct{}
	once          sync.Once
	wg            sync.WaitGroup
	onError       func(error)
}

type switchableBuffer struct {
	*bytes.Buffer
}

// WriteMetrics writes Metrics to the LPWriter.
func (w *LPWriter) WriteMetrics(m ...Metric) (int, error) {
	select {
	case <-w.stop:
		return 0, nil
	default:
	}
	w.lock.Lock()
	for i := range m {
		j, err := w.enc.Encode(m[i])
		if err != nil {
			return j, err
		}
	}
	w.asyncFlush()
	w.lock.Unlock()
	return 0, nil
}

// NewBufferingWriter creates a new BufferingWriter.
func (c *Client) NewBufferingWriter(bucket string, org string, flushInterval time.Duration, flushSize int, onError func(error)) *LPWriter {
	// if onError is nil set to a noop
	if onError == nil {
		onError = func(_ error) {}
	}
	w := &LPWriter{c: c, buf: switchableBuffer{&bytes.Buffer{}}, flushSize: flushSize, flushInterval: flushInterval, stop: make(chan struct{}), onError: onError}
	w.enc = lp.NewEncoder(&w.buf)
	w.enc.FailOnFieldErr(w.errOnFieldErr)
	return w
}

// Write writes name, time stamp, tag keys, tag values, field keys, and field values to an LPWriter.
func (w *LPWriter) Write(name []byte, ts time.Time, tagKeys, tagVals, fieldKeys [][]byte, fieldVals []interface{}) (int, error) {
	select {
	case <-w.stop:
		return 0, nil
	default:
	}
	w.lock.Lock()
	i, err := w.enc.Write(name, ts, tagKeys, tagVals, fieldKeys, fieldVals)
	// asyncronously flush if the size of the buffer is too big.
	if err != nil {
		return i, err
	}
	w.asyncFlush()
	w.lock.Unlock()
	return i, err
}
func (w *LPWriter) asyncFlush() {
	if w.flushSize > 0 && w.buf.Len() > w.flushSize {
		w.wg.Add(1)
		buf := w.buf.Buffer
		w.buf.Buffer = bufferPool.Get().(*bytes.Buffer)
		go func() {
			w.flush(context.TODO(), buf)
			if buf.Len() <= maxPooledBuffer {
				buf.Reset()
				bufferPool.Put(buf)
			}
			w.wg.Done()
		}()
	}
}

// Start starts an LPWriter, so that the writer can flush it out to influxdb.
func (w *LPWriter) Start() {
	w.lock.Lock()
	w.once = sync.Once{}
	if w.flushInterval != 0 {
		t := time.NewTicker(w.flushInterval)
		w.stopTicker = t.Stop
		w.flushChan = t.C
		w.wg.Add(1)
		go func() {
			for {
				select {
				case <-w.flushChan:
					err := w.Flush(context.Background())
					if err != nil {
						w.onError(err)
					}
				case <-w.stop:
					w.wg.Done()
					return
				}
			}
		}()
	} else {
		w.stopTicker = func() {}
	}
	w.lock.Unlock()
}

var bufferPool = sync.Pool{New: func() interface{} { return &bytes.Buffer{} }}

// Flush writes out the internal buffer to the database.
func (w *LPWriter) Flush(ctx context.Context) error {
	w.wg.Add(1)
	defer w.wg.Done()
	w.lock.Lock()
	if w.buf.Len() == 0 {
		w.lock.Unlock()
		return nil
	}
	buf := w.buf.Buffer
	w.buf.Buffer = bufferPool.Get().(*bytes.Buffer)
	w.lock.Unlock()
	err := w.flush(ctx, buf)
	if err != nil {
		return err
	}
	if buf.Len() <= maxPooledBuffer {
		buf.Reset()
		bufferPool.Put(buf)
	}
	return err
}

func (w *LPWriter) flush(ctx context.Context, buf *bytes.Buffer) error {

	cleanup := func() {}
	defer func() { cleanup() }()
	// early exit so we don't send empty buffers
doRequest:
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	req, err := w.c.makeWriteRequest(w.bucket, w.org, buf)
	if err != nil {
		return err
	}
	resp, err := w.c.httpClient.Do(req)
	if err != nil {
		return err
	}
	cleanup = func() {
		r := io.LimitReader(resp.Body, 1<<24) // we limit it because it is usually better to just reuse the body, but sometimes it isn't worth it.
		// throw away the rest of the body so the connection can be reused even if there is still stuff on the wire.
		_, _ = ioutil.ReadAll(r) // we don't care about the error here, it is just to empty the tcp buffer
		resp.Body.Close()
	}

	switch resp.StatusCode {
	case http.StatusOK, http.StatusNoContent:
	case http.StatusTooManyRequests:
		err = &genericRespError{
			Code:    resp.Status,
			Message: "too many requests too fast",
		}
		cleanup()
		if err2 := w.backoff(&w.tries, resp, err); err2 != nil {
			return err2
		}
		cleanup = func() {}
		goto doRequest
	case http.StatusServiceUnavailable:
		err = &genericRespError{
			Code:    resp.Status,
			Message: "service temporarily unavaliable",
		}
		cleanup()
		if err2 := w.backoff(&w.tries, resp, err); err2 != nil {
			return err2
		}
		cleanup = func() {
			w.lock.Unlock()
		}
		goto doRequest
	default:
		gwerr, err := parseWriteError(resp.Body)
		if err != nil {
			return err
		}

		return gwerr
	}
	// we don't defer and close till here, because of the retries.
	defer func() {
		r := io.LimitReader(resp.Body, 1<<16) // we limit it because it is usually better to just reuse the body, but sometimes it isn't worth it.
		_, err := ioutil.ReadAll(r)           // throw away the rest of the body so the connection gets reused.
		err2 := resp.Body.Close()
		if err == nil && err2 != nil {
			err = err2
		}
	}()
	return err
}

// backoff is a helper method for backoff, triesPtr must not be nil.
func (w *LPWriter) backoff(triesPtr *uint64, resp *http.Response, err error) error {
	tries := atomic.LoadUint64(triesPtr)
	if w.maxRetries >= 0 || int(tries) >= w.maxRetries {
		return maxRetriesExceededError{
			err:   err,
			tries: w.maxRetries,
		}
	}
	retry := 0
	if resp != nil {
		retryAfter := resp.Header.Get("Retry-After")
		retry, _ = strconv.Atoi(retryAfter) // we ignore the error here because an error already means retry is 0.
	}
	sleepFor := time.Duration(retry) * time.Second
	if retry == 0 { // if we didn't get a Retry-After or it is zero, instead switch to exponential backoff
		sleepFor = time.Duration(rand.Int63n(((1 << tries) - 1) * 10 * int64(time.Microsecond)))
	}
	if sleepFor > defaultMaxWait {
		sleepFor = defaultMaxWait
	}
	time.Sleep(sleepFor)
	atomic.AddUint64(triesPtr, 1)
	return nil
}

// Stop gracefully stops a started LPWriter.
func (w *LPWriter) Stop() {
	w.lock.Lock()
	w.once.Do(func() {
		close(w.stop)
		w.wg.Wait()
		w.stopTicker()
		w.stop = make(chan struct{})
	})
	w.lock.Unlock()
	w.wg.Wait()
}
