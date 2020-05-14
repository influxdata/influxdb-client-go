// Copyright 2020 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package api

import (
	"context"
	"github.com/influxdata/influxdb-client-go/domain"
)

// BucketsApi provides methods for managing Buckets in a InfluxDB server.
type BucketsApi interface {
	// GetBuckets returns all buckets, with the specified paging. Empty pagingOptions means the default paging (first 20 results).
	GetBuckets(ctx context.Context, pagingOptions ...PagingOption) (*[]domain.Bucket, error)
	// FindBucketByName returns a bucket found using bucketName.
	FindBucketByName(ctx context.Context, bucketName string) (*domain.Bucket, error)
	// FindBucketById returns a bucket found using bucketId.
	FindBucketById(ctx context.Context, bucketId string) (*domain.Bucket, error)
	// FindBucketsByOrgId returns buckets belonging to the organization with Id orgId, with the specified paging. Empty pagingOptions means the default paging (first 20 results).
	FindBucketsByOrgId(ctx context.Context, orgId string, pagingOptions ...PagingOption) (*[]domain.Bucket, error)
	// FindBucketsByOrgName returns buckets belonging to the organization with name orgName, with the specified paging. Empty pagingOptions means the default paging (first 20 results).
	FindBucketsByOrgName(ctx context.Context, orgName string, pagingOptions ...PagingOption) (*[]domain.Bucket, error)
	// CreateBucket creates a new bucket.
	CreateBucket(ctx context.Context, bucket *domain.Bucket) (*domain.Bucket, error)
	// CreateBucketWithName creates a new bucket with bucketName in organization org, with retention specified in rules. Empty rules means infinite retention.
	CreateBucketWithName(ctx context.Context, org *domain.Organization, bucketName string, rules ...domain.RetentionRule) (*domain.Bucket, error)
	// CreateBucketWithNameWithId creates a new bucket with bucketName in organization with orgId, with retention specified in rules. Empty rules means infinite retention.
	CreateBucketWithNameWithId(ctx context.Context, orgId, bucketName string, rules ...domain.RetentionRule) (*domain.Bucket, error)
	// UpdateBucket updates a bucket.
	UpdateBucket(ctx context.Context, bucket *domain.Bucket) (*domain.Bucket, error)
	// DeleteBucket deletes a bucket.
	DeleteBucket(ctx context.Context, bucket *domain.Bucket) error
	// DeleteBucketWithId deletes a bucket with bucketId.
	DeleteBucketWithId(ctx context.Context, bucketId string) error
	// GetMembers returns members of a bucket.
	GetMembers(ctx context.Context, bucket *domain.Bucket) (*[]domain.ResourceMember, error)
	// GetMembersWithId returns members of a bucket with bucketId.
	GetMembersWithId(ctx context.Context, bucketId string) (*[]domain.ResourceMember, error)
	// AddMember adds a member to a bucket.
	AddMember(ctx context.Context, bucket *domain.Bucket, user *domain.User) (*domain.ResourceMember, error)
	// AddMember adds a member with id memberId to a bucket with bucketId.
	AddMemberWithId(ctx context.Context, bucketId, memberId string) (*domain.ResourceMember, error)
	// RemoveMember removes a member from a bucket.
	RemoveMember(ctx context.Context, bucket *domain.Bucket, user *domain.User) error
	// RemoveMember removes a member with id memberId from a bucket with bucketId.
	RemoveMemberWithId(ctx context.Context, bucketId, memberId string) error
	// GetOwners returns owners of a bucket.
	GetOwners(ctx context.Context, bucket *domain.Bucket) (*[]domain.ResourceOwner, error)
	// GetOwnersWithId returns owners of a bucket with bucketId.
	GetOwnersWithId(ctx context.Context, bucketId string) (*[]domain.ResourceOwner, error)
	// AddOwner adds an owner to a bucket.
	AddOwner(ctx context.Context, bucket *domain.Bucket, user *domain.User) (*domain.ResourceOwner, error)
	// AddOwner adds an owner with id memberId to a bucket with bucketId.
	AddOwnerWithId(ctx context.Context, bucketId, memberId string) (*domain.ResourceOwner, error)
	// RemoveOwner removes an owner from a bucket.
	RemoveOwner(ctx context.Context, bucket *domain.Bucket, user *domain.User) error
	// RemoveOwner removes a member with id memberId from a bucket with bucketId.
	RemoveOwnerWithId(ctx context.Context, bucketId, memberId string) error
	// GetLogs returns operation log entries for a bucket, with the specified paging. Empty pagingOptions means the default paging (first 20 results).
	GetLogs(ctx context.Context, bucket *domain.Bucket, pagingOptions ...PagingOption) (*[]domain.OperationLog, error)
	//GetLogsWithId returns operation log entries for bucket with id bucketId, with the specified paging. Empty pagingOptions means the default paging (first 20 results).
	GetLogsWithId(ctx context.Context, bucketId string, pagingOptions ...PagingOption) (*[]domain.OperationLog, error)
}

type bucketsApiImpl struct {
	apiClient *domain.ClientWithResponses
}

func NewBucketsApi(apiClient *domain.ClientWithResponses) BucketsApi {
	return &bucketsApiImpl{
		apiClient: apiClient,
	}
}

func (b *bucketsApiImpl) GetBuckets(ctx context.Context, pagingOptions ...PagingOption) (*[]domain.Bucket, error) {
	return b.getBuckets(ctx, nil, pagingOptions...)
}

func (b *bucketsApiImpl) getBuckets(ctx context.Context, params *domain.GetBucketsParams, pagingOptions ...PagingOption) (*[]domain.Bucket, error) {
	if params == nil {
		params = &domain.GetBucketsParams{}
	}
	options := defaultPaging()
	for _, opt := range pagingOptions {
		opt(options)
	}
	params.Limit = &options.limit
	params.Offset = &options.offset

	response, err := b.apiClient.GetBucketsWithResponse(ctx, params)
	if err != nil {
		return nil, err
	}
	if response.JSONDefault != nil {
		return nil, domain.DomainErrorToError(response.JSONDefault, response.StatusCode())
	}
	return response.JSON200.Buckets, nil
}

func (b *bucketsApiImpl) FindBucketByName(ctx context.Context, bucketName string) (*domain.Bucket, error) {
	params := &domain.GetBucketsParams{Name: &bucketName}
	response, err := b.apiClient.GetBucketsWithResponse(ctx, params)
	if err != nil {
		return nil, err
	}
	if response.JSONDefault != nil {
		return nil, domain.DomainErrorToError(response.JSONDefault, response.StatusCode())
	}
	return &(*response.JSON200.Buckets)[0], nil
}

func (b *bucketsApiImpl) FindBucketById(ctx context.Context, bucketId string) (*domain.Bucket, error) {
	params := &domain.GetBucketsIDParams{}
	response, err := b.apiClient.GetBucketsIDWithResponse(ctx, bucketId, params)
	if err != nil {
		return nil, err
	}
	if response.JSONDefault != nil {
		return nil, domain.DomainErrorToError(response.JSONDefault, response.StatusCode())
	}
	return response.JSON200, nil
}

func (b *bucketsApiImpl) FindBucketsByOrgId(ctx context.Context, orgId string, pagingOptions ...PagingOption) (*[]domain.Bucket, error) {
	params := &domain.GetBucketsParams{OrgID: &orgId}
	return b.getBuckets(ctx, params, pagingOptions...)
}

func (b *bucketsApiImpl) FindBucketsByOrgName(ctx context.Context, orgName string, pagingOptions ...PagingOption) (*[]domain.Bucket, error) {
	params := &domain.GetBucketsParams{Org: &orgName}
	return b.getBuckets(ctx, params, pagingOptions...)
}

func (b *bucketsApiImpl) CreateBucket(ctx context.Context, bucket *domain.Bucket) (*domain.Bucket, error) {
	params := &domain.PostBucketsParams{}
	bucketReq := &domain.PostBucketRequest{
		Description:    bucket.Description,
		Name:           bucket.Name,
		OrgID:          bucket.OrgID,
		RetentionRules: bucket.RetentionRules,
		Rp:             bucket.Rp,
	}
	response, err := b.apiClient.PostBucketsWithResponse(ctx, params, domain.PostBucketsJSONRequestBody(*bucketReq))
	if err != nil {
		return nil, err
	}
	if response.JSONDefault != nil {
		return nil, domain.DomainErrorToError(response.JSONDefault, response.StatusCode())
	}
	return response.JSON201, nil
}

func (b *bucketsApiImpl) CreateBucketWithNameWithId(ctx context.Context, orgId, bucketName string, rules ...domain.RetentionRule) (*domain.Bucket, error) {
	params := &domain.PostBucketsParams{}
	bucket := &domain.PostBucketRequest{Name: bucketName, OrgID: &orgId, RetentionRules: rules}
	response, err := b.apiClient.PostBucketsWithResponse(ctx, params, domain.PostBucketsJSONRequestBody(*bucket))
	if err != nil {
		return nil, err
	}
	if response.JSON422 != nil {
		return nil, domain.DomainErrorToError(response.JSON422, response.StatusCode())
	}
	if response.JSONDefault != nil {
		return nil, domain.DomainErrorToError(response.JSONDefault, response.StatusCode())
	}
	return response.JSON201, nil
}
func (b *bucketsApiImpl) CreateBucketWithName(ctx context.Context, org *domain.Organization, bucketName string, rules ...domain.RetentionRule) (*domain.Bucket, error) {
	return b.CreateBucketWithNameWithId(ctx, *org.Id, bucketName, rules...)
}

func (b *bucketsApiImpl) DeleteBucket(ctx context.Context, bucket *domain.Bucket) error {
	return b.DeleteBucketWithId(ctx, *bucket.Id)
}

func (b *bucketsApiImpl) DeleteBucketWithId(ctx context.Context, bucketId string) error {
	params := &domain.DeleteBucketsIDParams{}
	response, err := b.apiClient.DeleteBucketsIDWithResponse(ctx, bucketId, params)
	if err != nil {
		return err
	}
	if response.JSONDefault != nil {
		return domain.DomainErrorToError(response.JSONDefault, response.StatusCode())
	}
	if response.JSON404 != nil {
		return domain.DomainErrorToError(response.JSON404, response.StatusCode())
	}
	return nil
}

func (b *bucketsApiImpl) UpdateBucket(ctx context.Context, bucket *domain.Bucket) (*domain.Bucket, error) {
	params := &domain.PatchBucketsIDParams{}
	response, err := b.apiClient.PatchBucketsIDWithResponse(ctx, *bucket.Id, params, domain.PatchBucketsIDJSONRequestBody(*bucket))
	if err != nil {
		return nil, err
	}
	if response.JSONDefault != nil {
		return nil, domain.DomainErrorToError(response.JSONDefault, response.StatusCode())
	}
	return response.JSON200, nil
}

func (b *bucketsApiImpl) GetMembers(ctx context.Context, bucket *domain.Bucket) (*[]domain.ResourceMember, error) {
	return b.GetMembersWithId(ctx, *bucket.Id)
}

func (b *bucketsApiImpl) GetMembersWithId(ctx context.Context, bucketId string) (*[]domain.ResourceMember, error) {
	params := &domain.GetBucketsIDMembersParams{}
	response, err := b.apiClient.GetBucketsIDMembersWithResponse(ctx, bucketId, params)
	if err != nil {
		return nil, err
	}
	if response.JSONDefault != nil {
		return nil, domain.DomainErrorToError(response.JSONDefault, response.StatusCode())
	}
	return response.JSON200.Users, nil
}

func (b *bucketsApiImpl) AddMember(ctx context.Context, bucket *domain.Bucket, user *domain.User) (*domain.ResourceMember, error) {
	return b.AddMemberWithId(ctx, *bucket.Id, *user.Id)
}

func (b *bucketsApiImpl) AddMemberWithId(ctx context.Context, bucketId, memberId string) (*domain.ResourceMember, error) {
	params := &domain.PostBucketsIDMembersParams{}
	body := &domain.PostBucketsIDMembersJSONRequestBody{Id: memberId}
	response, err := b.apiClient.PostBucketsIDMembersWithResponse(ctx, bucketId, params, *body)
	if err != nil {
		return nil, err
	}
	if response.JSONDefault != nil {
		return nil, domain.DomainErrorToError(response.JSONDefault, response.StatusCode())
	}
	return response.JSON201, nil
}

func (b *bucketsApiImpl) RemoveMember(ctx context.Context, bucket *domain.Bucket, user *domain.User) error {
	return b.RemoveMemberWithId(ctx, *bucket.Id, *user.Id)
}

func (b *bucketsApiImpl) RemoveMemberWithId(ctx context.Context, bucketId, memberId string) error {
	params := &domain.DeleteBucketsIDMembersIDParams{}
	response, err := b.apiClient.DeleteBucketsIDMembersIDWithResponse(ctx, bucketId, memberId, params)
	if err != nil {
		return err
	}
	if response.JSONDefault != nil {
		return domain.DomainErrorToError(response.JSONDefault, response.StatusCode())
	}
	return nil
}

func (b *bucketsApiImpl) GetOwners(ctx context.Context, bucket *domain.Bucket) (*[]domain.ResourceOwner, error) {
	return b.GetOwnersWithId(ctx, *bucket.Id)
}

func (b *bucketsApiImpl) GetOwnersWithId(ctx context.Context, bucketId string) (*[]domain.ResourceOwner, error) {
	params := &domain.GetBucketsIDOwnersParams{}
	response, err := b.apiClient.GetBucketsIDOwnersWithResponse(ctx, bucketId, params)
	if err != nil {
		return nil, err
	}
	if response.JSONDefault != nil {
		return nil, domain.DomainErrorToError(response.JSONDefault, response.StatusCode())
	}
	return response.JSON200.Users, nil
}

func (b *bucketsApiImpl) AddOwner(ctx context.Context, bucket *domain.Bucket, user *domain.User) (*domain.ResourceOwner, error) {
	return b.AddOwnerWithId(ctx, *bucket.Id, *user.Id)
}

func (b *bucketsApiImpl) AddOwnerWithId(ctx context.Context, bucketId, ownerId string) (*domain.ResourceOwner, error) {
	params := &domain.PostBucketsIDOwnersParams{}
	body := &domain.PostBucketsIDOwnersJSONRequestBody{Id: ownerId}
	response, err := b.apiClient.PostBucketsIDOwnersWithResponse(ctx, bucketId, params, *body)
	if err != nil {
		return nil, err
	}
	if response.JSONDefault != nil {
		return nil, domain.DomainErrorToError(response.JSONDefault, response.StatusCode())
	}
	return response.JSON201, nil
}

func (b *bucketsApiImpl) RemoveOwner(ctx context.Context, bucket *domain.Bucket, user *domain.User) error {
	return b.RemoveOwnerWithId(ctx, *bucket.Id, *user.Id)
}

func (b *bucketsApiImpl) RemoveOwnerWithId(ctx context.Context, bucketId, memberId string) error {
	params := &domain.DeleteBucketsIDOwnersIDParams{}
	response, err := b.apiClient.DeleteBucketsIDOwnersIDWithResponse(ctx, bucketId, memberId, params)
	if err != nil {
		return err
	}
	if response.JSONDefault != nil {
		return domain.DomainErrorToError(response.JSONDefault, response.StatusCode())
	}
	return nil
}

func (b *bucketsApiImpl) GetLogs(ctx context.Context, bucket *domain.Bucket, pagingOptions ...PagingOption) (*[]domain.OperationLog, error) {
	return b.GetLogsWithId(ctx, *bucket.Id, pagingOptions...)
}

func (b *bucketsApiImpl) GetLogsWithId(ctx context.Context, bucketId string, pagingOptions ...PagingOption) (*[]domain.OperationLog, error) {
	params := &domain.GetBucketsIDLogsParams{}
	options := defaultPaging()
	for _, opt := range pagingOptions {
		opt(options)
	}
	params.Limit = &options.limit
	params.Offset = &options.offset
	response, err := b.apiClient.GetBucketsIDLogsWithResponse(ctx, bucketId, params)
	if err != nil {
		return nil, err
	}
	if response.JSONDefault != nil {
		return nil, domain.DomainErrorToError(response.JSONDefault, response.StatusCode())
	}
	return response.JSON200.Logs, nil
}
