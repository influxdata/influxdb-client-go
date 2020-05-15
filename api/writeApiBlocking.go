// Copyright 2020 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package api

import (
	"context"
	"github.com/influxdata/influxdb-client-go/api/write"
	"github.com/influxdata/influxdb-client-go/internal/http"
	iwrite "github.com/influxdata/influxdb-client-go/internal/write"
	"strings"
)

// WriteApiBlocking offers blocking methods for writing time series data synchronously into an InfluxDB server.
type WriteApiBlocking interface {
	// WriteRecord writes line protocol record(s) into bucket.
	// WriteRecord writes without implicit batching. Batch is created from given number of records
	// Non-blocking alternative is available in the WriteApi interface
	WriteRecord(ctx context.Context, line ...string) error
	// WritePoint data point into bucket.
	// WritePoint writes without implicit batching. Batch is created from given number of points
	// Non-blocking alternative is available in the WriteApi interface
	WritePoint(ctx context.Context, point ...*write.Point) error
}

// writeApiBlockingImpl implements WriteApiBlocking interface
type writeApiBlockingImpl struct {
	service      *iwrite.Service
	writeOptions *write.Options
}

// creates writeApiBlockingImpl for org and bucket with underlying client
func NewWriteApiBlockingImpl(org string, bucket string, service http.Service, writeOptions *write.Options) *writeApiBlockingImpl {
	return &writeApiBlockingImpl{service: iwrite.NewService(org, bucket, service, writeOptions), writeOptions: writeOptions}
}

func (w *writeApiBlockingImpl) write(ctx context.Context, line string) error {
	err := w.service.HandleWrite(ctx, iwrite.NewBatch(line, w.writeOptions.RetryInterval()))
	return err
}

func (w *writeApiBlockingImpl) WriteRecord(ctx context.Context, line ...string) error {
	if len(line) > 0 {
		var sb strings.Builder
		for _, line := range line {
			b := []byte(line)
			b = append(b, 0xa)
			if _, err := sb.Write(b); err != nil {
				return err
			}
		}
		return w.write(ctx, sb.String())
	}
	return nil
}

func (w *writeApiBlockingImpl) WritePoint(ctx context.Context, point ...*write.Point) error {
	line, err := w.service.EncodePoints(point...)
	if err != nil {
		return err
	}
	return w.write(ctx, line)
}
