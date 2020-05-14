package api_test

import (
	"context"
	influxdb2 "github.com/influxdata/influxdb-client-go/api"
	"github.com/influxdata/influxdb-client-go/domain"
)

func ExampleBucketsApi() {
	// Create influxdb client
	client := influxdb2.NewClient("http://localhost:9999", "my-token")

	ctx := context.Background()
	// Get Organizations API client
	bucketsApi := client.BucketsApi()

	// Get organization that will own new bucket
	org, err := client.OrganizationsApi().FindOrganizationByName(ctx, "my-org")
	if err != nil {
		panic(err)
	}
	// Create  a bucket with 1 day retention policy
	bucket, err := bucketsApi.CreateBucketWithName(ctx, org, "bucket-sensors", domain.RetentionRule{EverySeconds: 3600 * 24})
	if err != nil {
		panic(err)
	}

	// Update description of the bucket
	desc := "Bucket for sensor data"
	bucket.Description = &desc
	bucket, err = bucketsApi.UpdateBucket(ctx, bucket)
	if err != nil {
		panic(err)
	}

	// Close the client
	client.Close()
}
