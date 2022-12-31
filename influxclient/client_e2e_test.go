//go:build e2e
// +build e2e

// Copyright 2020-2021 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package influxclient_test

import (
	"github.com/influxdata/influxdb-client-go/v3/influxclient"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReady(t *testing.T) {
	client, ctx := newClient(t)

	up, err := client.Ready(ctx)
	require.NoError(t, err)
	assert.NotZero(t, up)
}

func TestHealth(t *testing.T) {
	client, ctx := newClient(t)

	health, err := client.Health(ctx)
	require.NoError(t, err)
	assert.NotNil(t, health)
	assert.NotEmpty(t, health.Name)
	assert.Equal(t, "influxdb", health.Name)
	assert.NotEmpty(t, health.Status)
	assert.Equal(t, "pass", string(health.Status))
	assert.NotEmpty(t, health.Commit)
	assert.NotEmpty(t, health.Version)
}

func TestPing(t *testing.T) {
	client, ctx := newClient(t)

	err := client.Ping(ctx)
	require.NoError(t, err)
}

func TestDeletePoints(t *testing.T) {
	client, ctx := newClient(t)

	err := client.DeletePoints(ctx, nil)
	assert.Error(t, err)

	err = client.DeletePoints(ctx, &influxclient.DeleteParams{})
	assert.Error(t, err)

	err = client.DeletePoints(ctx, &influxclient.DeleteParams{
		Org: orgName,
	})
	assert.Error(t, err)

	err = client.DeletePoints(ctx, &influxclient.DeleteParams{
		Bucket: bucketName,
	})
	assert.Error(t, err)

	err = client.DeletePoints(ctx, &influxclient.DeleteParams{
		Org:    orgName,
		Bucket: bucketName,
	})
	assert.Error(t, err)

	err = client.DeletePoints(ctx, &influxclient.DeleteParams{
		Org:    orgName,
		Bucket: bucketName,
		Start:  time.Now(),
	})
	assert.Error(t, err)

	err = client.DeletePoints(ctx, &influxclient.DeleteParams{
		Org:    orgName,
		Bucket: bucketName,
		Stop:   time.Now(),
	})
	assert.Error(t, err)

	err = client.DeletePoints(ctx, &influxclient.DeleteParams{
		Org:    orgName,
		Bucket: bucketName,
		// without predicate
		Start: time.Now().AddDate(0, 0, -1),
		Stop:  time.Now(),
	})
	assert.NoError(t, err)

	org, err := client.OrganizationsAPI().FindOne(ctx, &influxclient.Filter{Name: orgName})
	require.NoError(t, err)
	require.NotNil(t, org)
	bucket, err := client.BucketsAPI().FindOne(ctx, &influxclient.Filter{Name: bucketName})
	require.NoError(t, err)
	require.NotNil(t, bucket)

	err = client.DeletePoints(ctx, &influxclient.DeleteParams{
		OrgID:     *org.Id,
		BucketID:  *bucket.Id,
		Predicate: `_measurement="sensorData"`,
		Start:     time.Now().AddDate(0, 0, -1),
		Stop:      time.Now(),
	})
	assert.NoError(t, err)
}
