package influxclient

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"
)

// BytesWrite is a function writing data to a bucket
type BytesWrite func(ctx context.Context, bucket string, bs []byte) error

// WriteBuffer stores lines after line and flushes batch when maxLength (bach size) is  reached
// or maxBytes exceeds
type WriteBuffer struct {
	length    int
	bytes     int
	lines     []byte
	maxLength int
	maxBytes  int
	flushFn   func(bytes []byte)
}

func (w *WriteBuffer) Add(line []byte) {
	if w.bytes+len(line) > w.maxBytes {
		w.Flush()
	}
	w.lines = append(w.lines, line...)
	w.length++
	w.bytes += len(line)
	if w.length == w.maxLength {
		w.Flush()
	}
}

func (w *WriteBuffer) Flush() {
	buff := w.Reset()
	if len(buff) > 0 {
		w.flushFn(buff)
	}
}

func (w *WriteBuffer) Reset() []byte {
	var ret []byte
	if w.bytes > 0 {
		ret = make([]byte, w.bytes)
		copy(ret, w.lines[0:w.bytes])
		w.bytes = 0
		w.length = 0
		w.lines = w.lines[:0]
	}
	return ret
}

// PointsWriter is asynchronous writer with automated batching and retrying capabilities.
// It is parametrized by the WriteParams.
// It is obtained using the Client.PointsWriter()
// Use Write, WriteData or WritePoints for sending data
// Any error encountered during asynchronous processing is reported by WriteParams.WriteFailed callback.
// It must be created using NewPointsWriter
// All functions are thread-safe and can be used from different go-routines.
type PointsWriter struct {
	writer        BytesWrite
	batchCh       chan *batch
	bufferCh      chan []byte
	stopCh        chan struct{}
	flushCh       chan struct{}
	flushT        *time.Timer
	params        WriteParams
	bucket        string
	writeBuffer   *WriteBuffer
	retryBuffer   *RetryBuffer
	retryStrategy RetryStrategy
}

type batch struct {
	lines             []byte
	remainingAttempts int
	expires           time.Time
}

// NewPointsWriter creates fast asynchronous PointsWriter writing to a bucket using given writer according the params
func NewPointsWriter(writer BytesWrite, bucket string, params WriteParams) *PointsWriter {

	write := &PointsWriter{
		writer:   writer,
		bufferCh: make(chan []byte, 1),
		batchCh:  make(chan *batch, 1),
		flushCh:  make(chan struct{}),
		stopCh:   make(chan struct{}),
		params:   params,
		bucket:   bucket,
	}
	write.writeBuffer = &WriteBuffer{
		maxLength: params.BatchSize,
		maxBytes:  params.MaxBatchBytes,
		flushFn: func(bytes []byte) {
			write.sendBatch(bytes, params.MaxRetries, time.Now().Add(time.Duration(params.MaxRetryTime)*time.Millisecond))
		},
	}
	write.retryStrategy = GetRetryStrategyFactory()(params.RetryParams)
	write.retryBuffer = NewRetryBuffer(params.RetryBufferLimit, write.sendBatch, params.WriteRetrySkipped)

	go write.writeProc()
	go write.bufferProc()
	write.scheduleFlush()
	return write
}

// Write asynchronously writes line protocol record(s) to the server.
// Multiple records must be separated by the new line character (\n).
func (p *PointsWriter) Write(line []byte) {
	if len(line) > 0 {
		p.bufferCh <- line
	}
}

// WritePoints asynchronously writes all the given points to the server.
// Any error encountered during encoding points is reported by WriteParams.WriteFailed callback.
func (p *PointsWriter) WritePoints(points ...*Point) {
	for _, pt := range points {
		bts, err := pt.MarshalBinary(p.params.Precision)
		if err != nil {
			mess := fmt.Sprintf("Point encoding failed: %v", err)
			log.Printf("[W] PointsWriter: %s", mess)
			if p.params.WriteFailed != nil {
				p.params.WriteFailed(errors.New(mess), nil, 0, time.Time{})
			}
			continue
		}
		p.Write(bts)
	}
}

// WriteData asynchronously encodes fields of custom points into line protocol
// and writes line protocol record(s) to the server into the given bucket.
// Any error encountered during asynchronous processing is reported by WriteParams.WriteFailed callback.
// Each custom point must be annotated with 'lp' prefix and values measurement,tag, field or timestamp.
// Valid point must contain measurement and at least one field.
//
// A field with timestamp must be of a type time.Time
//
//	 type TemperatureSensor struct {
//		  Measurement string `lp:"measurement"`
//		  Sensor string `lp:"tag,sensor"`
//		  ID string `lp:"tag,device_id"`
//		  Temp float64 `lp:"field,temperature"`
//		  Hum int	`lp:"field,humidity"`
//		  Time time.Time `lp:"timestamp,temperature"`
//		  Description string `lp:"-"`
//	 }
func (p *PointsWriter) WriteData(points ...interface{}) {
	for _, d := range points {
		byts, err := encode(d, p.params.Precision)
		if err != nil {
			mess := fmt.Sprintf("Point encoding failed: %v", err)
			log.Printf("[W] PointsWriter: %s", mess)
			if p.params.WriteFailed != nil {
				p.params.WriteFailed(errors.New(mess), nil, 0, time.Time{})
			}
			continue
		}
		p.Write(byts)
	}
}

func (p *PointsWriter) scheduleFlush() {
	if p.flushT != nil {
		p.flushT.Stop()
	}
	p.flushT = time.AfterFunc(time.Duration(p.params.FlushInterval)*time.Millisecond, func() {
		//log.Println("[D] PointsWriter: timed flush")
		p.flushCh <- struct{}{}
	})
}

func (p *PointsWriter) sendBatch(lines []byte, remainingAttempts int, expires time.Time) {
	p.batchCh <- &batch{
		lines,
		remainingAttempts,
		expires,
	}
}

func (p *PointsWriter) bufferProc() {
	run := true
	for run {
		select {
		case b := <-p.bufferCh:
			p.writeBuffer.Add(b)
		case <-p.flushCh:
			p.writeBuffer.Flush()
			p.scheduleFlush()
		case <-p.stopCh:
			if p.flushT != nil {
				p.flushT.Stop()
			}
			run = false
		}
	}
	p.stopCh <- struct{}{}
}

func (p *PointsWriter) writeProc() {
	for batch := range p.batchCh {
		failedAttempts := p.params.MaxRetries - batch.remainingAttempts + 1
		if batch.expires.Before(time.Now()) {
			err := errors.New("max retry time exceeded")
			log.Printf("[W] PointsWriter: %s", err.Error())
			if p.params.WriteFailed != nil {
				p.params.WriteFailed(err, batch.lines, failedAttempts, batch.expires)
			}
			continue
		}

		if err := p.writer(context.Background(), p.bucket, batch.lines); err != nil {
			retry := true
			if se, ok := err.(*ServerError); ok {
				retry = se.StatusCode >= 429
			}
			if p.params.WriteFailed != nil {
				retry = p.params.WriteFailed(err, batch.lines, failedAttempts, batch.expires)
			}
			if retry {
				p.retryBuffer.AddLines(
					batch.lines,
					batch.remainingAttempts-1,
					p.retryStrategy.NextDelay(err, failedAttempts-1),
					batch.expires)
				log.Printf("[W] PointsWriter: write to InfluxDB failed (attempt: %d): %v", failedAttempts, err)
			} else {
				log.Printf("[E] PointsWriter:  write to InfluxDB failed (attempt: %d): %v", failedAttempts, err)
			}

		} else {
			p.retryStrategy.Success()
		}
	}
	p.stopCh <- struct{}{}
}

// Flush asynchronously flushes write buffer.
// This enforces sending data on demand, even when flush conditions (batch size, flush interval, max batch bytes)
// are not met.
func (p *PointsWriter) Flush() {
	p.retryBuffer.Flush()
	p.flushCh <- struct{}{}
	for len(p.bufferCh) > 0 {
		<-time.After(time.Millisecond)
	}
	for len(p.batchCh) > 0 {
		<-time.After(time.Millisecond)
	}

}

// Close stops internal routines and closes resources
// Must be called by user at the end
func (p *PointsWriter) Close() {
	p.retryBuffer.Close()
	p.stopCh <- struct{}{}
	close(p.bufferCh)
	close(p.batchCh)
	<-p.stopCh
	<-p.stopCh
	close(p.stopCh)
}
