// Copyright 2020-2021 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package write_test

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"

	"github.com/influxdata/influxdb-client-go/v2/api/write"
)

func TestDefaultOptions(t *testing.T) {
	opts := write.DefaultOptions()
	assert.EqualValues(t, 5_000, opts.BatchSize())
	assert.EqualValues(t, false, opts.UseGZip())
	assert.EqualValues(t, 1_000, opts.FlushInterval())
	assert.EqualValues(t, time.Nanosecond, opts.Precision())
	assert.EqualValues(t, 50_000, opts.RetryBufferLimit())
	assert.EqualValues(t, 5_000, opts.RetryInterval())
	assert.EqualValues(t, 5, opts.MaxRetries())
	assert.EqualValues(t, 125_000, opts.MaxRetryInterval())
	assert.EqualValues(t, 180_000, opts.MaxRetryTime())
	assert.EqualValues(t, 2, opts.ExponentialBase())
	assert.EqualValues(t, "", opts.Consistency())
	assert.Len(t, opts.DefaultTags(), 0)
}

func TestSettingsOptions(t *testing.T) {
	opts := write.DefaultOptions().
		SetBatchSize(5).
		SetUseGZip(true).
		SetFlushInterval(5_000).
		SetPrecision(time.Millisecond).
		SetRetryBufferLimit(5).
		SetRetryInterval(1_000).
		SetMaxRetries(7).
		SetMaxRetryInterval(150_000).
		SetExponentialBase(3).
		SetMaxRetryTime(200_000).
		AddDefaultTag("a", "1").
		AddDefaultTag("b", "2").
		SetConsistency(write.ConsistencyOne)
	assert.EqualValues(t, 5, opts.BatchSize())
	assert.EqualValues(t, true, opts.UseGZip())
	assert.EqualValues(t, 5000, opts.FlushInterval())
	assert.EqualValues(t, time.Millisecond, opts.Precision())
	assert.EqualValues(t, 5, opts.RetryBufferLimit())
	assert.EqualValues(t, 1000, opts.RetryInterval())
	assert.EqualValues(t, 7, opts.MaxRetries())
	assert.EqualValues(t, 150_000, opts.MaxRetryInterval())
	assert.EqualValues(t, 200_000, opts.MaxRetryTime())
	assert.EqualValues(t, 3, opts.ExponentialBase())
	assert.EqualValues(t, "one", opts.Consistency())
	assert.Len(t, opts.DefaultTags(), 2)
}
