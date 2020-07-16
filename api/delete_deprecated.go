// Copyright 2020 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package api

import (
	"context"
	"github.com/influxdata/influxdb-client-go/domain"
	"time"
)

// DeleteApi provides methods for deleting time series data from buckets.
// Deleted series are selected by the time range specified by start and stop arguments and optional predicate string which contains condition for selecting data for deletion, such as:
// tag1="value1" and (tag2="value2" and tag3!="value3"). Empty predicate string means all data from the given time range will be deleted. See https://v2.docs.influxdata.com/v2.0/reference/syntax/delete-predicate/
// for more info about predicate syntax.
// Deprecated: Use DeleteAPI instead.
type DeleteApi interface {
	// Delete deletes series selected by by the time range specified by start and stop arguments and optional predicate string from the bucket bucket belonging to the organization org.
	Delete(ctx context.Context, org *domain.Organization, bucket *domain.Bucket, start, stop time.Time, predicate string) error
	// Delete deletes series selected by by the time range specified by start and stop arguments and optional predicate string from the bucket with Id bucketId belonging to the organization with Id orgId.
	DeleteWithId(ctx context.Context, orgId, bucketId string, start, stop time.Time, predicate string) error
	// Delete deletes series selected by by the time range specified by start and stop arguments and optional predicate string from the bucket with name bucketName belonging to the organization with name orgName.
	DeleteWithName(ctx context.Context, orgName, bucketName string, start, stop time.Time, predicate string) error
}

type deleteApiImpl struct {
	deleteAPI DeleteAPI
}

// Deprecated: Use NewDeleteAPI instead
func NewDeleteApi(apiClient *domain.ClientWithResponses) DeleteApi {
	return &deleteApiImpl{
		deleteAPI: NewDeleteAPI(apiClient),
	}
}

func (d *deleteApiImpl) Delete(ctx context.Context, org *domain.Organization, bucket *domain.Bucket, start, stop time.Time, predicate string) error {
	return d.deleteAPI.Delete(ctx, org, bucket, start, stop, predicate)
}

func (d *deleteApiImpl) DeleteWithId(ctx context.Context, orgId, bucketId string, start, stop time.Time, predicate string) error {
	return d.deleteAPI.DeleteWithID(ctx, orgId, bucketId, start, stop, predicate)
}

func (d *deleteApiImpl) DeleteWithName(ctx context.Context, orgName, bucketName string, start, stop time.Time, predicate string) error {
	return d.deleteAPI.DeleteWithName(ctx, orgName, bucketName, start, stop, predicate)
}
