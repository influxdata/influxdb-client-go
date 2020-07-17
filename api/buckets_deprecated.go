// Copyright 2020 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package api

import (
	"context"
	"github.com/influxdata/influxdb-client-go/domain"
)

//lint:file-ignore ST1003 This is deprecated API to be removed in next release.

// BucketsApi provides methods for managing Buckets in a InfluxDB server.
// Deprecated: Use BucketsAPI instead.
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
}

type bucketsApiImpl struct {
	bucketsAPI BucketsAPI
}

// NewBucketsApi creates instance of BucketsApi
// Deprecated: Use NewBucketsAPI instead
func NewBucketsApi(apiClient *domain.ClientWithResponses) BucketsApi {
	return &bucketsApiImpl{
		bucketsAPI: NewBucketsAPI(apiClient),
	}
}

func (b *bucketsApiImpl) GetBuckets(ctx context.Context, pagingOptions ...PagingOption) (*[]domain.Bucket, error) {
	return b.bucketsAPI.GetBuckets(ctx, pagingOptions...)
}

func (b *bucketsApiImpl) FindBucketByName(ctx context.Context, bucketName string) (*domain.Bucket, error) {
	return b.bucketsAPI.FindBucketByName(ctx, bucketName)
}

func (b *bucketsApiImpl) FindBucketById(ctx context.Context, bucketId string) (*domain.Bucket, error) {
	return b.bucketsAPI.FindBucketByID(ctx, bucketId)
}

func (b *bucketsApiImpl) FindBucketsByOrgId(ctx context.Context, orgId string, pagingOptions ...PagingOption) (*[]domain.Bucket, error) {
	return b.bucketsAPI.FindBucketsByOrgID(ctx, orgId, pagingOptions...)
}

func (b *bucketsApiImpl) FindBucketsByOrgName(ctx context.Context, orgName string, pagingOptions ...PagingOption) (*[]domain.Bucket, error) {
	return b.bucketsAPI.FindBucketsByOrgName(ctx, orgName, pagingOptions...)
}

func (b *bucketsApiImpl) CreateBucket(ctx context.Context, bucket *domain.Bucket) (*domain.Bucket, error) {
	return b.bucketsAPI.CreateBucket(ctx, bucket)
}

func (b *bucketsApiImpl) CreateBucketWithNameWithId(ctx context.Context, orgId, bucketName string, rules ...domain.RetentionRule) (*domain.Bucket, error) {
	return b.bucketsAPI.CreateBucketWithNameWithID(ctx, orgId, bucketName, rules...)
}
func (b *bucketsApiImpl) CreateBucketWithName(ctx context.Context, org *domain.Organization, bucketName string, rules ...domain.RetentionRule) (*domain.Bucket, error) {
	return b.bucketsAPI.CreateBucketWithName(ctx, org, bucketName, rules...)
}

func (b *bucketsApiImpl) DeleteBucket(ctx context.Context, bucket *domain.Bucket) error {
	return b.bucketsAPI.DeleteBucket(ctx, bucket)
}

func (b *bucketsApiImpl) DeleteBucketWithId(ctx context.Context, bucketId string) error {
	return b.bucketsAPI.DeleteBucketWithID(ctx, bucketId)
}

func (b *bucketsApiImpl) UpdateBucket(ctx context.Context, bucket *domain.Bucket) (*domain.Bucket, error) {
	return b.bucketsAPI.UpdateBucket(ctx, bucket)
}

func (b *bucketsApiImpl) GetMembers(ctx context.Context, bucket *domain.Bucket) (*[]domain.ResourceMember, error) {
	return b.bucketsAPI.GetMembers(ctx, bucket)
}

func (b *bucketsApiImpl) GetMembersWithId(ctx context.Context, bucketId string) (*[]domain.ResourceMember, error) {
	return b.bucketsAPI.GetMembersWithID(ctx, bucketId)
}

func (b *bucketsApiImpl) AddMember(ctx context.Context, bucket *domain.Bucket, user *domain.User) (*domain.ResourceMember, error) {
	return b.bucketsAPI.AddMember(ctx, bucket, user)
}

func (b *bucketsApiImpl) AddMemberWithId(ctx context.Context, bucketId, memberId string) (*domain.ResourceMember, error) {
	return b.bucketsAPI.AddMemberWithID(ctx, bucketId, memberId)
}

func (b *bucketsApiImpl) RemoveMember(ctx context.Context, bucket *domain.Bucket, user *domain.User) error {
	return b.bucketsAPI.RemoveMember(ctx, bucket, user)
}

func (b *bucketsApiImpl) RemoveMemberWithId(ctx context.Context, bucketId, memberId string) error {
	return b.bucketsAPI.RemoveMemberWithID(ctx, bucketId, memberId)
}

func (b *bucketsApiImpl) GetOwners(ctx context.Context, bucket *domain.Bucket) (*[]domain.ResourceOwner, error) {
	return b.bucketsAPI.GetOwners(ctx, bucket)
}

func (b *bucketsApiImpl) GetOwnersWithId(ctx context.Context, bucketId string) (*[]domain.ResourceOwner, error) {
	return b.bucketsAPI.GetOwnersWithID(ctx, bucketId)
}

func (b *bucketsApiImpl) AddOwner(ctx context.Context, bucket *domain.Bucket, user *domain.User) (*domain.ResourceOwner, error) {
	return b.bucketsAPI.AddOwner(ctx, bucket, user)
}

func (b *bucketsApiImpl) AddOwnerWithId(ctx context.Context, bucketId, ownerId string) (*domain.ResourceOwner, error) {
	return b.bucketsAPI.AddOwnerWithID(ctx, bucketId, ownerId)
}

func (b *bucketsApiImpl) RemoveOwner(ctx context.Context, bucket *domain.Bucket, user *domain.User) error {
	return b.bucketsAPI.RemoveOwner(ctx, bucket, user)
}

func (b *bucketsApiImpl) RemoveOwnerWithId(ctx context.Context, bucketId, memberId string) error {
	return b.bucketsAPI.RemoveMemberWithID(ctx, bucketId, memberId)
}
