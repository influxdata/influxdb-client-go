// Copyright 2020 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package api

import "github.com/influxdata/influxdb-client-go/v2/domain"

type PagingOption func(p *Paging)

// Paging holds pagination parameters for various Get* functions of InfluxDB 2 API
// Not the all options are usable for some Get* functions
type Paging struct {
	// Starting offset for returning items
	// Default 0.
	offset domain.Offset
	// Maximum number of items returned.
	// Default 20, minimum 1 and maximum 100.
	limit domain.Limit
	// What field should be used for sorting
	sortBy string
	// Changes sorting direction
	descending domain.Descending
	// The last resource ID from which to seek from (but not including).
	// This is to be used instead of `offset`.
	after domain.After
}

// defaultPagingOptions returns default paging options: offset 0, limit 20, default sorting, ascending
func defaultPaging() *Paging {
	return &Paging{limit: 20, offset: 0, sortBy: "", descending: false, after: ""}
}

// PagingWithLimit sets limit option - maximum number of items returned.
// Default 20, minimum 1 and maximum 100.
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

// PagingWithOffset set starting offset for returning items. Default 0.
func PagingWithOffset(offset int) PagingOption {
	return func(p *Paging) {
		p.offset = domain.Offset(offset)
	}
}

// PagingWithSortBy sets field name which should be used for sorting
func PagingWithSortBy(sortBy string) PagingOption {
	return func(p *Paging) {
		p.sortBy = sortBy
	}
}

// PagingWithDescending changes sorting direction
func PagingWithDescending(descending bool) PagingOption {
	return func(p *Paging) {
		p.descending = domain.Descending(descending)
	}
}

// PagingWithAfter set after option - the last resource ID from which to seek from (but not including).
// This is to be used instead of `offset`.
func PagingWithAfter(after string) PagingOption {
	return func(p *Paging) {
		p.after = domain.After(after)
	}
}
