// +build e2e

// Copyright 2020 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package influxdb2_test

import (
	"context"
	"fmt"
	"github.com/influxdata/influxdb-client-go/api/write"
	"strconv"
	"strings"
	"testing"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go"
	"github.com/influxdata/influxdb-client-go/api"
	"github.com/influxdata/influxdb-client-go/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetup(t *testing.T) {
	client := influxdb2.NewClientWithOptions("http://localhost:9999", "", influxdb2.DefaultOptions().SetLogLevel(2))
	response, err := client.Setup(context.Background(), "my-user", "my-password", "my-org", "my-bucket", 0)
	if err != nil {
		t.Error(err)
	}
	require.NotNil(t, response)
	authToken = *response.Auth.Token
	fmt.Println("Token:" + authToken)

	_, err = client.Setup(context.Background(), "my-user", "my-password", "my-org", "my-bucket", 0)
	require.NotNil(t, err)
	assert.Equal(t, "conflict: onboarding has already been completed", err.Error())
}

func TestWriteDeprecated(t *testing.T) {
	client := influxdb2.NewClientWithOptions("http://localhost:9999", authToken, influxdb2.DefaultOptions().SetLogLevel(3))
	writeApi := client.WriteApi("my-org", "my-bucket")
	errCh := writeApi.Errors()
	errorsCount := 0
	go func() {
		for err := range errCh {
			errorsCount++
			fmt.Println("Write error: ", err.Error())
		}
	}()
	timestamp := time.Now()
	for i, f := 0, 3.3; i < 10; i++ {
		writeApi.WriteRecord(fmt.Sprintf("test,a=%d,b=local f=%.2f,i=%di %d", i%2, f, i, timestamp.UnixNano()))
		//writeApi.Flush()
		f += 3.3
		timestamp = timestamp.Add(time.Nanosecond)
	}

	for i, f := int64(10), 33.0; i < 20; i++ {
		p := influxdb2.NewPoint("test",
			map[string]string{"a": strconv.FormatInt(i%2, 10), "b": "static"},
			map[string]interface{}{"f": f, "i": i},
			timestamp)
		writeApi.WritePoint(p)
		f += 3.3
		timestamp = timestamp.Add(time.Nanosecond)
	}

	err := client.WriteApiBlocking("my-org", "my-bucket").WritePoint(context.Background(), influxdb2.NewPointWithMeasurement("test").
		AddTag("a", "3").AddField("i", 20).AddField("f", 4.4))
	assert.Nil(t, err)

	client.Close()
	assert.Equal(t, 0, errorsCount)

}

func TestQueryRawDeprecated(t *testing.T) {
	client := influxdb2.NewClient("http://localhost:9999", authToken)

	queryApi := client.QueryApi("my-org")
	res, err := queryApi.QueryRaw(context.Background(), `from(bucket:"my-bucket")|> range(start: -24h) |> filter(fn: (r) => r._measurement == "test")`, influxdb2.DefaultDialect())
	if err != nil {
		t.Error(err)
	} else {
		fmt.Println("QueryResult:")
		fmt.Println(res)
	}
}

func TestQueryDeprecated(t *testing.T) {
	client := influxdb2.NewClient("http://localhost:9999", authToken)

	queryApi := client.QueryApi("my-org")
	fmt.Println("QueryResult")
	result, err := queryApi.Query(context.Background(), `from(bucket:"my-bucket")|> range(start: -24h) |> filter(fn: (r) => r._measurement == "test")`)
	if err != nil {
		t.Error(err)
	} else {
		rows := 0
		for result.Next() {
			rows++
			if result.TableChanged() {
				fmt.Printf("table: %s\n", result.TableMetadata().String())
			}
			fmt.Printf("row: %sv\n", result.Record().String())
		}
		if result.Err() != nil {
			t.Error(result.Err())
		}
		assert.Equal(t, 42, rows)
	}

}

func TestAuthorizationsApi(t *testing.T) {
	client := influxdb2.NewClient("http://localhost:9999", authToken)
	authApi := client.AuthorizationsApi()
	listRes, err := authApi.GetAuthorizations(context.Background())
	require.Nil(t, err)
	require.NotNil(t, listRes)
	assert.Len(t, *listRes, 1)

	orgName := "my-org"
	org, err := client.OrganizationsApi().FindOrganizationByName(context.Background(), orgName)
	require.Nil(t, err)
	require.NotNil(t, org)
	assert.Equal(t, orgName, org.Name)

	permission := &domain.Permission{
		Action: domain.PermissionActionWrite,
		Resource: domain.Resource{
			Type: domain.ResourceTypeBuckets,
		},
	}
	permissions := []domain.Permission{*permission}

	auth, err := authApi.CreateAuthorizationWithOrgId(context.Background(), *org.Id, permissions)
	require.Nil(t, err)
	require.NotNil(t, auth)
	assert.Equal(t, domain.AuthorizationUpdateRequestStatusActive, *auth.Status, *auth.Status)

	listRes, err = authApi.GetAuthorizations(context.Background())
	require.Nil(t, err)
	require.NotNil(t, listRes)
	assert.Len(t, *listRes, 2)

	listRes, err = authApi.FindAuthorizationsByUserName(context.Background(), "my-user")
	require.Nil(t, err)
	require.NotNil(t, listRes)
	assert.Len(t, *listRes, 2)

	listRes, err = authApi.FindAuthorizationsByOrgId(context.Background(), *org.Id)
	require.Nil(t, err)
	require.NotNil(t, listRes)
	assert.Len(t, *listRes, 2)

	listRes, err = authApi.FindAuthorizationsByOrgName(context.Background(), "my-org")
	require.Nil(t, err)
	require.NotNil(t, listRes)
	assert.Len(t, *listRes, 2)

	listRes, err = authApi.FindAuthorizationsByOrgName(context.Background(), "not-existent-org")
	require.Nil(t, listRes)
	require.NotNil(t, err)
	//assert.Len(t, *listRes, 0)

	auth, err = authApi.UpdateAuthorizationStatus(context.Background(), *auth.Id, domain.AuthorizationUpdateRequestStatusInactive)
	require.Nil(t, err)
	require.NotNil(t, auth)
	assert.Equal(t, domain.AuthorizationUpdateRequestStatusInactive, *auth.Status, *auth.Status)

	listRes, err = authApi.GetAuthorizations(context.Background())
	require.Nil(t, err)
	require.NotNil(t, listRes)
	assert.Len(t, *listRes, 2)

	err = authApi.DeleteAuthorization(context.Background(), *auth.Id)
	require.Nil(t, err)

	listRes, err = authApi.GetAuthorizations(context.Background())
	require.Nil(t, err)
	require.NotNil(t, listRes)
	assert.Len(t, *listRes, 1)

}

func TestOrganizations(t *testing.T) {
	client := influxdb2.NewClient("http://localhost:9999", authToken)
	orgsApi := client.OrganizationsApi()
	usersApi := client.UsersApi()
	orgName := "my-org-2"
	orgDescription := "my-org 2 description"
	ctx := context.Background()
	invalidID := "aaaaaaaaaaaaaaaa"

	orgList, err := orgsApi.GetOrganizations(ctx)
	require.Nil(t, err)
	require.NotNil(t, orgList)
	assert.Len(t, *orgList, 1)

	//test error
	org, err := orgsApi.CreateOrganizationWithName(ctx, "")
	assert.NotNil(t, err)
	require.Nil(t, org)

	org, err = orgsApi.CreateOrganizationWithName(ctx, orgName)
	require.Nil(t, err)
	require.NotNil(t, org)
	assert.Equal(t, orgName, org.Name)

	//test duplicit org
	_, err = orgsApi.CreateOrganizationWithName(ctx, orgName)
	require.NotNil(t, err)

	org.Description = &orgDescription

	org, err = orgsApi.UpdateOrganization(ctx, org)
	require.Nil(t, err)
	require.NotNil(t, org)
	assert.Equal(t, orgDescription, *org.Description)

	orgList, err = orgsApi.GetOrganizations(ctx)

	require.Nil(t, err)
	require.NotNil(t, orgList)
	assert.Len(t, *orgList, 2)

	permission := &domain.Permission{
		Action: domain.PermissionActionWrite,
		Resource: domain.Resource{
			Type: domain.ResourceTypeBuckets,
		},
	}
	permissions := []domain.Permission{*permission}

	//create authorization for new org
	auth, err := client.AuthorizationsApi().CreateAuthorizationWithOrgId(context.Background(), *org.Id, permissions)
	require.Nil(t, err)
	require.NotNil(t, auth)

	// create client with new auth token without permission
	clientOrg2 := influxdb2.NewClient("http://localhost:9999", *auth.Token)

	orgList, err = clientOrg2.OrganizationsApi().GetOrganizations(ctx)
	require.Nil(t, err)
	require.NotNil(t, orgList)
	assert.Len(t, *orgList, 0)

	org2, err := orgsApi.FindOrganizationByName(ctx, orgName)
	require.Nil(t, err)
	require.NotNil(t, org2)

	//find unknown org
	org2, err = orgsApi.FindOrganizationByName(ctx, "not-existetn-org")
	assert.NotNil(t, err)
	assert.Nil(t, org2)

	//find org using token without org permission
	org2, err = clientOrg2.OrganizationsApi().FindOrganizationByName(ctx, org.Name)
	assert.NotNil(t, err)
	assert.Nil(t, org2)

	client.AuthorizationsApi().DeleteAuthorization(ctx, *auth.Id)

	members, err := orgsApi.GetMembers(ctx, org)
	require.Nil(t, err)
	require.NotNil(t, members)
	require.Len(t, *members, 0)

	user, err := usersApi.CreateUserWithName(ctx, "user-01")
	require.Nil(t, err)
	require.NotNil(t, user)

	member, err := orgsApi.AddMember(ctx, org, user)
	require.Nil(t, err)
	require.NotNil(t, member)
	assert.Equal(t, *user.Id, *member.Id)
	assert.Equal(t, user.Name, member.Name)

	// Add member with invalid id
	member, err = orgsApi.AddMemberWithId(ctx, *org.Id, invalidID)
	assert.NotNil(t, err)
	assert.Nil(t, member)

	members, err = orgsApi.GetMembers(ctx, org)
	require.Nil(t, err)
	require.NotNil(t, members)
	require.Len(t, *members, 1)

	// get member with invalid id
	members, err = orgsApi.GetMembersWithId(ctx, invalidID)
	assert.NotNil(t, err)
	assert.Nil(t, members)

	org2, err = orgsApi.FindOrganizationById(ctx, *org.Id)
	require.Nil(t, err)
	require.NotNil(t, org2)
	assert.Equal(t, org.Name, org2.Name)

	// find invalid id
	org2, err = orgsApi.FindOrganizationById(ctx, invalidID)
	assert.NotNil(t, err)
	assert.Nil(t, org2)

	orgs, err := orgsApi.FindOrganizationsByUserId(ctx, *user.Id)
	require.Nil(t, err)
	require.NotNil(t, orgs)
	require.Len(t, *orgs, 1)
	assert.Equal(t, org.Name, (*orgs)[0].Name)

	// look for not existent
	orgs, err = orgsApi.FindOrganizationsByUserId(ctx, invalidID)
	assert.Nil(t, err)
	assert.NotNil(t, orgs)
	assert.Len(t, *orgs, 0)

	orgName2 := "my-org-3"

	org2, err = orgsApi.CreateOrganizationWithName(ctx, orgName2)
	require.Nil(t, err)
	require.NotNil(t, org2)
	assert.Equal(t, orgName2, org2.Name)

	orgList, err = orgsApi.GetOrganizations(ctx)
	require.Nil(t, err)
	require.NotNil(t, orgList)
	assert.Len(t, *orgList, 3)

	owners, err := orgsApi.GetOwners(ctx, org2)
	assert.Nil(t, err)
	assert.NotNil(t, owners)
	assert.Len(t, *owners, 1)

	//get owners with invalid id
	owners, err = orgsApi.GetOwnersWithId(ctx, invalidID)
	assert.NotNil(t, err)
	assert.Nil(t, owners)

	owner, err := orgsApi.AddOwner(ctx, org2, user)
	require.Nil(t, err)
	require.NotNil(t, owner)

	// add owner with invalid ID
	owner, err = orgsApi.AddOwnerWithId(ctx, *org2.Id, invalidID)
	assert.NotNil(t, err)
	assert.Nil(t, owner)

	owners, err = orgsApi.GetOwners(ctx, org2)
	require.Nil(t, err)
	require.NotNil(t, owners)
	assert.Len(t, *owners, 2)

	u, err := usersApi.FindUserByName(ctx, "my-user")
	require.Nil(t, err)
	require.NotNil(t, u)

	err = orgsApi.RemoveOwner(ctx, org2, u)
	require.Nil(t, err)

	// remove owner with invalid ID
	err = orgsApi.RemoveOwnerWithId(ctx, invalidID, invalidID)
	assert.NotNil(t, err)

	owners, err = orgsApi.GetOwners(ctx, org2)
	require.Nil(t, err)
	require.NotNil(t, owners)
	assert.Len(t, *owners, 1)

	orgs, err = orgsApi.FindOrganizationsByUserId(ctx, *user.Id)
	require.Nil(t, err)
	require.NotNil(t, orgs)
	require.Len(t, *orgs, 2)

	err = orgsApi.RemoveMember(ctx, org, user)
	require.Nil(t, err)

	// remove invalid memberID
	err = orgsApi.RemoveMemberWithId(ctx, invalidID, invalidID)
	assert.NotNil(t, err)

	members, err = orgsApi.GetMembers(ctx, org)
	require.Nil(t, err)
	require.NotNil(t, members)
	require.Len(t, *members, 0)

	err = orgsApi.DeleteOrganization(ctx, org)
	require.Nil(t, err)

	err = orgsApi.DeleteOrganization(ctx, org2)
	assert.Nil(t, err)

	// delete invalid org
	err = orgsApi.DeleteOrganizationWithId(ctx, invalidID)
	assert.NotNil(t, err)

	orgList, err = orgsApi.GetOrganizations(ctx)
	require.Nil(t, err)
	require.NotNil(t, orgList)
	assert.Len(t, *orgList, 1)

	err = usersApi.DeleteUser(ctx, user)
	require.Nil(t, err)

}

func TestUsers(t *testing.T) {
	client := influxdb2.NewClient("http://localhost:9999", authToken)

	usersApi := client.UsersApi()

	me, err := usersApi.Me(context.Background())
	assert.Nil(t, err)
	assert.NotNil(t, me)

	users, err := usersApi.GetUsers(context.Background())
	require.Nil(t, err)
	require.NotNil(t, users)
	assert.Len(t, *users, 1)

	user, err := usersApi.CreateUserWithName(context.Background(), "user-01")
	require.Nil(t, err)
	require.NotNil(t, user)

	users, err = usersApi.GetUsers(context.Background())
	require.Nil(t, err)
	require.NotNil(t, users)
	assert.Len(t, *users, 2)

	status := domain.UserStatusInactive
	user.Status = &status
	user, err = usersApi.UpdateUser(context.Background(), user)
	require.Nil(t, err)
	require.NotNil(t, user)
	assert.Equal(t, status, *user.Status)

	user, err = usersApi.FindUserById(context.Background(), *user.Id)
	require.Nil(t, err)
	require.NotNil(t, user)

	err = usersApi.UpdateUserPassword(context.Background(), user, "my-password")
	require.Nil(t, err)

	err = usersApi.DeleteUser(context.Background(), user)
	require.Nil(t, err)

	users, err = usersApi.GetUsers(context.Background())
	require.Nil(t, err)
	require.NotNil(t, users)
	assert.Len(t, *users, 1)
}

func TestDelete(t *testing.T) {
	ctx := context.Background()
	client := influxdb2.NewClient("http://localhost:9999", authToken)
	writeApi := client.WriteApiBlocking("my-org", "my-bucket")
	queryApi := client.QueryApi("my-org")
	tmStart := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	writeF := func(start time.Time, count int64) time.Time {
		tm := start
		for i, f := int64(0), 0.0; i < count; i++ {
			p := write.NewPoint("test",
				map[string]string{"a": strconv.FormatInt(i%2, 10), "b": "static"},
				map[string]interface{}{"f": f, "i": i},
				tm)
			err := writeApi.WritePoint(ctx, p)
			require.Nil(t, err, err)
			f += 1.2
			tm = tm.Add(time.Minute)
		}
		return tm
	}
	countF := func(start, stop time.Time) int64 {
		result, err := queryApi.Query(ctx, `from(bucket:"my-bucket")|> range(start: `+start.Format(time.RFC3339)+`, stop:`+stop.Format(time.RFC3339)+`) 
		|> filter(fn: (r) => r._measurement == "test" and r._field == "f")
		|> drop(columns: ["a", "b"])
		|> pivot(rowKey:["_time"], columnKey: ["_field"], valueColumn: "_value")
		|> count(column: "f")`)

		require.Nil(t, err, err)
		count := int64(0)
		if result.Next() {
			require.NotNil(t, result.Record().ValueByKey("f"))
			count = result.Record().ValueByKey("f").(int64)
		}
		return count
	}
	tmEnd := writeF(tmStart, 100)
	assert.Equal(t, int64(100), countF(tmStart, tmEnd))
	deleteApi := client.DeleteApi()

	org, err := client.OrganizationsApi().FindOrganizationByName(ctx, "my-org")
	require.Nil(t, err, err)
	require.NotNil(t, org)

	bucket, err := client.BucketsApi().FindBucketByName(ctx, "my-bucket")
	require.Nil(t, err, err)
	require.NotNil(t, bucket)

	err = deleteApi.DeleteWithName(ctx, "my-org", "my-bucket", tmStart, tmEnd, "")
	require.Nil(t, err, err)
	assert.Equal(t, int64(0), countF(tmStart, tmEnd))

	tmEnd = writeF(tmStart, 100)
	assert.Equal(t, int64(100), countF(tmStart, tmEnd))

	err = deleteApi.DeleteWithId(ctx, *org.Id, *bucket.Id, tmStart, tmEnd, "a=1")
	require.Nil(t, err, err)
	assert.Equal(t, int64(50), countF(tmStart, tmEnd))

	err = deleteApi.Delete(ctx, org, bucket, tmStart.Add(50*time.Minute), tmEnd, "b=static")
	require.Nil(t, err, err)
	assert.Equal(t, int64(25), countF(tmStart, tmEnd))

	err = deleteApi.DeleteWithName(ctx, "org", "my-bucket", tmStart.Add(50*time.Minute), tmEnd, "b=static")
	require.NotNil(t, err, err)
	assert.True(t, strings.Contains(err.Error(), "not found"))

	err = deleteApi.DeleteWithName(ctx, "my-org", "bucket", tmStart.Add(50*time.Minute), tmEnd, "b=static")
	require.NotNil(t, err, err)
	assert.True(t, strings.Contains(err.Error(), "not found"))
}

func TestBuckets(t *testing.T) {
	ctx := context.Background()
	client := influxdb2.NewClient("http://localhost:9999", authToken)

	bucketsApi := client.BucketsApi()

	buckets, err := bucketsApi.GetBuckets(ctx)
	require.Nil(t, err, err)
	require.NotNil(t, buckets)
	//at least three buckets, my-bucket and two system buckets.
	assert.True(t, len(*buckets) > 2)

	// test find existing bucket
	bucket, err := bucketsApi.FindBucketByName(ctx, "my-bucket")
	require.Nil(t, err, err)
	require.NotNil(t, bucket)
	assert.Equal(t, "my-bucket", bucket.Name)

	//test find non-existing bucket
	bucket, err = bucketsApi.FindBucketByName(ctx, "not existing bucket")
	assert.NotNil(t, err)
	assert.Nil(t, bucket)

	// create organizatiton for bucket
	org, err := client.OrganizationsApi().CreateOrganizationWithName(ctx, "bucket-org")
	require.Nil(t, err)
	require.NotNil(t, org)

	// collect all buckets including system ones created for new organization
	buckets, err = bucketsApi.GetBuckets(ctx)
	require.Nil(t, err, err)
	//store #all buckets before creating new ones
	bucketsNum := len(*buckets)

	name := "bucket-x"
	b, err := bucketsApi.CreateBucketWithName(ctx, org, name, domain.RetentionRule{EverySeconds: 3600 * 1}, domain.RetentionRule{EverySeconds: 3600 * 24})
	require.Nil(t, err, err)
	require.NotNil(t, b)
	assert.Equal(t, name, b.Name)
	assert.Len(t, b.RetentionRules, 1)

	// Test update
	desc := "bucket description"
	b.Description = &desc
	b.RetentionRules = []domain.RetentionRule{{EverySeconds: 60}}
	b, err = bucketsApi.UpdateBucket(ctx, b)
	require.Nil(t, err, err)
	require.NotNil(t, b)
	assert.Equal(t, name, b.Name)
	assert.Equal(t, desc, *b.Description)
	assert.Len(t, b.RetentionRules, 1)

	// Test owners
	userOwner, err := client.UsersApi().CreateUserWithName(ctx, "bucket-owner")
	require.Nil(t, err, err)
	require.NotNil(t, userOwner)

	owners, err := bucketsApi.GetOwners(ctx, b)
	require.Nil(t, err, err)
	require.NotNil(t, owners)
	assert.Len(t, *owners, 0)

	owner, err := bucketsApi.AddOwner(ctx, b, userOwner)
	require.Nil(t, err, err)
	require.NotNil(t, owner)
	assert.Equal(t, *userOwner.Id, *owner.Id)

	owners, err = bucketsApi.GetOwners(ctx, b)
	require.Nil(t, err, err)
	require.NotNil(t, owners)
	assert.Len(t, *owners, 1)

	err = bucketsApi.RemoveOwnerWithId(ctx, *b.Id, *(&(*owners)[0]).Id)
	require.Nil(t, err, err)

	owners, err = bucketsApi.GetOwners(ctx, b)
	require.Nil(t, err, err)
	require.NotNil(t, owners)
	assert.Len(t, *owners, 0)

	//test failures
	_, err = bucketsApi.AddOwnerWithId(ctx, "000000000000000", *userOwner.Id)
	assert.NotNil(t, err)

	_, err = bucketsApi.AddOwnerWithId(ctx, *b.Id, "000000000000000")
	assert.NotNil(t, err)

	_, err = bucketsApi.GetOwnersWithId(ctx, "000000000000000")
	assert.NotNil(t, err)

	err = bucketsApi.RemoveOwnerWithId(ctx, *b.Id, "000000000000000")
	assert.NotNil(t, err)

	err = bucketsApi.RemoveOwnerWithId(ctx, "000000000000000", *userOwner.Id)
	assert.NotNil(t, err)

	// Test members
	userMember, err := client.UsersApi().CreateUserWithName(ctx, "bucket-member")
	require.Nil(t, err, err)
	require.NotNil(t, userMember)

	members, err := bucketsApi.GetMembers(ctx, b)
	require.Nil(t, err, err)
	require.NotNil(t, members)
	assert.Len(t, *members, 0)

	member, err := bucketsApi.AddMember(ctx, b, userMember)
	require.Nil(t, err, err)
	require.NotNil(t, member)
	assert.Equal(t, *userMember.Id, *member.Id)

	members, err = bucketsApi.GetMembers(ctx, b)
	require.Nil(t, err, err)
	require.NotNil(t, members)
	assert.Len(t, *members, 1)

	err = bucketsApi.RemoveMemberWithId(ctx, *b.Id, *(&(*members)[0]).Id)
	require.Nil(t, err, err)

	members, err = bucketsApi.GetMembers(ctx, b)
	require.Nil(t, err, err)
	require.NotNil(t, members)
	assert.Len(t, *members, 0)

	//test failures
	_, err = bucketsApi.AddMemberWithId(ctx, "000000000000000", *userMember.Id)
	assert.NotNil(t, err)

	_, err = bucketsApi.AddMemberWithId(ctx, *b.Id, "000000000000000")
	assert.NotNil(t, err)

	_, err = bucketsApi.GetMembersWithId(ctx, "000000000000000")
	assert.NotNil(t, err)

	err = bucketsApi.RemoveMemberWithId(ctx, *b.Id, "000000000000000")
	assert.NotNil(t, err)

	err = bucketsApi.RemoveMemberWithId(ctx, "000000000000000", *userMember.Id)
	assert.NotNil(t, err)

	err = bucketsApi.DeleteBucketWithId(ctx, *b.Id)
	assert.Nil(t, err, err)

	err = client.UsersApi().DeleteUser(ctx, userOwner)
	assert.Nil(t, err, err)

	err = client.UsersApi().DeleteUser(ctx, userMember)
	assert.Nil(t, err, err)

	//test failures
	_, err = bucketsApi.FindBucketById(ctx, *b.Id)
	assert.NotNil(t, err)

	_, err = bucketsApi.UpdateBucket(ctx, b)
	assert.NotNil(t, err)

	b.OrgID = b.Id
	_, err = bucketsApi.CreateBucket(ctx, b)
	assert.NotNil(t, err)

	// create bucket by object
	b = &domain.Bucket{
		Description:    &desc,
		Name:           name,
		OrgID:          org.Id,
		RetentionRules: []domain.RetentionRule{{EverySeconds: 3600}},
	}

	b, err = bucketsApi.CreateBucket(ctx, b)
	require.Nil(t, err, err)
	require.NotNil(t, b)
	assert.Equal(t, name, b.Name)
	assert.Equal(t, *org.Id, *b.OrgID)
	assert.Equal(t, desc, *b.Description)
	assert.Len(t, b.RetentionRules, 1)

	// fail duplicit name
	_, err = bucketsApi.CreateBucketWithName(ctx, org, b.Name)
	assert.NotNil(t, err)

	// fail org not found
	_, err = bucketsApi.CreateBucketWithNameWithId(ctx, *b.Id, b.Name)
	assert.NotNil(t, err)

	err = bucketsApi.DeleteBucketWithId(ctx, *b.Id)
	assert.Nil(t, err, err)

	err = bucketsApi.DeleteBucketWithId(ctx, *b.Id)
	assert.NotNil(t, err)

	// create new buckets inside org
	for i := 0; i < 30; i++ {
		name := fmt.Sprintf("bucket-%03d", i)
		b, err := bucketsApi.CreateBucketWithName(ctx, org, name)
		require.Nil(t, err, err)
		require.NotNil(t, b)
		assert.Equal(t, name, b.Name)
	}

	// test paging, 1st page
	buckets, err = bucketsApi.GetBuckets(ctx)
	require.Nil(t, err, err)
	require.NotNil(t, buckets)
	assert.Len(t, *buckets, 20)
	// test paging, 2nd, last page
	buckets, err = bucketsApi.GetBuckets(ctx, api.PagingWithOffset(20))
	require.Nil(t, err, err)
	require.NotNil(t, buckets)
	//+2 is a bug, when using offset>4 there are returned also system buckets
	assert.Len(t, *buckets, 10+2+bucketsNum)
	// test paging with increase limit to cover all buckets
	buckets, err = bucketsApi.GetBuckets(ctx, api.PagingWithLimit(100))
	require.Nil(t, err, err)
	require.NotNil(t, buckets)
	assert.Len(t, *buckets, 30+bucketsNum)
	// test filtering buckets by org id
	buckets, err = bucketsApi.FindBucketsByOrgId(ctx, *org.Id, api.PagingWithLimit(100))
	require.Nil(t, err, err)
	require.NotNil(t, buckets)
	assert.Len(t, *buckets, 30+2)
	// test filtering buckets by org name
	buckets, err = bucketsApi.FindBucketsByOrgName(ctx, org.Name, api.PagingWithLimit(100))
	require.Nil(t, err, err)
	require.NotNil(t, buckets)
	assert.Len(t, *buckets, 30+2)
	// delete buckete
	for _, b := range *buckets {
		if strings.HasPrefix(b.Name, "bucket-") {
			err = bucketsApi.DeleteBucket(ctx, &b)
			assert.Nil(t, err, err)
		}
	}
	// check all created buckets deleted
	buckets, err = bucketsApi.FindBucketsByOrgName(ctx, org.Name, api.PagingWithLimit(100))
	require.Nil(t, err, err)
	require.NotNil(t, buckets)
	assert.Len(t, *buckets, 2)

	err = client.OrganizationsApi().DeleteOrganization(ctx, org)
	assert.Nil(t, err, err)

	// should fail with org not found
	_, err = bucketsApi.FindBucketsByOrgName(ctx, org.Name, api.PagingWithLimit(100))
	assert.NotNil(t, err)
}

func TestLabels(t *testing.T) {
	client := influxdb2.NewClientWithOptions("http://localhost:9999", authToken, influxdb2.DefaultOptions().SetLogLevel(3))
	labelsApi := client.LabelsApi()
	orgApi := client.OrganizationsApi()

	ctx := context.Background()

	myorg, err := orgApi.FindOrganizationByName(ctx, "my-org")
	require.Nil(t, err, err)
	require.NotNil(t, myorg)

	labels, err := labelsApi.GetLabels(ctx)
	require.Nil(t, err, err)
	require.NotNil(t, labels)
	assert.Len(t, *labels, 0)

	labelName := "Active State"
	props := map[string]string{"color": "#33ffddd", "description": "Marks org active"}
	label, err := labelsApi.CreateLabelWithName(ctx, myorg, labelName, props)
	require.Nil(t, err, err)
	require.NotNil(t, label)
	assert.Equal(t, labelName, *label.Name)
	require.NotNil(t, label.Properties)
	assert.Equal(t, props, label.Properties.AdditionalProperties)

	//remove properties
	label.Properties.AdditionalProperties = map[string]string{"color": "", "description": ""}
	label2, err := labelsApi.UpdateLabel(ctx, label)
	require.Nil(t, err, err)
	require.NotNil(t, label2)
	assert.Equal(t, labelName, *label2.Name)
	assert.Nil(t, label2.Properties)

	label2, err = labelsApi.FindLabelById(ctx, *label.Id)
	require.Nil(t, err, err)
	require.NotNil(t, label2)
	assert.Equal(t, labelName, *label2.Name)

	label2, err = labelsApi.FindLabelById(ctx, "000000000000000")
	require.NotNil(t, err, err)
	require.Nil(t, label2)

	label2, err = labelsApi.FindLabelByName(ctx, *myorg.Id, labelName)
	require.Nil(t, err, err)
	require.NotNil(t, label2)
	assert.Equal(t, labelName, *label2.Name)

	label2, err = labelsApi.FindLabelByName(ctx, *myorg.Id, "wrong label")
	require.NotNil(t, err, err)
	require.Nil(t, label2)

	labels, err = labelsApi.GetLabels(ctx)
	require.Nil(t, err, err)
	require.NotNil(t, labels)
	assert.Len(t, *labels, 1)

	labels, err = labelsApi.FindLabelsByOrg(ctx, myorg)
	require.Nil(t, err, err)
	require.NotNil(t, labels)
	assert.Len(t, *labels, 1)

	labels, err = labelsApi.FindLabelsByOrgId(ctx, *myorg.Id)
	require.Nil(t, err, err)
	require.NotNil(t, labels)
	assert.Len(t, *labels, 1)

	labels, err = labelsApi.FindLabelsByOrgId(ctx, "000000000000000")
	require.NotNil(t, err, err)
	require.Nil(t, labels)

	// duplicate label
	label2, err = labelsApi.CreateLabelWithName(ctx, myorg, labelName, nil)
	require.NotNil(t, err, err)
	require.Nil(t, label2)

	labels, err = orgApi.GetLabels(ctx, myorg)
	require.Nil(t, err, err)
	require.NotNil(t, labels)
	assert.Len(t, *labels, 0)

	org, err := orgApi.CreateOrganizationWithName(ctx, "org1")
	require.Nil(t, err, err)
	require.NotNil(t, org)

	labels, err = orgApi.GetLabels(ctx, org)
	require.Nil(t, err, err)
	require.NotNil(t, labels)
	assert.Len(t, *labels, 0)

	labelx, err := orgApi.AddLabel(ctx, org, label)
	require.Nil(t, err, err)
	require.NotNil(t, labelx)

	labels, err = orgApi.GetLabels(ctx, org)
	require.Nil(t, err, err)
	require.NotNil(t, labels)
	assert.Len(t, *labels, 1)

	err = orgApi.RemoveLabel(ctx, org, label)
	require.Nil(t, err, err)

	labels, err = orgApi.GetLabels(ctx, org)
	require.Nil(t, err, err)
	require.NotNil(t, labels)
	assert.Len(t, *labels, 0)

	labels, err = orgApi.GetLabelsWithId(ctx, "000000000000000")
	require.NotNil(t, err, err)
	require.Nil(t, labels)

	label2, err = orgApi.AddLabelWithId(ctx, *org.Id, "000000000000000")
	require.NotNil(t, err, err)
	require.Nil(t, label2)

	label2, err = orgApi.AddLabelWithId(ctx, "000000000000000", "000000000000000")
	require.NotNil(t, err, err)
	require.Nil(t, label2)

	err = orgApi.RemoveLabelWithId(ctx, *org.Id, "000000000000000")
	require.NotNil(t, err, err)
	require.Nil(t, label2)

	err = orgApi.RemoveLabelWithId(ctx, "000000000000000", "000000000000000")
	require.NotNil(t, err, err)
	require.Nil(t, label2)

	err = orgApi.DeleteOrganization(ctx, org)
	assert.Nil(t, err, err)

	err = labelsApi.DeleteLabel(ctx, label)
	require.Nil(t, err, err)

	err = labelsApi.DeleteLabel(ctx, label)
	require.NotNil(t, err, err)
}
