// Copyright 2020 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

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
	retryInterval uint
	retries       uint
}

func NewBatch(data string, retryInterval uint) *Batch {
	return &Batch{
		batch:         data,
		retryInterval: retryInterval,
	}
}

type Service struct {
	org              string
	bucket           string
	httpService      ihttp.Service
	url              string
	lastWriteAttempt time.Time
	retryQueue       *queue
	lock             sync.Mutex
	writeOptions     *write.Options
}

func NewService(org string, bucket string, httpService ihttp.Service, options *write.Options) *Service {

	retryBufferLimit := options.RetryBufferLimit() / options.BatchSize()
	if retryBufferLimit == 0 {
		retryBufferLimit = 1
	}
	return &Service{org: org, bucket: bucket, httpService: httpService, writeOptions: options, retryQueue: newQueue(int(retryBufferLimit))}
}

func (w *Service) HandleWrite(ctx context.Context, batch *Batch) error {
	log.Log.Debug("Write proc: received write request")
	batchToWrite := batch
	retrying := false
	for {
		select {
		case <-ctx.Done():
			log.Log.Debug("Write proc: ctx cancelled req")
			return ctx.Err()
		default:
		}
		if !w.retryQueue.isEmpty() {
			log.Log.Debug("Write proc: taking batch from retry queue")
			if !retrying {
				b := w.retryQueue.first()
				// Can we write? In case of retryable error we must wait a bit
				if w.lastWriteAttempt.IsZero() || time.Now().After(w.lastWriteAttempt.Add(time.Millisecond*time.Duration(b.retryInterval))) {
					retrying = true
				} else {
					log.Log.Warn("Write proc: cannot write yet, storing batch to queue")
					w.retryQueue.push(batch)
					batchToWrite = nil
				}
			}
			if retrying {
				batchToWrite = w.retryQueue.pop()
				batchToWrite.retries++
				if batch != nil {
					if w.retryQueue.push(batch) {
						log.Log.Warn("Write proc: Retry buffer full, discarding oldest batch")
					}
					batch = nil
				}
			}
		}
		if batchToWrite != nil {
			err := w.WriteBatch(ctx, batchToWrite)
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

func (w *Service) WriteBatch(ctx context.Context, batch *Batch) error {
	wUrl, err := w.WriteUrl()
	if err != nil {
		log.Log.Errorf("%s\n", err.Error())
		return err
	}
	var body io.Reader
	body = strings.NewReader(batch.batch)
	log.Log.Debugf("Writing batch: %s", batch.batch)
	if w.writeOptions.UseGZip() {
		body, err = gzip.CompressWithGzip(body)
		if err != nil {
			return err
		}
	}
	w.lastWriteAttempt = time.Now()
	perror := w.httpService.PostRequest(ctx, wUrl, body, func(req *http.Request) {
		if w.writeOptions.UseGZip() {
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
			log.Log.Errorf("Write error: %s\nChecking retry\n", perror.Error())
			if perror.RetryAfter > 0 {
				batch.retryInterval = perror.RetryAfter * 1000
			} else {
				batch.retryInterval = w.writeOptions.RetryInterval()
			}
			if batch.retries < w.writeOptions.MaxRetries() {
				log.Log.Errorf("Write error: \nBatch kept for retrying\n")
				if w.retryQueue.push(batch) {
					log.Log.Warn("Retry buffer full, discarding oldest batch")
				}
			} else {
				log.Log.Errorf("Write error: \nMax retrys, batch discarded\n")
			}
		} else {
			log.Log.Errorf("Write error: %s\n", perror.Error())
		}
		return perror
	}
	return nil
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

func (w *Service) WriteUrl() (string, error) {
	if w.url == "" {
		u, err := url.Parse(w.httpService.ServerApiUrl())
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
