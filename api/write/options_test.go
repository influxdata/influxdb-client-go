// Copyright 2020 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package write_test

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"

	"github.com/influxdata/influxdb-client-go/api/write"
)

func TestDefaultOptions(t *testing.T) {
	opts := write.DefaultOptions()
	assert.Equal(t, uint(5000), opts.BatchSize())
	assert.Equal(t, false, opts.UseGZip())
	assert.Equal(t, uint(1000), opts.FlushInterval())
	assert.Equal(t, time.Nanosecond, opts.Precision())
	assert.Equal(t, uint(10000), opts.RetryBufferLimit())
	assert.Equal(t, uint(1000), opts.RetryInterval())
	assert.Equal(t, uint(3), opts.MaxRetries())
}

func TestSettingsOptions(t *testing.T) {
	opts := write.DefaultOptions().
		SetBatchSize(5).
		SetUseGZip(true).
		SetFlushInterval(5000).
		SetPrecision(time.Millisecond).
		SetRetryBufferLimit(5).
		SetRetryInterval(5000).
		SetMaxRetries(7)
	assert.Equal(t, uint(5), opts.BatchSize())
	assert.Equal(t, true, opts.UseGZip())
	assert.Equal(t, uint(5000), opts.FlushInterval())
	assert.Equal(t, time.Millisecond, opts.Precision())
	assert.Equal(t, uint(5), opts.RetryBufferLimit())
	assert.Equal(t, uint(5000), opts.RetryInterval())
	assert.Equal(t, uint(7), opts.MaxRetries())
}
