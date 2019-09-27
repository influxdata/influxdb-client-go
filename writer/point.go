package writer

import (
	"io"
	"sync"
	"time"

	"github.com/influxdata/influxdb-client-go"
)

const defaultFlushInterval = 1 * time.Second

// MetricsWriteFlush is a type of metrics writer which is
// buffered and metrics can be flushed to
type MetricsWriteFlusher interface {
	Write(m ...influxdb.Metric) (int, error)
	Available() int
	Flush() error
}

// PointWriter delegates calls to Write to an underlying flushing writer
// implementation. It also periodically calls flush on the underlying writer and is safe
// to be called concurrently. As the flushing writer can also flush on calls to Write
// when the number of metrics being written exceeds the buffer capacity, it also ensures
// to reset its timer in this scenario as to avoid calling flush multiple times
type PointWriter struct {
	w             MetricsWriteFlusher
	flushInterval time.Duration
	resetTick     chan struct{}
	stopped       chan struct{}
	err           error
	mu            sync.Mutex
}

// NewPointWriter configures and returns a *PointWriter writer type
// The new writer will automatically begin scheduling periodic flushes based on the
// provided duration
func NewPointWriter(w MetricsWriteFlusher, flushInterval time.Duration) *PointWriter {
	writer := &PointWriter{
		w:             w,
		flushInterval: flushInterval,
		// buffer of one in order to not block writes
		resetTick: make(chan struct{}, 1),
		// stopped is closed once schedule has exited
		stopped: make(chan struct{}),
	}

	go writer.schedule()

	return writer
}

func (p *PointWriter) schedule() {
	defer close(p.stopped)

	ticker := time.NewTicker(p.flushInterval)

	for {
		select {
		case <-ticker.C:
			if err := func() error {
				p.mu.Lock()
				defer p.mu.Unlock()

				// return if error is now not nil
				if p.err != nil {
					return p.err
				}

				// between the recv on the ticker and the lock obtain
				// the reset tick could've been triggered so we check
				// and skip the flush if it did
				select {
				case <-p.resetTick:
					return nil
				default:
				}

				p.err = p.w.Flush()

				return p.err
			}(); err != nil {
				return
			}
		case _, ok := <-p.resetTick:
			if !ok {
				return
			}

			ticker.Stop()
			ticker = time.NewTicker(p.flushInterval)
		}
	}
}

// Write delegates to an underlying metrics writer
// If the delegating call is going to cause a flush, it signals
// to the schduled periodic flush to reset its timer
func (p *PointWriter) Write(m ...influxdb.Metric) (int, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.err != nil {
		return 0, p.err
	}

	// check if the underlying flush will flush
	if len(m) > p.w.Available() {
		// tell the ticker to reset flush interval
		select {
		case p.resetTick <- struct{}{}:
		default:
		}
	}

	var n int
	n, p.err = p.w.Write(m...)
	return n, p.err
}

// Close signals to stop flushing metrics and causes subsequent
// calls to Write to return a closed pipe error
// Close returns once scheduledge flushing has stopped
// Close does a final flush on return and returns any
// error from the final flush if it occurs
func (p *PointWriter) Close() error {
	p.mu.Lock()

	// signal close
	close(p.resetTick)

	// return err io closed pipe for subsequent writes
	p.err = io.ErrClosedPipe

	// release lock so scheduled may acknowledge and exit
	p.mu.Unlock()

	// wait until schedule exits
	<-p.stopped

	return p.w.Flush()
}
