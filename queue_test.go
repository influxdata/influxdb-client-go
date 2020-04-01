// Copyright 2020 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package influxdb2

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQueue(t *testing.T) {
	que := newQueue(2)
	assert.True(t, que.isEmpty())
	b := &batch{batch: "batch", retryInterval: 3, retries: 3}
	que.push(b)
	assert.False(t, que.isEmpty())
	b2 := que.pop()
	assert.Equal(t, b, b2)
	assert.True(t, que.isEmpty())

	que.push(b)
	que.push(b)
	assert.True(t, que.push(b))
	assert.False(t, que.isEmpty())
	que.pop()
	que.pop()
	assert.True(t, que.isEmpty())
	assert.Nil(t, que.pop())
	assert.True(t, que.isEmpty())
}
