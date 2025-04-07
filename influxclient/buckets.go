// Copyright 2021 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package influxclient

import (
	"context"
	"fmt"

	"github.com/influxdata/influxdb-client-go/v3/influxclient/model"
)

// BucketsAPI provides methods for managing buckets in a InfluxDB server.
type BucketsAPI struct {
	client *model.Client
}

// newBucketsAPI returns new BucketsAPI instance
func newBucketsAPI(client *model.Client) *BucketsAPI {
	return &BucketsAPI{client: client}
}

// Find returns all buckets matching the given filter.
func (a *BucketsAPI) Find(ctx context.Context, filter *Filter) ([]model.Bucket, error) {
	return a.getBuckets(ctx, filter)
}

// FindOne returns one label that matches the given filter.
func (a *BucketsAPI) FindOne(ctx context.Context, filter *Filter) (*model.Bucket, error) {
	buckets, err := a.getBuckets(ctx, filter)
	if err != nil {
		return nil, err
	}
	if len(buckets) > 0 {
		return &(buckets[0]), nil
	}
	return nil, fmt.Errorf("bucket not found")
}

// Create creates a new bucket with the given information.
// The label.Name field must be non-empty.
// The returned Bucket holds the ID of the new bucket.
func (a *BucketsAPI) Create(ctx context.Context, bucket *model.Bucket) (*model.Bucket, error) {
	if bucket == nil {
		return nil, fmt.Errorf("bucket cannot be nil")
	}
	if bucket.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if bucket.OrgID == nil {
		return nil, fmt.Errorf("orgId is required")
	}
	params := &model.PostBucketsAllParams{
		Body: model.PostBucketsJSONRequestBody{
			OrgID: *(bucket.OrgID),
			Name:  bucket.Name,
		},
	}
	if bucket.Description != nil {
		params.Body.Description = bucket.Description
	}
	if bucket.RetentionRules != nil {
		params.Body.RetentionRules = &bucket.RetentionRules
	}
	if bucket.Rp != nil {
		params.Body.Rp = bucket.Rp
	}
	if bucket.SchemaType != nil {
		params.Body.SchemaType = bucket.SchemaType
	}
	return a.client.PostBuckets(ctx, params)
}

// Update updates information about a bucket.
// The bucket ID and OrgID fields must be specified.
func (a *BucketsAPI) Update(ctx context.Context, bucket *model.Bucket) (*model.Bucket, error) {
	if bucket == nil {
		return nil, fmt.Errorf("bucket cannot be nil")
	}
	if bucket.Id == nil {
		return nil, fmt.Errorf("bucket ID is required")
	}
	if bucket.OrgID == nil {
		return nil, fmt.Errorf("orgId is required")
	}
	params := &model.PatchBucketsIDAllParams{
		BucketID: *(bucket.Id),
		Body: model.PatchBucketsIDJSONRequestBody{
			Name: &bucket.Name,
		},
	}
	if bucket.Description != nil {
		params.Body.Description = bucket.Description
	}
	if bucket.RetentionRules != nil {
		rules := make(model.PatchRetentionRules, len(bucket.RetentionRules))
		for i, r := range bucket.RetentionRules {
			var patchRuleType *model.PatchRetentionRuleType
			if r.Type != nil && *r.Type != model.RetentionRuleTypeExpire {
				return nil, fmt.Errorf("unsupported retention rule type: %v", r.Type)
			}
			rules[i] = model.PatchRetentionRule{
				EverySeconds:              r.EverySeconds,
				ShardGroupDurationSeconds: r.ShardGroupDurationSeconds,
				Type:                      patchRuleType,
			}
		}
		params.Body.RetentionRules = &rules
	}
	return a.client.PatchBucketsID(ctx, params)
}

// Delete deletes the bucket with the given ID.
func (a *BucketsAPI) Delete(ctx context.Context, bucketID string) error {
	params := &model.DeleteBucketsIDAllParams{
		BucketID: bucketID,
	}
	return a.client.DeleteBucketsID(ctx, params)
}

// getBuckets create request for GET on /buckets according to the filter and validates returned structure
func (a *BucketsAPI) getBuckets(ctx context.Context, filter *Filter) ([]model.Bucket, error) {
	params := &model.GetBucketsParams{}
	if filter != nil {
		if filter.ID != "" {
			params.Id = &filter.ID
		}
		if filter.Name != "" {
			params.Name = &filter.Name
		}
		if filter.OrgName != "" {
			params.Org = &filter.OrgName
		}
		if filter.OrgID != "" {
			params.OrgID = &filter.OrgID
		}
		if filter.Limit > 0 {
			limit := model.Limit(filter.Limit)
			params.Limit = &limit
		}
		if filter.Offset > 0 {
			offset := model.Offset(filter.Offset)
			params.Offset = &offset
		}
	}
	response, err := a.client.GetBuckets(ctx, params)
	if err != nil {
		return nil, err
	}
	return *response.Buckets, nil
}
