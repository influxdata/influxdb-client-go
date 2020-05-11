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
type DeleteApi interface {
	// Delete deletes series selected by by the time range specified by start and stop arguments and optional predicate string from the bucket bucket belonging to the organization org.
	Delete(ctx context.Context, org *domain.Organization, bucket *domain.Bucket, start, stop time.Time, predicate string) error
	// Delete deletes series selected by by the time range specified by start and stop arguments and optional predicate string from the bucket with Id bucketId belonging to the organization with Id orgId.
	DeleteWithId(ctx context.Context, orgId, bucketId string, start, stop time.Time, predicate string) error
	// Delete deletes series selected by by the time range specified by start and stop arguments and optional predicate string from the bucket with name bucketName belonging to the organization with name orgName.
	DeleteWithName(ctx context.Context, orgName, bucketName string, start, stop time.Time, predicate string) error
}

type deleteApiImpl struct {
	apiClient *domain.ClientWithResponses
}

func NewDeleteApi(apiClient *domain.ClientWithResponses) DeleteApi {
	return &deleteApiImpl{
		apiClient: apiClient,
	}
}

func (d *deleteApiImpl) delete(ctx context.Context, params *domain.PostDeleteParams, conditions *domain.DeletePredicateRequest) error {
	resp, err := d.apiClient.PostDeleteWithResponse(ctx, params, domain.PostDeleteJSONRequestBody(*conditions))
	if err != nil {
		return err
	}
	if resp.JSON404 != nil {
		return domain.DomainErrorToError(resp.JSON404, resp.StatusCode())
	}
	if resp.JSON403 != nil {
		return domain.DomainErrorToError(resp.JSON403, resp.StatusCode())
	}
	if resp.JSON400 != nil {
		return domain.DomainErrorToError(resp.JSON400, resp.StatusCode())
	}
	if resp.JSONDefault != nil {
		return domain.DomainErrorToError(resp.JSONDefault, resp.StatusCode())
	}
	return nil
}

func (d *deleteApiImpl) Delete(ctx context.Context, org *domain.Organization, bucket *domain.Bucket, start, stop time.Time, predicate string) error {
	params := &domain.PostDeleteParams{
		OrgID:    org.Id,
		BucketID: bucket.Id,
	}
	conditions := &domain.DeletePredicateRequest{
		Predicate: &predicate,
		Start:     start,
		Stop:      stop,
	}
	return d.delete(ctx, params, conditions)
}

func (d *deleteApiImpl) DeleteWithId(ctx context.Context, orgId, bucketId string, start, stop time.Time, predicate string) error {
	params := &domain.PostDeleteParams{
		OrgID:    &orgId,
		BucketID: &bucketId,
	}
	conditions := &domain.DeletePredicateRequest{
		Predicate: &predicate,
		Start:     start,
		Stop:      stop,
	}
	return d.delete(ctx, params, conditions)
}

func (d *deleteApiImpl) DeleteWithName(ctx context.Context, orgName, bucketName string, start, stop time.Time, predicate string) error {
	params := &domain.PostDeleteParams{
		Org:    &orgName,
		Bucket: &bucketName,
	}
	conditions := &domain.DeletePredicateRequest{
		Predicate: &predicate,
		Start:     start,
		Stop:      stop,
	}
	return d.delete(ctx, params, conditions)
}
