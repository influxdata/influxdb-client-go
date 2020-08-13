// Copyright 2020 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

// Package write provides service and its stuff
package write

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/influxdb-client-go/api/write"
	"github.com/influxdata/influxdb-client-go/internal/gzip"
	ihttp "github.com/influxdata/influxdb-client-go/internal/http"
	"github.com/influxdata/influxdb-client-go/internal/log"
	lp "github.com/influxdata/line-protocol"
)

type Batch struct {
	batch         string
	retryDelay    uint
	retryAttempts uint
	evicted       bool
}

func NewBatch(data string, retryDelay uint) *Batch {
	return &Batch{
		batch:      data,
		retryDelay: retryDelay,
	}
}

type Service struct {
	org                  string
	bucket               string
	httpService          ihttp.Service
	url                  string
	lastWriteAttempt     time.Time
	retryQueue           *queue
	lock                 sync.Mutex
	writeOptions         *write.Options
	retryExponentialBase uint
}

func NewService(org string, bucket string, httpService ihttp.Service, options *write.Options) *Service {

	retryBufferLimit := options.RetryBufferLimit() / options.BatchSize()
	if retryBufferLimit == 0 {
		retryBufferLimit = 1
	}
	return &Service{org: org, bucket: bucket, httpService: httpService, writeOptions: options, retryQueue: newQueue(int(retryBufferLimit)), retryExponentialBase: 5}
}

func (w *Service) HandleWrite(ctx context.Context, batch *Batch) error {
	log.Debug("Write proc: received write request")
	batchToWrite := batch
	retrying := false
	for {
		select {
		case <-ctx.Done():
			log.Debug("Write proc: ctx cancelled req")
			return ctx.Err()
		default:
		}
		if !w.retryQueue.isEmpty() {
			log.Debug("Write proc: taking batch from retry queue")
			if !retrying {
				b := w.retryQueue.first()
				// Can we write? In case of retryable error we must wait a bit
				if w.lastWriteAttempt.IsZero() || time.Now().After(w.lastWriteAttempt.Add(time.Millisecond*time.Duration(b.retryDelay))) {
					retrying = true
				} else {
					log.Warn("Write proc: cannot write yet, storing batch to queue")
					if w.retryQueue.push(batch) {
						log.Warn("Write proc: Retry buffer full, discarding oldest batch")
					}
					batchToWrite = nil
				}
			}
			if retrying {
				batchToWrite = w.retryQueue.first()
				batchToWrite.retryAttempts++
				if batch != nil { //store actual batch to retry queue
					if w.retryQueue.push(batch) {
						log.Warn("Write proc: Retry buffer full, discarding oldest batch")
					}
					batch = nil
				}
			}
		}
		// write batch
		if batchToWrite != nil {
			perror := w.WriteBatch(ctx, batchToWrite)
			if perror != nil {
				if perror.StatusCode >= http.StatusTooManyRequests {
					log.Errorf("Write error: %s\nBatch kept for retrying\n", perror.Error())
					if perror.RetryAfter > 0 {
						batchToWrite.retryDelay = perror.RetryAfter * 1000
					} else {
						exp := uint(1)
						for i := uint(0); i < batchToWrite.retryAttempts; i++ {
							exp = exp * w.retryExponentialBase
						}
						batchToWrite.retryDelay = min(w.writeOptions.RetryInterval()*exp, w.writeOptions.MaxRetryInterval())
					}
					if batchToWrite.retryAttempts == 0 {
						if w.retryQueue.push(batch) {
							log.Warn("Retry buffer full, discarding oldest batch")
						}
					} else if batchToWrite.retryAttempts == w.writeOptions.MaxRetries() {
						log.Warn("Reached maximum number of retries, discarding batch")
						if !batchToWrite.evicted {
							w.retryQueue.pop()
						}
					}
				} else {
					log.Errorf("Write error: %s\n", perror.Error())
				}
				return perror
			} else {
				if retrying && !batchToWrite.evicted {
					w.retryQueue.pop()
				}
				batchToWrite = nil
			}
		} else {
			break
		}
	}
	return nil
}

func (w *Service) WriteBatch(ctx context.Context, batch *Batch) *ihttp.Error {
	wURL, err := w.WriteURL()
	if err != nil {
		log.Errorf("%s\n", err.Error())
		return ihttp.NewError(err)
	}
	var body io.Reader
	body = strings.NewReader(batch.batch)
	log.Debugf("Writing batch: %s", batch.batch)
	if w.writeOptions.UseGZip() {
		body, err = gzip.CompressWithGzip(body)
		if err != nil {
			return ihttp.NewError(err)
		}
	}
	w.lastWriteAttempt = time.Now()
	perror := w.httpService.PostRequest(ctx, wURL, body, func(req *http.Request) {
		if w.writeOptions.UseGZip() {
			req.Header.Set("Content-Encoding", "gzip")
		}
	}, func(r *http.Response) error {
		// discard body so connection can be reused
		// _, _ = io.Copy(ioutil.Discard, r.Body)
		// _ = r.Body.Close()
		return nil
	})
	return perror
}

type pointWithDefaultTags struct {
	point       *write.Point
	defaultTags map[string]string
}

// Name returns the name of measurement of a point.
func (p *pointWithDefaultTags) Name() string {
	return p.point.Name()
}

// Time is the timestamp of a Point.
func (p *pointWithDefaultTags) Time() time.Time {
	return p.point.Time()
}

// FieldList returns a slice containing the fields of a Point.
func (p *pointWithDefaultTags) FieldList() []*lp.Field {
	return p.point.FieldList()
}

func (p *pointWithDefaultTags) TagList() []*lp.Tag {
	tags := make([]*lp.Tag, 0, len(p.point.TagList())+len(p.defaultTags))
	tags = append(tags, p.point.TagList()...)
	for k, v := range p.defaultTags {
		if !existTag(p.point.TagList(), k) {
			tags = append(tags, &lp.Tag{
				Key:   k,
				Value: v,
			})
		}
	}
	sort.Slice(tags, func(i, j int) bool { return tags[i].Key < tags[j].Key })
	return tags
}

func existTag(tags []*lp.Tag, key string) bool {
	for _, tag := range tags {
		if key == tag.Key {
			return true
		}
	}
	return false
}

func (w *Service) EncodePoints(points ...*write.Point) (string, error) {
	var buffer bytes.Buffer
	e := lp.NewEncoder(&buffer)
	e.SetFieldTypeSupport(lp.UintSupport)
	e.FailOnFieldErr(true)
	e.SetPrecision(w.writeOptions.Precision())
	for _, point := range points {
		_, err := e.Encode(w.pointToEncode(point))
		if err != nil {
			return "", err
		}
	}
	return buffer.String(), nil
}

func (w *Service) pointToEncode(point *write.Point) lp.Metric {
	var m lp.Metric
	if len(w.writeOptions.DefaultTags()) > 0 {
		m = &pointWithDefaultTags{
			point:       point,
			defaultTags: w.writeOptions.DefaultTags(),
		}
	} else {
		m = point
	}
	return m
}

func (w *Service) WriteURL() (string, error) {
	if w.url == "" {
		u, err := url.Parse(w.httpService.ServerAPIURL())
		if err != nil {
			return "", err
		}
		u, err = u.Parse("write")
		if err != nil {
			return "", err
		}
		params := u.Query()
		params.Set("org", w.org)
		params.Set("bucket", w.bucket)
		params.Set("precision", precisionToString(w.writeOptions.Precision()))
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

func min(a, b uint) uint {
	if a > b {
		return b
	}
	return a
}
