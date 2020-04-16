// Copyright 2020 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package influxdb2

import (
	"bytes"
	"context"
	ihttp "github.com/influxdata/influxdb-client-go/internal/http"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/influxdb-client-go/internal/gzip"
	"github.com/influxdata/influxdb-client-go/internal/log"
	lp "github.com/influxdata/line-protocol"
)

var logger log.Logger

type batch struct {
	batch         string
	retryInterval uint
	retries       uint
}

type writeService struct {
	org              string
	bucket           string
	httpService      ihttp.Service
	url              string
	lastWriteAttempt time.Time
	retryQueue       *queue
	lock             sync.Mutex
	client           InfluxDBClient
}

func newWriteService(org string, bucket string, httpService ihttp.Service, client InfluxDBClient) *writeService {
	logger.SetDebugLevel(client.Options().LogLevel())
	retryBufferLimit := client.Options().RetryBufferLimit() / client.Options().BatchSize()
	if retryBufferLimit == 0 {
		retryBufferLimit = 1
	}
	return &writeService{org: org, bucket: bucket, httpService: httpService, client: client, retryQueue: newQueue(int(retryBufferLimit))}
}

func (w *writeService) handleWrite(ctx context.Context, batch *batch) error {
	logger.Debug("Write proc: received write request")
	batchToWrite := batch
	retrying := false
	for {
		select {
		case <-ctx.Done():
			logger.Debug("Write proc: ctx cancelled req")
			return ctx.Err()
		default:
		}
		if !w.retryQueue.isEmpty() {
			logger.Debug("Write proc: taking batch from retry queue")
			if !retrying {
				b := w.retryQueue.first()
				// Can we write? In case of retryable error we must wait a bit
				if w.lastWriteAttempt.IsZero() || time.Now().After(w.lastWriteAttempt.Add(time.Millisecond*time.Duration(b.retryInterval))) {
					retrying = true
				} else {
					logger.Warn("Write proc: cannot write yet, storing batch to queue")
					w.retryQueue.push(batch)
					batchToWrite = nil
				}
			}
			if retrying {
				batchToWrite = w.retryQueue.pop()
				batchToWrite.retries++
				if batch != nil {
					if w.retryQueue.push(batch) {
						logger.Warn("Write proc: Retry buffer full, discarding oldest batch")
					}
					batch = nil
				}
			}
		}
		if batchToWrite != nil {
			err := w.writeBatch(ctx, batchToWrite)
			batchToWrite = nil
			if err != nil {
				return err
			}
		} else {
			break
		}
	}
	return nil
}

func (w *writeService) writeBatch(ctx context.Context, batch *batch) error {
	wUrl, err := w.writeUrl()
	if err != nil {
		logger.Errorf("%s\n", err.Error())
		return err
	}
	var body io.Reader
	body = strings.NewReader(batch.batch)
	logger.Debugf("Writing batch: %s", batch.batch)
	if w.client.Options().UseGZip() {
		body, err = gzip.CompressWithGzip(body)
		if err != nil {
			return err
		}
	}
	w.lastWriteAttempt = time.Now()
	perror := w.httpService.PostRequest(ctx, wUrl, body, func(req *http.Request) {
		if w.client.Options().UseGZip() {
			req.Header.Set("Content-Encoding", "gzip")
		}
	}, func(r *http.Response) error {
		// discard body so connection can be reused
		//_, _ = io.Copy(ioutil.Discard, r.Body)
		//_ = r.Body.Close()
		return nil
	})

	if perror != nil {
		if perror.StatusCode == http.StatusTooManyRequests || perror.StatusCode == http.StatusServiceUnavailable {
			logger.Errorf("Write error: %s\nBatch kept for retrying\n", perror.Error())
			if perror.RetryAfter > 0 {
				batch.retryInterval = perror.RetryAfter * 1000
			} else {
				batch.retryInterval = w.client.Options().RetryInterval()
			}
			if batch.retries < w.client.Options().MaxRetries() {
				if w.retryQueue.push(batch) {
					logger.Warn("Retry buffer full, discarding oldest batch")
				}
			}
		} else {
			logger.Errorf("Write error: %s\n", perror.Error())
		}
		return perror
	}
	return nil
}

func (w *writeService) encodePoints(points ...*Point) (string, error) {
	var buffer bytes.Buffer
	e := lp.NewEncoder(&buffer)
	e.SetFieldTypeSupport(lp.UintSupport)
	e.FailOnFieldErr(true)
	e.SetPrecision(w.client.Options().Precision())
	for _, point := range points {
		_, err := e.Encode(point)
		if err != nil {
			return "", err
		}
	}
	return buffer.String(), nil
}

func (w *writeService) writeUrl() (string, error) {
	if w.url == "" {
		u, err := url.Parse(w.httpService.ServerApiUrl())
		if err != nil {
			return "", err
		}
		u.Path = path.Join(u.Path, "write")

		params := u.Query()
		params.Set("org", w.org)
		params.Set("bucket", w.bucket)
		params.Set("precision", precisionToString(w.client.Options().Precision()))
		u.RawQuery = params.Encode()
		w.lock.Lock()
		w.url = u.String()
		w.lock.Unlock()
	}
	return w.url, nil
}

func precisionToString(precision time.Duration) string {
	prec := "ns"
	switch precision {
	case time.Microsecond:
		prec = "us"
	case time.Millisecond:
		prec = "ms"
	case time.Second:
		prec = "s"
	}
	return prec
}
