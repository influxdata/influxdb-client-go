// +build e2e

// Copyright 2020-2021 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package api_test

import (
	"context"
	"strconv"
	"strings"
	"testing"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
	"github.com/influxdata/influxdb-client-go/v2/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeleteAPI(t *testing.T) {
	ctx := context.Background()
	client := influxdb2.NewClient(serverURL, authToken)
	writeAPI := client.WriteAPIBlocking("my-org", "my-bucket")
	queryAPI := client.QueryAPI("my-org")
	tmStart := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	writeF := func(start time.Time, count int64) time.Time {
		tm := start
		for i, f := int64(0), 0.0; i < count; i++ {
			p := write.NewPoint("test",
				map[string]string{"a": strconv.FormatInt(i%2, 10), "b": "static"},
				map[string]interface{}{"f": f, "i": i},
				tm)
			err := writeAPI.WritePoint(ctx, p)
			require.NoError(t, err)
			f += 1.2
			tm = tm.Add(time.Minute)
		}
		return tm
	}
	countF := func(start, stop time.Time) int64 {
		result, err := queryAPI.Query(ctx, `from(bucket:"my-bucket")|> range(start: `+start.Format(time.RFC3339)+`, stop:`+stop.Format(time.RFC3339)+`) 
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
	deleteAPI := client.DeleteAPI()

	org, err := client.OrganizationsAPI().FindOrganizationByName(ctx, "my-org")
	require.Nil(t, err, err)
	require.NotNil(t, org)

	bucket, err := client.BucketsAPI().FindBucketByName(ctx, "my-bucket")
	require.Nil(t, err, err)
	require.NotNil(t, bucket)

	err = deleteAPI.DeleteWithName(ctx, "my-org", "my-bucket", tmStart, tmEnd, "")
	require.Nil(t, err, err)
	assert.Equal(t, int64(0), countF(tmStart, tmEnd))

	tmEnd = writeF(tmStart, 100)
	assert.Equal(t, int64(100), countF(tmStart, tmEnd))

	err = deleteAPI.DeleteWithID(ctx, *org.Id, *bucket.Id, tmStart, tmEnd, "a=1")
	require.Nil(t, err, err)
	assert.Equal(t, int64(50), countF(tmStart, tmEnd))

	err = deleteAPI.Delete(ctx, org, bucket, tmStart.Add(50*time.Minute), tmEnd, "b=static")
	require.Nil(t, err, err)
	assert.Equal(t, int64(25), countF(tmStart, tmEnd))

	err = deleteAPI.DeleteWithName(ctx, "org", "my-bucket", tmStart.Add(50*time.Minute), tmEnd, "b=static")
	require.NotNil(t, err, err)
	assert.True(t, strings.Contains(err.Error(), "not found"))

	err = deleteAPI.DeleteWithName(ctx, "my-org", "bucket", tmStart.Add(50*time.Minute), tmEnd, "b=static")
	require.NotNil(t, err, err)
	assert.True(t, strings.Contains(err.Error(), "not found"))
}

func TestDeleteAPI_failing(t *testing.T) {
	ctx := context.Background()
	client := influxdb2.NewClient(serverURL, authToken)
	deleteAPI := client.DeleteAPI()

	invalidID := "xcv"
	notExistentID := "1000000000000000"

	tmStart := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	tmEnd := tmStart.Add(time.Second)
	err := deleteAPI.DeleteWithID(ctx, notExistentID, invalidID, tmStart, tmEnd, "a=1")
	assert.NotNil(t, err)

	err = deleteAPI.DeleteWithID(ctx, notExistentID, notExistentID, tmStart, tmEnd, "a=1")
	assert.NotNil(t, err)

	org, err := client.OrganizationsAPI().FindOrganizationByName(ctx, "my-org")
	assert.Nil(t, err, err)
	assert.NotNil(t, org)

	bucket, err := client.BucketsAPI().FindBucketByName(ctx, "my-bucket")
	assert.Nil(t, err, err)
	assert.NotNil(t, bucket)

	org, err = client.OrganizationsAPI().CreateOrganizationWithName(ctx, "org1")
	require.Nil(t, err)
	require.NotNil(t, org)

	permission := &domain.Permission{
		Action: domain.PermissionActionWrite,
		Resource: domain.Resource{
			Type: domain.ResourceTypeOrgs,
		},
	}
	permissions := []domain.Permission{*permission}

	//create authorization for new org
	auth, err := client.AuthorizationsAPI().CreateAuthorizationWithOrgID(context.Background(), *org.Id, permissions)
	require.Nil(t, err)
	require.NotNil(t, auth)

	// create client with new auth token without permission for buckets
	clientOrg2 := influxdb2.NewClient(serverURL, *auth.Token)
	// test 403
	err = clientOrg2.DeleteAPI().Delete(ctx, org, bucket, tmStart, tmStart.Add(50*time.Minute), "b=static")
	assert.NotNil(t, err)

	err = client.AuthorizationsAPI().DeleteAuthorization(ctx, auth)
	assert.Nil(t, err)

	err = client.OrganizationsAPI().DeleteOrganization(ctx, org)
	assert.Nil(t, err)
}

func TestDeleteAPI_requestFailing(t *testing.T) {
	ctx := context.Background()
	client := influxdb2.NewClient("serverURL", authToken)
	deleteAPI := client.DeleteAPI()

	anID := "1000000000000000"

	err := deleteAPI.DeleteWithName(ctx, anID, anID, time.Now(), time.Now(), "")
	assert.NotNil(t, err)
}
