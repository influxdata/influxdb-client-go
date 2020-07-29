// Copyright 2020 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package api

import (
	"github.com/influxdata/influxdb-client-go/domain"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPaging(t *testing.T) {
	paging := &Paging{}
	PagingWithOffset(10)(paging)
	PagingWithLimit(100)(paging)
	PagingWithSortBy("name")(paging)
	PagingWithDescending(true)(paging)
	assert.True(t, paging.descending)
	assert.Equal(t, domain.Limit(100), paging.limit)
	assert.Equal(t, domain.Offset(10), paging.offset)
	assert.Equal(t, "name", paging.sortBy)

	paging = &Paging{}
	PagingWithLimit(0)(paging)
	assert.Equal(t, domain.Limit(1), paging.limit)

	paging = &Paging{}
	PagingWithLimit(1000)(paging)
	assert.Equal(t, domain.Limit(100), paging.limit)
}
