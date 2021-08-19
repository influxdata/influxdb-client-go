// Copyright 2020-2021 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package write

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQueue(t *testing.T) {
	que := newQueue(2)
	assert.True(t, que.isEmpty())
	assert.Nil(t, que.first())
	assert.Nil(t, que.pop())
	b := &Batch{Batch: "batch", RetryDelay: 3, RetryAttempts: 3}
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
