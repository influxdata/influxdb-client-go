// +build e2e

// Copyright 2020 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package api_test

import (
	"context"
	"fmt"
	"github.com/influxdata/influxdb-client-go/v2/log"
	"strings"
	"testing"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBucketsAPI(t *testing.T) {
	ctx := context.Background()
	client := influxdb2.NewClient(serverURL, authToken)

	bucketsAPI := client.BucketsAPI()

	buckets, err := bucketsAPI.GetBuckets(ctx)
	require.Nil(t, err, err)
	require.NotNil(t, buckets)
	//at least three buckets, my-bucket and two system buckets.
	assert.True(t, len(*buckets) > 2)

	// test find existing bucket
	bucket, err := bucketsAPI.FindBucketByName(ctx, "my-bucket")
	require.Nil(t, err, err)
	require.NotNil(t, bucket)
	assert.Equal(t, "my-bucket", bucket.Name)

	//test find non-existing bucket
	bucket, err = bucketsAPI.FindBucketByName(ctx, "not existing bucket")
	assert.NotNil(t, err)
	assert.Nil(t, bucket)

	// create organizatiton for bucket
	org, err := client.OrganizationsAPI().CreateOrganizationWithName(ctx, "bucket-org")
	require.Nil(t, err)
	require.NotNil(t, org)

	name := "bucket-x"
	b, err := bucketsAPI.CreateBucketWithName(ctx, org, name, domain.RetentionRule{EverySeconds: 3600 * 1}, domain.RetentionRule{EverySeconds: 3600 * 24})
	require.Nil(t, err, err)
	require.NotNil(t, b)
	assert.Equal(t, name, b.Name)
	assert.Len(t, b.RetentionRules, 1)

	// Test update
	desc := "bucket description"
	b.Description = &desc
	b.RetentionRules = []domain.RetentionRule{{EverySeconds: 60}}
	b, err = bucketsAPI.UpdateBucket(ctx, b)
	require.Nil(t, err, err)
	require.NotNil(t, b)
	assert.Equal(t, name, b.Name)
	assert.Equal(t, desc, *b.Description)
	assert.Len(t, b.RetentionRules, 1)

	b, err = bucketsAPI.FindBucketByID(ctx, *b.Id)
	require.Nil(t, err, err)
	require.NotNil(t, b)

	// Test owners
	userOwner, err := client.UsersAPI().CreateUserWithName(ctx, "bucket-owner")
	require.Nil(t, err, err)
	require.NotNil(t, userOwner)

	owners, err := bucketsAPI.GetOwners(ctx, b)
	require.Nil(t, err, err)
	require.NotNil(t, owners)
	assert.Len(t, *owners, 0)

	owner, err := bucketsAPI.AddOwner(ctx, b, userOwner)
	require.Nil(t, err, err)
	require.NotNil(t, owner)
	assert.Equal(t, *userOwner.Id, *owner.Id)

	owners, err = bucketsAPI.GetOwners(ctx, b)
	require.Nil(t, err, err)
	require.NotNil(t, owners)
	assert.Len(t, *owners, 1)

	err = bucketsAPI.RemoveOwner(ctx, b, &(*owners)[0].User)
	require.Nil(t, err, err)

	owners, err = bucketsAPI.GetOwners(ctx, b)
	require.Nil(t, err, err)
	require.NotNil(t, owners)
	assert.Len(t, *owners, 0)

	// Test members
	userMember, err := client.UsersAPI().CreateUserWithName(ctx, "bucket-member")
	require.Nil(t, err, err)
	require.NotNil(t, userMember)

	members, err := bucketsAPI.GetMembers(ctx, b)
	require.Nil(t, err, err)
	require.NotNil(t, members)
	assert.Len(t, *members, 0)

	member, err := bucketsAPI.AddMember(ctx, b, userMember)
	require.Nil(t, err, err)
	require.NotNil(t, member)
	assert.Equal(t, *userMember.Id, *member.Id)

	members, err = bucketsAPI.GetMembers(ctx, b)
	require.Nil(t, err, err)
	require.NotNil(t, members)
	assert.Len(t, *members, 1)

	err = bucketsAPI.RemoveMember(ctx, b, &(*members)[0].User)
	require.Nil(t, err, err)

	members, err = bucketsAPI.GetMembers(ctx, b)
	require.Nil(t, err, err)
	require.NotNil(t, members)
	assert.Len(t, *members, 0)

	err = bucketsAPI.DeleteBucketWithID(ctx, *b.Id)
	assert.Nil(t, err, err)

	err = client.UsersAPI().DeleteUser(ctx, userOwner)
	assert.Nil(t, err, err)

	err = client.UsersAPI().DeleteUser(ctx, userMember)
	assert.Nil(t, err, err)

	//test failures
	_, err = bucketsAPI.FindBucketByID(ctx, *b.Id)
	assert.NotNil(t, err)

	_, err = bucketsAPI.UpdateBucket(ctx, b)
	assert.NotNil(t, err)

	b.OrgID = b.Id
	_, err = bucketsAPI.CreateBucket(ctx, b)
	assert.NotNil(t, err)

	// create bucket by object
	b = &domain.Bucket{
		Description:    &desc,
		Name:           name,
		OrgID:          org.Id,
		RetentionRules: []domain.RetentionRule{{EverySeconds: 3600}},
	}

	b, err = bucketsAPI.CreateBucket(ctx, b)
	require.Nil(t, err, err)
	require.NotNil(t, b)
	assert.Equal(t, name, b.Name)
	assert.Equal(t, *org.Id, *b.OrgID)
	assert.Equal(t, desc, *b.Description)
	assert.Len(t, b.RetentionRules, 1)

	// fail duplicit name
	_, err = bucketsAPI.CreateBucketWithName(ctx, org, b.Name)
	assert.NotNil(t, err)

	// fail org not found
	_, err = bucketsAPI.CreateBucketWithNameWithID(ctx, *b.Id, b.Name)
	assert.NotNil(t, err)

	err = bucketsAPI.DeleteBucketWithID(ctx, *b.Id)
	assert.Nil(t, err, err)
	//delete already deleted
	err = bucketsAPI.DeleteBucketWithID(ctx, *b.Id)
	assert.NotNil(t, err)

	err = client.OrganizationsAPI().DeleteOrganization(ctx, org)
	assert.Nil(t, err, err)

	// should fail with org not found
	_, err = bucketsAPI.FindBucketsByOrgName(ctx, org.Name, api.PagingWithLimit(100))
	assert.NotNil(t, err)
}

func TestBucketsAPI_paging(t *testing.T) {
	ctx := context.Background()
	client := influxdb2.NewClientWithOptions(serverURL, authToken, influxdb2.DefaultOptions().SetLogLevel(log.DebugLevel))

	bucketsAPI := client.BucketsAPI()

	// create organizatiton for buckets
	org, err := client.OrganizationsAPI().CreateOrganizationWithName(ctx, "bucket-org")
	require.Nil(t, err)
	require.NotNil(t, org)

	// collect all buckets including system ones created for new organization
	buckets, err := bucketsAPI.GetBuckets(ctx)
	require.Nil(t, err, err)
	//store #all buckets before creating new ones (typically 5 - 2xsytem buckets (_tasks, _monitoring) + initial bucket "my-bucket")
	bucketsNum := len(*buckets)

	// create new buckets inside org
	for i := 0; i < 30; i++ {
		name := fmt.Sprintf("bucket-%03d", i)
		b, err := bucketsAPI.CreateBucketWithName(ctx, org, name)
		require.Nil(t, err, err)
		require.NotNil(t, b)
		assert.Equal(t, name, b.Name)
	}

	// test paging, 1st page
	buckets, err = bucketsAPI.GetBuckets(ctx)
	require.Nil(t, err, err)
	require.NotNil(t, buckets)
	assert.Len(t, *buckets, 20)
	// test paging, 2nd, last page
	buckets, err = bucketsAPI.GetBuckets(ctx, api.PagingWithOffset(20))
	require.Nil(t, err, err)
	require.NotNil(t, buckets)
	// should return 15, but sometimes repeats system buckets also in 2nd page
	assert.True(t, len(*buckets) >= 10+bucketsNum, "Invalid len: %d >= %d", len(*buckets), 10+bucketsNum)
	// test paging with increase limit to cover all buckets
	buckets, err = bucketsAPI.GetBuckets(ctx, api.PagingWithLimit(100))
	require.Nil(t, err, err)
	require.NotNil(t, buckets)
	assert.Len(t, *buckets, 30+bucketsNum)
	// test filtering buckets by org id
	buckets, err = bucketsAPI.FindBucketsByOrgID(ctx, *org.Id, api.PagingWithLimit(100))
	require.Nil(t, err, err)
	require.NotNil(t, buckets)
	//+2 for system buckets
	assert.Len(t, *buckets, 30+2)
	// test filtering buckets by org name
	buckets, err = bucketsAPI.FindBucketsByOrgName(ctx, org.Name, api.PagingWithLimit(100))
	require.Nil(t, err, err)
	require.NotNil(t, buckets)
	//+2 for system buckets
	assert.Len(t, *buckets, 30+2)
	// delete buckete
	for _, b := range *buckets {
		if strings.HasPrefix(b.Name, "bucket-") {
			err = bucketsAPI.DeleteBucket(ctx, &b)
			assert.Nil(t, err, err)
		}
	}
	// check all created buckets deleted
	buckets, err = bucketsAPI.FindBucketsByOrgName(ctx, org.Name, api.PagingWithLimit(100))
	require.Nil(t, err, err)
	require.NotNil(t, buckets)
	assert.Len(t, *buckets, 2)

	err = client.OrganizationsAPI().DeleteOrganization(ctx, org)
	assert.Nil(t, err, err)

}
func TestBucketsAPI_failures(t *testing.T) {
	ctx := context.Background()
	client := influxdb2.NewClient(serverURL, authToken)

	bucketsAPI := client.BucketsAPI()

	invalidID := "000000000000000"
	notExistingID := "1000000000000000"

	//test failures
	_, err := bucketsAPI.AddMemberWithID(ctx, invalidID, notExistingID)
	assert.NotNil(t, err)

	_, err = bucketsAPI.AddMemberWithID(ctx, notExistingID, invalidID)
	assert.NotNil(t, err)

	_, err = bucketsAPI.GetMembersWithID(ctx, invalidID)
	assert.NotNil(t, err)

	err = bucketsAPI.RemoveMemberWithID(ctx, notExistingID, invalidID)
	assert.NotNil(t, err)

	err = bucketsAPI.RemoveMemberWithID(ctx, invalidID, notExistingID)
	assert.NotNil(t, err)

	//test failures
	_, err = bucketsAPI.AddOwnerWithID(ctx, invalidID, notExistingID)
	assert.NotNil(t, err)

	_, err = bucketsAPI.AddOwnerWithID(ctx, notExistingID, invalidID)
	assert.NotNil(t, err)

	_, err = bucketsAPI.GetOwnersWithID(ctx, invalidID)
	assert.NotNil(t, err)

	err = bucketsAPI.RemoveOwnerWithID(ctx, notExistingID, invalidID)
	assert.NotNil(t, err)

	err = bucketsAPI.RemoveOwnerWithID(ctx, invalidID, notExistingID)
	assert.NotNil(t, err)

	//delete with invalid id
	err = bucketsAPI.DeleteBucketWithID(ctx, invalidID)
	assert.NotNil(t, err)

}

func TestBucketsAPI_requestFailing(t *testing.T) {
	ctx := context.Background()
	client := influxdb2.NewClient("htp://localhost:9990", authToken)
	bucketsAPI := client.BucketsAPI()

	anID := "1000000000000000"
	bucket := &domain.Bucket{Id: &anID, OrgID: &anID}
	user := &domain.User{Id: &anID}

	_, err := bucketsAPI.GetBuckets(ctx)
	assert.NotNil(t, err)

	_, err = bucketsAPI.FindBucketByID(ctx, anID)
	assert.NotNil(t, err)

	_, err = bucketsAPI.FindBucketByName(ctx, anID)
	assert.NotNil(t, err)

	_, err = bucketsAPI.CreateBucket(ctx, bucket)
	assert.NotNil(t, err)

	_, err = bucketsAPI.UpdateBucket(ctx, bucket)
	assert.NotNil(t, err)

	err = bucketsAPI.DeleteBucket(ctx, bucket)
	assert.NotNil(t, err)

	_, err = bucketsAPI.GetMembers(ctx, bucket)
	assert.NotNil(t, err)

	_, err = bucketsAPI.AddMember(ctx, bucket, user)
	assert.NotNil(t, err)

	err = bucketsAPI.RemoveMember(ctx, bucket, user)
	assert.NotNil(t, err)

	_, err = bucketsAPI.GetOwners(ctx, bucket)
	assert.NotNil(t, err)

	_, err = bucketsAPI.AddOwner(ctx, bucket, user)
	assert.NotNil(t, err)

	err = bucketsAPI.RemoveOwner(ctx, bucket, user)
	assert.NotNil(t, err)
}
