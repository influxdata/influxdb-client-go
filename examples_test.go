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
	apiClient := client.APIClient()
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
	// Fill required fields of the DBRP struct
	dbrp := domain.DBRPCreate{
		BucketID:        *b.Id,
		Database:        "my-bucket",
		Default:         &yes,
		OrgID:           o.Id,
		RetentionPolicy: "autogen",
	}

	params := &domain.PostDBRPAllParams{
		Body: domain.PostDBRPJSONRequestBody(dbrp),
	}
	// Call server API
	newDbrp, err := apiClient.PostDBRP(ctx, params)
	if err != nil {
		panic(err)
	}
	// Check generated response errors

	fmt.Printf("Created DBRP: %#v\n", newDbrp)
}

func ExampleClient_checkAPICall() {
	// This example shows how to perform custom server API invocation for checks API

	// Create client. You need an admin token for creating DBRP mapping
	client := influxdb2.NewClient("http://localhost:8086", "my-token")

	// Always close client at the end
	defer client.Close()

	ctx := context.Background()

	// Create a new threshold check
	greater := domain.GreaterThreshold{}
	greater.Value = 10.0
	lc := domain.CheckStatusLevelCRIT
	greater.Level = &lc
	greater.AllValues = &[]bool{true}[0]

	lesser := domain.LesserThreshold{}
	lesser.Value = 1.0
	lo := domain.CheckStatusLevelOK
	lesser.Level = &lo

	rang := domain.RangeThreshold{}
	rang.Min = 3.0
	rang.Max = 8.0
	lw := domain.CheckStatusLevelWARN
	rang.Level = &lw

	thresholds := []domain.Threshold{&greater, &lesser, &rang}

	// Get organization where check will be created
	org, err := client.OrganizationsAPI().FindOrganizationByName(ctx, "my-org")
	if err != nil {
		panic(err)
	}

	// Prepare necessary parameters
	msg := "Check: ${ r._check_name } is: ${ r._level }"
	flux := `from(bucket: "foo") |> range(start: -1d, stop: now()) |> aggregateWindow(every: 1m, fn: mean) |> filter(fn: (r) => r._field == "usage_user") |> yield()`
	every := "1h"
	offset := "0s"

	c := domain.ThresholdCheck{
		CheckBaseExtend: domain.CheckBaseExtend{
			CheckBase: domain.CheckBase{
				Name:   "My threshold check",
				OrgID:  *org.Id,
				Query:  domain.DashboardQuery{Text: &flux},
				Status: domain.TaskStatusTypeActive,
			},
			Every:                 &every,
			Offset:                &offset,
			StatusMessageTemplate: &msg,
		},
		Thresholds: &thresholds,
	}
	params := domain.CreateCheckAllParams{
		Body: &c,
	}
	// Call checks API using internal API client
	check, err := client.APIClient().CreateCheck(context.Background(), &params)
	if err != nil {
		panic(err)
	}
	// Optionally verify type
	if check.Type() != string(domain.ThresholdCheckTypeThreshold) {
		panic("Check type is not threshold")
	}
	// Cast check to threshold check
	thresholdCheck := check.(*domain.ThresholdCheck)
	fmt.Printf("Created threshold check with id %s\n", *thresholdCheck.Id)
}
