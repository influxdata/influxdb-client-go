// Copyright 2020-2021 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package influxdb2_test

import (
	"context"
	"fmt"

	"github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/domain"
)

func ExampleClient_newClient() {
	// Create a new client using an InfluxDB server base URL and an authentication token
	client := influxdb2.NewClient("http://localhost:8086", "my-token")

	// Always close client at the end
	defer client.Close()
}

func ExampleClient_newClientWithOptions() {
	// Create a new client using an InfluxDB server base URL and an authentication token
	// Create client and set batch size to 20
	client := influxdb2.NewClientWithOptions("http://localhost:8086", "my-token",
		influxdb2.DefaultOptions().SetBatchSize(20))

	// Always close client at the end
	defer client.Close()
}

func ExampleClient_customServerAPICall() {
	// This example shows how to perform custom server API invocation for any endpoint
	// Here we will create a DBRP mapping which allows using buckets in legacy write and query (InfluxQL) endpoints

	// Create client. You need an admin token for creating DBRP mapping
	client := influxdb2.NewClient("http://localhost:8086", "my-token")

	// Always close client at the end
	defer client.Close()

	// Get generated client for server API calls
	apiClient := domain.NewClientWithResponses(client.HTTPService())
	ctx := context.Background()

	// Get a bucket we would like to query using InfluxQL
	b, err := client.BucketsAPI().FindBucketByName(ctx, "my-bucket")
	if err != nil {
		panic(err)
	}
	// Get an organization that will own the mapping
	o, err := client.OrganizationsAPI().FindOrganizationByName(ctx, "my-org")
	if err != nil {
		panic(err)
	}

	yes := true
	db := "my-bucket"
	rp := "autogen"
	// Fill required fields of the DBRP struct
	dbrp := domain.DBRP{
		BucketID:        b.Id,
		Database:        &db,
		Default:         &yes,
		OrgID:           o.Id,
		RetentionPolicy: &rp,
	}

	params := &domain.PostDBRPParams{}
	// Call server API
	resp, err := apiClient.PostDBRPWithResponse(ctx, params, domain.PostDBRPJSONRequestBody(dbrp))
	if err != nil {
		panic(err)
	}
	// Check generated response errors
	if resp.JSONDefault != nil {
		panic(resp.JSONDefault.Message)
	}
	// Check generated response errors
	if resp.JSON400 != nil {
		panic(resp.JSON400.Message)
	}

	// Use API call result
	newDbrp := resp.JSON201
	fmt.Printf("Created DBRP: %#v\n", newDbrp)
}
