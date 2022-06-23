//go:build e2e
// +build e2e

// Copyright 2020-2021 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package influxclient_test

import (
	"fmt"
	"strings"
	"testing"

	. "github.com/influxdata/influxdb-client-go/influxclient"
	"github.com/influxdata/influxdb-client-go/influxclient/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBucketsAPI(t *testing.T) {
	client, ctx := newClient(t)
	bucketsAPI := client.BucketsAPI()

	// find without filter
	buckets, err := bucketsAPI.Find(ctx, nil)
	require.Nil(t, err, err)
	require.NotNil(t, buckets)
	// at least three buckets, my-bucket (created during onboarding) and two system buckets.
	assert.True(t, len(buckets) > 2)

	// find
	buckets, err = bucketsAPI.Find(ctx, &Filter{
		OrgName: orgName,
	})
	require.Nil(t, err, err)
	require.NotNil(t, buckets)
	// at least three buckets, my-bucket and two system buckets.
	assert.True(t, len(buckets) > 2)

	// test find existing bucket by name
	buckets, err = bucketsAPI.Find(ctx, &Filter{
		Name: bucketName,
	})
	require.Nil(t, err, err)
	require.NotNil(t, buckets)
	assert.Len(t, buckets, 1)
	assert.Equal(t, bucketName, buckets[0].Name)

	// test find existing bucket by name
	bucket, err := bucketsAPI.FindOne(ctx, &Filter{
		OrgName: orgName,
		Name:    bucketName,
	})
	require.Nil(t, err, err)
	require.NotNil(t, bucket)
	assert.Equal(t, bucketName, bucket.Name)

	// test find non-existing bucket
	bucket, err = bucketsAPI.FindOne(ctx, &Filter{
		Name: "not existing bucket",
	})
	assert.Error(t, err)
	assert.Nil(t, bucket)

	// test find non-existing bucket
	buckets, err = bucketsAPI.Find(ctx, &Filter{
		Name: "not existing bucket",
	})
	assert.NoError(t, err)
	assert.NotNil(t, buckets)
	assert.Len(t, buckets, 0)

	// create organization for bucket
	org, err := client.OrganizationAPI().Create(ctx, &model.Organization{
		Name: "bucket-org",
	})
	require.NoError(t, err)
	require.NotNil(t, org)
	defer client.OrganizationAPI().Delete(ctx, safeId(org.Id))

	// test org buckets
	buckets, err = bucketsAPI.Find(ctx, &Filter{
		OrgID: *org.Id,
	})
	assert.NoError(t, err)
	assert.NotNil(t, buckets)
	//+2 for system buckets
	assert.Len(t, buckets, 2)

	// create org bucket
	name := "bucket-x"
	b, err := bucketsAPI.Create(ctx, &model.Bucket{
		OrgID: org.Id,
		Name:  name,
		RetentionRules: model.RetentionRules{
			{
				EverySeconds: 3600 * 12,
			},
		},
	})
	require.Nil(t, err, err)
	require.NotNil(t, b)
	defer bucketsAPI.Delete(ctx, safeId(b.Id))
	assert.Equal(t, name, b.Name)
	assert.Len(t, b.RetentionRules, 1)
	assert.Equal(t, b.RetentionRules[0].EverySeconds, int64(3600*12))

	// Test update
	desc := "bucket description"
	b.Description = &desc
	b.RetentionRules = []model.RetentionRule{{EverySeconds: 3600}}
	b, err = bucketsAPI.Update(ctx, b)
	require.Nil(t, err, err)
	require.NotNil(t, b)
	assert.Equal(t, name, b.Name)
	assert.Equal(t, desc, *b.Description)
	assert.Len(t, b.RetentionRules, 1)
	assert.Equal(t, b.RetentionRules[0].EverySeconds, int64(3600))

	// create org bucket with all options
	namex := "bucket-x all"
	descx := "Bucket X all"
	rpx := "0"
	schemaType := model.SchemaTypeImplicit
	bx, err := bucketsAPI.Create(ctx, &model.Bucket{
		OrgID:       org.Id,
		Name:        namex,
		Description: &descx,
		RetentionRules: model.RetentionRules{
			{
				EverySeconds: 3600 * 12,
			},
		},
		Rp:         &rpx,
		SchemaType: &schemaType,
	})
	require.Nil(t, err, err)
	require.NotNil(t, bx)
	defer bucketsAPI.Delete(ctx, safeId(bx.Id))
	assert.Equal(t, namex, bx.Name)
	assert.Equal(t, descx, *bx.Description)
	assert.Len(t, bx.RetentionRules, 1)
	assert.Equal(t, int64(3600*12), bx.RetentionRules[0].EverySeconds)
	//assert.NotNil(t, bx.SchemaType, "%v", bx.SchemaType)
	assert.Equal(t, rpx, *bx.Rp)

	// Find by ID
	b, err = bucketsAPI.FindOne(ctx, &Filter{
		ID: *b.Id,
	})
	require.Nil(t, err, err)
	require.NotNil(t, b)

	/* TODO UsersAPI does not support these in v3
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

	err = bucketsAPI.RemoveOwnerWithID(ctx, *b.Id, *(*owners)[0].Id)
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

	err = bucketsAPI.RemoveMemberWithID(ctx, *b.Id, *(*members)[0].Id)
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
	*/

	// attempt to create bucket with existing name should fail
	bucket, err = bucketsAPI.Create(ctx, &model.Bucket{
		OrgID: org.Id,
		Name:  b.Name,
	})
	assert.Error(t, err)
	assert.Nil(t, bucket)

	// delete bucket
	err = bucketsAPI.Delete(ctx, *b.Id)
	assert.Nil(t, err, err)

	// trying to delete already deleted should fail
	err = bucketsAPI.Delete(ctx, *b.Id)
	assert.Error(t, err)
}

func TestBucketsAPI_paging(t *testing.T) {
	client, ctx := newClient(t)
	bucketsAPI := client.BucketsAPI()

	// create organization for buckets
	org, err := client.OrganizationAPI().Create(ctx, &model.Organization{
		Name: "bucket-paging-org",
	})
	require.NoError(t, err)
	require.NotNil(t, org)
	defer client.OrganizationAPI().Delete(ctx, safeId(org.Id))

	// collect all buckets including system ones created for new organization
	buckets, err := bucketsAPI.Find(ctx, nil)
	require.Nil(t, err, err)
	// store #all buckets before creating new ones (typically 3 - 2x system buckets (_tasks, _monitoring) + initial bucket "my-bucket")
	bucketsNum := len(buckets)

	// create new buckets inside org
	for i := 0; i < 30; i++ {
		name := fmt.Sprintf("bucket-%03d", i)
		b, err := bucketsAPI.Create(ctx, &model.Bucket{
			OrgID: org.Id,
			Name:  name,
		})
		require.Nil(t, err, err)
		require.NotNil(t, b)
		defer bucketsAPI.Delete(ctx, safeId(b.Id))
		assert.Equal(t, name, b.Name)
	}

	// default page size is 20
	defaultPageSize := 20

	// test paging, 1st page
	buckets, err = bucketsAPI.Find(ctx, nil)
	require.Nil(t, err, err)
	require.NotNil(t, buckets)
	assert.Len(t, buckets, defaultPageSize)

	// test paging, 2nd, last page
	buckets, err = bucketsAPI.Find(ctx, &Filter{
		Offset: 20,
	})
	require.Nil(t, err, err)
	require.NotNil(t, buckets)

	// should return 15, but sometimes repeats system buckets also in 2nd page
	assert.True(t, len(buckets) >= 10+bucketsNum, "Invalid len: %d >= %d", len(buckets), 10+bucketsNum)

	// test paging with increase limit to cover all buckets
	buckets, err = bucketsAPI.Find(ctx, &Filter{
		Limit: 100,
	})
	require.Nil(t, err, err)
	require.NotNil(t, buckets)
	assert.Len(t, buckets, 30+bucketsNum)

	// test filtering buckets by org id
	buckets, err = bucketsAPI.Find(ctx, &Filter{
		OrgID: *org.Id,
		Limit: 100,
	})
	require.Nil(t, err, err)
	require.NotNil(t, buckets)
	//+2 for system buckets
	assert.Len(t, buckets, 30+2)

	// test filtering buckets by org name
	buckets, err = bucketsAPI.Find(ctx, &Filter{
		OrgName: org.Name,
		Limit:   100,
	})
	require.Nil(t, err, err)
	require.NotNil(t, buckets)
	//+2 for system buckets
	assert.Len(t, buckets, 30+2)

	// delete buckets
	for _, b := range buckets {
		if strings.HasPrefix(b.Name, "bucket-") {
			err = bucketsAPI.Delete(ctx, *b.Id)
			assert.Nil(t, err, err)
		}
	}

	// check all created buckets deleted
	buckets, err = bucketsAPI.Find(ctx, &Filter{
		OrgID: *org.Id,
		Limit: 100,
	})
	require.Nil(t, err, err)
	require.NotNil(t, buckets)
	assert.Len(t, buckets, 2)
}

func TestBucketsAPI_failing(t *testing.T) {
	client, ctx := newClient(t)
	bucketsAPI := client.BucketsAPI()

	bucket, err := bucketsAPI.Create(ctx, nil)
	assert.Error(t, err)
	assert.Nil(t, bucket)

	bucket, err = bucketsAPI.Create(ctx, &model.Bucket{})
	assert.Error(t, err)
	assert.Nil(t, bucket)

	bucket, err = bucketsAPI.Create(ctx, &model.Bucket{
		Name:  "a bucket",
		OrgID: nil,
	})
	assert.Error(t, err)
	assert.Nil(t, bucket)

	bucket, err = bucketsAPI.Update(ctx, nil)
	assert.Error(t, err)
	assert.Nil(t, bucket)

	bucket, err = bucketsAPI.Update(ctx, &model.Bucket{})
	assert.Error(t, err)
	assert.Nil(t, bucket)

	bucket, err = bucketsAPI.Update(ctx, &model.Bucket{
		Id: &notExistingID,
	})
	assert.Error(t, err)
	assert.Nil(t, bucket)

	bucket, err = bucketsAPI.Create(ctx, &model.Bucket{
		OrgID: &invalidID,
		Name:  "bucket-y",
	})
	assert.Error(t, err)
	assert.Nil(t, bucket)

	bucket, err = bucketsAPI.FindOne(ctx, &Filter{
		OrgID: invalidID,
	})
	assert.Error(t, err)
	assert.Nil(t, bucket)

	bucket, err = bucketsAPI.FindOne(ctx, &Filter{
		OrgID: invalidID,
	})
	assert.Error(t, err)
	assert.Nil(t, bucket)

	bucket, err = bucketsAPI.Create(ctx, &model.Bucket{
		OrgID: &notInitializedID,
		Name:  "bucket-y",
	})
	assert.Error(t, err)
	assert.Nil(t, bucket)
}

func TestBucketsAPI_skip(t *testing.T) {
	t.Skip("Should fail but does not")

	client, ctx := newClient(t)
	bucketsAPI := client.BucketsAPI()

	buckets, err := bucketsAPI.Find(ctx, &Filter{
		OrgID: invalidID,
	})
	assert.NotNil(t, err, "Should fail because organization with id %s does not exist", invalidID)
	assert.Nil(t, buckets, "Should be nil because organization with id %s does not exist", invalidID)
}
