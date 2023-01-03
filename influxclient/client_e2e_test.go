//go:build e2e
// +build e2e

// Copyright 2020-2021 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package influxclient_test

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/influxdata/influxdb-client-go/v3/influxclient"

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

func TestWriteAndQueryExample(t *testing.T) {

	wp := influxclient.DefaultWriteParams
	// Set batch size to write 100 points in 2 batches
	wp.BatchSize = 50
	failedWrites := 0
	// Set callback for failed writes
	wp.WriteFailed = func(err error, lines []byte, attempt int, expires time.Time) bool {
		failedWrites++
		return true
	}
	// Create client with custom WriteParams
	client, err := influxclient.New(influxclient.Params{
		ServerURL:    serverURL,
		AuthToken:    authToken,
		Organization: orgName,
		WriteParams:  wp,
	})
	if err != nil {
		panic(err)
	}
	//writer := client.PointsWriter("iot_center")
	writer := client.PointsWriter("my-bucket")

	start := time.Now().Add(-100 * time.Second)
	ts := start
	// write some points
	for i := 0; i < 100; i++ {
		// create point
		p := influxclient.NewPointWithMeasurement("stat").
			AddTag("id", fmt.Sprintf("rack_%v", i%10)).
			AddTag("vendor", "AWS").
			AddTag("hostname", fmt.Sprintf("host_%v", i%100)).
			AddField("temperature", rand.Float64()*80.0).
			AddField("disk_free", rand.Float64()*1000.0).
			AddField("disk_total", (i/10+1)*1000000).
			AddField("mem_total", (i/100+1)*10000000).
			AddField("mem_free", rand.Uint64()).
			SetTimestamp(ts)
		// write asynchronously
		writer.WritePoints(p)
		ts = ts.Add(time.Second)
	}
	// writer.Close() MUST be called at the end to ensure completing background operations and cleaning resources
	writer.Close()
	assert.EqualValues(t, 0, failedWrites)
	query := fmt.Sprintf(`from(bucket: "my-bucket")
		|> range(start: %s)
		|> filter(fn: (r) => r._measurement == "stat")
		|> pivot(rowKey:["_time"], columnKey: ["_field"], valueColumn: "_value")
		|> group()
		|> count(column: "disk_free")`, start.Format(time.RFC3339Nano))

	res, err := client.Query(context.Background(), query, nil)
	require.NoError(t, err)

	defer func() {
		_ = res.Close()
	}()

	val := &struct {
		Time  time.Time `csv:"_time"`
		Count int64     `csv:"disk_free"`
	}{}

	lines := 0
	for res.NextRow() {
		err = res.Decode(val)
		require.NoError(t, err)
		lines++
	}
	require.NoError(t, res.Err())
	assert.EqualValues(t, 1, lines)
	assert.EqualValues(t, 100, val.Count)

	err = client.DeletePoints(context.Background(), &influxclient.DeleteParams{
		Bucket:    bucketName,
		Org:       orgName,
		Predicate: `_measurement="stat"`,
		Start:     start,
		Stop:      time.Now(),
	})
	assert.NoError(t, err)
	_ = client.Close()
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
