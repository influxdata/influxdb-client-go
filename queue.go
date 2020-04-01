// Copyright 2020 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package influxdb2

import "container/list"

type queue struct {
	list  *list.List
	limit int
}

func newQueue(limit int) *queue {
	return &queue{list: list.New(), limit: limit}
}
func (q *queue) push(batch *batch) bool {
	overWrite := false
	if q.list.Len() == q.limit {
		q.pop()
		overWrite = true
	}
	q.list.PushBack(batch)
	return overWrite
}

func (q *queue) pop() *batch {
	el := q.list.Front()
	if el != nil {
		q.list.Remove(el)
		return el.Value.(*batch)
	}
	return nil
}

func (q *queue) first() *batch {
	el := q.list.Front()
	return el.Value.(*batch)
}

func (q *queue) isEmpty() bool {
	return q.list.Len() == 0
}
