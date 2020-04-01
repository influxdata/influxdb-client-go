// Copyright 2020 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package influxdb2

import (
	"context"
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
	WritePoint(ctx context.Context, point ...*Point) error
}

// writeApiBlockingImpl implements WriteApiBlocking interface
type writeApiBlockingImpl struct {
	service *writeService
}

// creates writeApiBlockingImpl for org and bucket with underlying client
func newWriteApiBlockingImpl(org string, bucket string, client InfluxDBClient) *writeApiBlockingImpl {
	return &writeApiBlockingImpl{service: newWriteService(org, bucket, client)}
}

func (w *writeApiBlockingImpl) write(ctx context.Context, line string) error {
	err := w.service.handleWrite(ctx, &batch{
		batch:         line,
		retryInterval: w.service.client.Options().RetryInterval(),
	})
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

func (w *writeApiBlockingImpl) WritePoint(ctx context.Context, point ...*Point) error {
	line, err := w.service.encodePoints(point...)
	if err != nil {
		return err
	}
	return w.write(ctx, line)
}
