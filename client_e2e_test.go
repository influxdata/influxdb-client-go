// +build e2e

// Copyright 2020 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package influxdb2_test

import (
	"context"
	"flag"
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

var authToken string

func init() {
	flag.StringVar(&authToken, "token", "", "authentication token")
}

func TestReady(t *testing.T) {
	client := influxdb2.NewClient("http://localhost:9999", "")

	ok, err := client.Ready(context.Background())
	if err != nil {
		t.Error(err)
	}
	if !ok {
		t.Fail()
	}
}

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

func TestHealth(t *testing.T) {
	client := influxdb2.NewClient("http://localhost:9999", "")

	health, err := client.Health(context.Background())
	if err != nil {
		t.Error(err)
	}
	require.NotNil(t, health)
	assert.Equal(t, domain.HealthCheckStatusPass, health.Status)
}

func TestWrite(t *testing.T) {
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
	for i, f := 0, 3.3; i < 10; i++ {
		writeApi.WriteRecord(fmt.Sprintf("test,a=%d,b=local f=%.2f,i=%di", i%2, f, i))
		//writeApi.Flush()
		f += 3.3
	}

	for i, f := int64(10), 33.0; i < 20; i++ {
		p := influxdb2.NewPoint("test",
			map[string]string{"a": strconv.FormatInt(i%2, 10), "b": "static"},
			map[string]interface{}{"f": f, "i": i},
			time.Now())
		writeApi.WritePoint(p)
		f += 3.3
	}

	err := client.WriteApiBlocking("my-org", "my-bucket").WritePoint(context.Background(), influxdb2.NewPointWithMeasurement("test").
		AddTag("a", "3").AddField("i", 20))
	assert.Nil(t, err)

	client.Close()
	assert.Equal(t, 0, errorsCount)

}

func TestQueryRaw(t *testing.T) {
	client := influxdb2.NewClient("http://localhost:9999", authToken)

	queryApi := client.QueryApi("my-org")
	res, err := queryApi.QueryRaw(context.Background(), `from(bucket:"my-bucket")|> range(start: -1h) |> filter(fn: (r) => r._measurement == "test")`, influxdb2.DefaultDialect())
	if err != nil {
		t.Error(err)
	} else {
		fmt.Println("QueryResult:")
		fmt.Println(res)
	}
}

func TestQuery(t *testing.T) {
	client := influxdb2.NewClient("http://localhost:9999", authToken)

	queryApi := client.QueryApi("my-org")
	fmt.Println("QueryResult")
	result, err := queryApi.Query(context.Background(), `from(bucket:"my-bucket")|> range(start: -24h) |> filter(fn: (r) => r._measurement == "test")`)
	if err != nil {
		t.Error(err)
	} else {
		for result.Next() {
			if result.TableChanged() {
				fmt.Printf("table: %s\n", result.TableMetadata().String())
			}
			fmt.Printf("row: %sv\n", result.Record().String())
		}
		if result.Err() != nil {
			t.Error(result.Err())
		}
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

	orgList, err := orgsApi.GetOrganizations(context.Background())
	require.Nil(t, err)
	require.NotNil(t, orgList)
	assert.Len(t, *orgList, 1)

	org, err := orgsApi.CreateOrganizationWithName(context.Background(), orgName)
	require.Nil(t, err)
	require.NotNil(t, org)
	assert.Equal(t, orgName, org.Name)

	//test duplicit org
	_, err = orgsApi.CreateOrganizationWithName(context.Background(), orgName)
	require.NotNil(t, err)

	org.Description = &orgDescription

	org, err = orgsApi.UpdateOrganization(context.Background(), org)
	require.Nil(t, err)
	require.NotNil(t, org)
	assert.Equal(t, orgDescription, *org.Description)

	orgList, err = orgsApi.GetOrganizations(context.Background())

	require.Nil(t, err)
	require.NotNil(t, orgList)
	assert.Len(t, *orgList, 2)

	members, err := orgsApi.GetMembers(context.Background(), org)
	require.Nil(t, err)
	require.NotNil(t, members)
	require.Len(t, *members, 0)

	user, err := usersApi.CreateUserWithName(context.Background(), "user-01")
	require.Nil(t, err)
	require.NotNil(t, user)

	member, err := orgsApi.AddMember(context.Background(), org, user)
	require.Nil(t, err)
	require.NotNil(t, member)
	assert.Equal(t, *user.Id, *member.Id)
	assert.Equal(t, user.Name, member.Name)

	members, err = orgsApi.GetMembers(context.Background(), org)
	require.Nil(t, err)
	require.NotNil(t, members)
	require.Len(t, *members, 1)

	org2, err := orgsApi.FindOrganizationById(context.Background(), *org.Id)
	require.Nil(t, err)
	require.NotNil(t, org2)
	assert.Equal(t, org.Name, org2.Name)

	orgs, err := orgsApi.FindOrganizationsByUserId(context.Background(), *user.Id)
	require.Nil(t, err)
	require.NotNil(t, orgs)
	require.Len(t, *orgs, 1)
	assert.Equal(t, org.Name, (*orgs)[0].Name)

	orgName2 := "my-org-3"

	org2, err = orgsApi.CreateOrganizationWithName(context.Background(), orgName2)
	require.Nil(t, err)
	require.NotNil(t, org2)
	assert.Equal(t, orgName2, org2.Name)

	orgList, err = orgsApi.GetOrganizations(context.Background())
	require.Nil(t, err)
	require.NotNil(t, orgList)
	assert.Len(t, *orgList, 3)

	owners, err := orgsApi.GetOwners(context.Background(), org2)
	require.Nil(t, err)
	require.NotNil(t, owners)
	assert.Len(t, *owners, 1)

	owner, err := orgsApi.AddOwner(context.Background(), org2, user)
	require.Nil(t, err)
	require.NotNil(t, owner)

	owners, err = orgsApi.GetOwners(context.Background(), org2)
	require.Nil(t, err)
	require.NotNil(t, owners)
	assert.Len(t, *owners, 2)

	u, err := usersApi.FindUserByName(context.Background(), "my-user")
	require.Nil(t, err)
	require.NotNil(t, u)

	err = orgsApi.RemoveOwner(context.Background(), org2, u)
	require.Nil(t, err)

	owners, err = orgsApi.GetOwners(context.Background(), org2)
	require.Nil(t, err)
	require.NotNil(t, owners)
	assert.Len(t, *owners, 1)

	orgs, err = orgsApi.FindOrganizationsByUserId(context.Background(), *user.Id)
	require.Nil(t, err)
	require.NotNil(t, orgs)
	require.Len(t, *orgs, 2)

	err = orgsApi.RemoveMember(context.Background(), org, user)
	require.Nil(t, err)

	members, err = orgsApi.GetMembers(context.Background(), org)
	require.Nil(t, err)
	require.NotNil(t, members)
	require.Len(t, *members, 0)

	err = orgsApi.DeleteOrganization(context.Background(), org)
	require.Nil(t, err)

	err = orgsApi.DeleteOrganization(context.Background(), org2)
	require.Nil(t, err)

	orgList, err = orgsApi.GetOrganizations(context.Background())
	require.Nil(t, err)
	require.NotNil(t, orgList)
	assert.Len(t, *orgList, 1)

	err = usersApi.DeleteUser(context.Background(), user)
	require.Nil(t, err)
}

func TestUsers(t *testing.T) {
	client := influxdb2.NewClient("http://localhost:9999", authToken)

	usersApi := client.UsersApi()

	me, err := usersApi.Me(context.Background())
	require.Nil(t, err)
	require.NotNil(t, me)

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
	// test find non-existing bucket, bug - returns system buckets
	//bucket, err = bucketsApi.FindBucketByName(ctx, "not existing bucket")
	//require.NotNil(t, err)
	//require.Nil(t, bucket)

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

	// No logs returned https://github.com/influxdata/influxdb/issues/18048
	//logs, err := bucketsApi.GetLogs(ctx, b)
	//require.Nil(t, err, err)
	//require.NotNil(t, logs)
	//assert.Len(t, *logs, 0)

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
