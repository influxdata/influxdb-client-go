// Copyright 2020 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package api

import "github.com/influxdata/influxdb-client-go/v2/domain"

type PagingOption func(p *Paging)

// Paging holds pagination parameters for various Get* functions of InfluxDB 2 API
type Paging struct {
	offset     domain.Offset
	limit      domain.Limit
	sortBy     string
	descending bool
}

// defaultPagingOptions returns default paging options: offset 0, limit 20, default sorting, ascending
func defaultPaging() *Paging {
	return &Paging{limit: 20, offset: 0, sortBy: "", descending: false}
}

func PagingWithLimit(limit int) PagingOption {
	return func(p *Paging) {
		if limit > 100 {
			limit = 100
		}
		if limit < 1 {
			limit = 1
		}
		p.limit = domain.Limit(limit)
	}
}

func PagingWithOffset(offset int) PagingOption {
	return func(p *Paging) {
		p.offset = domain.Offset(offset)
	}
}

func PagingWithSortBy(sortBy string) PagingOption {
	return func(p *Paging) {
		p.sortBy = sortBy
	}
}

func PagingWithDescending(descending bool) PagingOption {
	return func(p *Paging) {
		p.descending = descending
	}
}
