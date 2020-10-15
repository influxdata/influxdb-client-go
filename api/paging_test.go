// Copyright 2020 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package api

import (
	"testing"

	"github.com/influxdata/influxdb-client-go/v2/domain"
	"github.com/stretchr/testify/assert"
)

func TestPaging(t *testing.T) {
	paging := &Paging{}
	PagingWithOffset(10)(paging)
	PagingWithLimit(100)(paging)
	PagingWithSortBy("name")(paging)
	PagingWithDescending(true)(paging)
	PagingWithAfter("1111")(paging)
	assert.True(t, bool(paging.descending))
	assert.Equal(t, domain.Limit(100), paging.limit)
	assert.Equal(t, domain.Offset(10), paging.offset)
	assert.Equal(t, "name", paging.sortBy)
	assert.Equal(t, domain.After("1111"), paging.after)

	paging = &Paging{}
	PagingWithLimit(1)(paging)
	assert.Equal(t, domain.Limit(1), paging.limit)

	paging = &Paging{}
	PagingWithLimit(1000)(paging)
	assert.Equal(t, domain.Limit(1000), paging.limit)
}
