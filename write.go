// Copyright 2020 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package influxdb2

import (
	"context"
	"github.com/influxdata/influxdb-client-go/internal/http"
	"strings"
	"time"
)

// WriteApiBlocking is Write client interface with non-blocking methods for writing time series data asynchronously in batches into an InfluxDB server.
type WriteApi interface {
	// WriteRecord writes asynchronously line protocol record into bucket.
	// WriteRecord adds record into the buffer which is sent on the background when it reaches the batch size.
	// Blocking alternative is available in the WriteApiBlocking interface
	WriteRecord(line string)
	// WritePoint writes asynchronously Point into bucket.
	// WritePoint adds Point into the buffer which is sent on the background when it reaches the batch size.
	// Blocking alternative is available in the WriteApiBlocking interface
	WritePoint(point *Point)
	// Flush forces all pending writes from the buffer to be sent
	Flush()
	// Flushes all pending writes and stop async processes. After this the Write client cannot be used
	Close()
	// Errors returns a channel for reading errors which occurs during async writes.
	// Must be called before performing any writes for errors to be collected.
	// The chan is unbuffered and must be drained or the writer will block.
	Errors() <-chan error
}

type writeApiImpl struct {
	service     *writeService
	writeBuffer []string

	writeCh      chan *batch
	bufferCh     chan string
	writeStop    chan int
	bufferStop   chan int
	bufferFlush  chan int
	doneCh       chan int
	errCh        chan error
	bufferInfoCh chan writeBuffInfoReq
	writeInfoCh  chan writeBuffInfoReq
}

type writeBuffInfoReq struct {
	writeBuffLen int
}

func newWriteApiImpl(org string, bucket string, service http.Service, client InfluxDBClient) *writeApiImpl {
	w := &writeApiImpl{
		service:      newWriteService(org, bucket, service, client),
		writeBuffer:  make([]string, 0, client.Options().BatchSize()+1),
		writeCh:      make(chan *batch),
		doneCh:       make(chan int),
		bufferCh:     make(chan string),
		bufferStop:   make(chan int),
		writeStop:    make(chan int),
		bufferFlush:  make(chan int),
		bufferInfoCh: make(chan writeBuffInfoReq),
		writeInfoCh:  make(chan writeBuffInfoReq),
	}
	go w.bufferProc()
	go w.writeProc()

	return w
}

func (w *writeApiImpl) Errors() <-chan error {
	if w.errCh == nil {
		w.errCh = make(chan error)
	}
	return w.errCh
}

func (w *writeApiImpl) Flush() {
	w.bufferFlush <- 1
	w.waitForFlushing()
}

func (w *writeApiImpl) waitForFlushing() {
	for {
		w.bufferInfoCh <- writeBuffInfoReq{}
		writeBuffInfo := <-w.bufferInfoCh
		if writeBuffInfo.writeBuffLen == 0 {
			break
		}
		logger.Info("Waiting buffer is flushed")
		time.Sleep(time.Millisecond)
	}
	for {
		w.writeInfoCh <- writeBuffInfoReq{}
		writeBuffInfo := <-w.writeInfoCh
		if writeBuffInfo.writeBuffLen == 0 {
			break
		}
		logger.Info("Waiting buffer is flushed")
		time.Sleep(time.Millisecond)
	}
	//time.Sleep(time.Millisecond)
}

func (w *writeApiImpl) bufferProc() {
	logger.Info("Buffer proc started")
	ticker := time.NewTicker(time.Duration(w.service.client.Options().FlushInterval()) * time.Millisecond)
x:
	for {
		select {
		case line := <-w.bufferCh:
			w.writeBuffer = append(w.writeBuffer, line)
			if len(w.writeBuffer) == int(w.service.client.Options().BatchSize()) {
				w.flushBuffer()
			}
		case <-ticker.C:
			w.flushBuffer()
		case <-w.bufferFlush:
			w.flushBuffer()
		case <-w.bufferStop:
			ticker.Stop()
			w.flushBuffer()
			break x
		case buffInfo := <-w.bufferInfoCh:
			buffInfo.writeBuffLen = len(w.bufferInfoCh)
			w.bufferInfoCh <- buffInfo
		}
	}
	logger.Info("Buffer proc finished")
	w.doneCh <- 1
}

func (w *writeApiImpl) flushBuffer() {
	if len(w.writeBuffer) > 0 {
		//go func(lines []string) {
		logger.Info("sending batch")
		batch := &batch{batch: buffer(w.writeBuffer)}
		w.writeCh <- batch
		//	lines = lines[:0]
		//}(w.writeBuffer)
		//w.writeBuffer = make([]string,0, w.service.client.Options.BatchSize+1)
		w.writeBuffer = w.writeBuffer[:0]
	}
}

func (w *writeApiImpl) writeProc() {
	logger.Info("Write proc started")
x:
	for {
		select {
		case batch := <-w.writeCh:
			err := w.service.handleWrite(context.Background(), batch)
			if err != nil && w.errCh != nil {
				w.errCh <- err
			}
		case <-w.writeStop:
			logger.Info("Write proc: received stop")
			break x
		case buffInfo := <-w.writeInfoCh:
			buffInfo.writeBuffLen = len(w.writeCh)
			w.writeInfoCh <- buffInfo
		}
	}
	logger.Info("Write proc finished")
	w.doneCh <- 1
}

func (w *writeApiImpl) Close() {
	if w.writeCh != nil {
		// Flush outstanding metrics
		w.Flush()
		w.bufferStop <- 1
		//wait for buffer proc
		<-w.doneCh
		close(w.bufferStop)
		close(w.bufferFlush)
		close(w.bufferCh)
		w.writeStop <- 1
		//wait for the write proc
		<-w.doneCh
		close(w.writeCh)
		close(w.writeStop)
		close(w.writeInfoCh)
		close(w.bufferInfoCh)
		w.bufferInfoCh = nil
		w.writeInfoCh = nil
		w.writeCh = nil
		w.writeStop = nil
		w.bufferFlush = nil
		w.bufferStop = nil
		if w.errCh != nil {
			close(w.errCh)
			w.errCh = nil
		}
	}
}

func (w *writeApiImpl) WriteRecord(line string) {
	b := []byte(line)
	b = append(b, 0xa)
	w.bufferCh <- string(b)
}

func (w *writeApiImpl) WritePoint(point *Point) {
	//w.bufferCh <- point.ToLineProtocol(w.service.client.Options().Precision)
	line, err := w.service.encodePoints(point)
	if err != nil {
		logger.Errorf("point encoding error: %s\n", err.Error())
	} else {
		w.bufferCh <- line
	}
}

func buffer(lines []string) string {
	return strings.Join(lines, "")
}
