// +build e2e

// Copyright 2020-2021 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package influxclient_test

import (
	"context"
	"fmt"
	"os"

	"github.com/influxdata/influxdb-client-go/influxclient"
	"github.com/influxdata/influxdb-client-go/influxclient/model"
)

func ExampleClient_newClient() {
	// Create a new client using an InfluxDB server base URL and an authentication token
	client, err := influxclient.New(influxclient.Params{
		ServerURL: serverURL,
		AuthToken: authToken})

	_ = err
	_ = client

	// Output:
}

func ExampleClient_newClientWithOptions() {
	// Create a new client using an InfluxDB server base URL and an authentication token
	// Create client and set batch size to 20
	client, err := influxclient.New(influxclient.Params{
		ServerURL: serverURL,
		AuthToken: authToken,
		BatchSize: 20})

	_ = err
	_ = client

	// Output:
}

func ExampleClient_customServerAPICall() {
	// This example shows how to perform custom server API invocation for any endpoint
	// Here we will create a DBRP mapping which allows using buckets in legacy write and query (InfluxQL) endpoints

	// Create client. You need an admin token for creating DBRP mapping
	client, err := influxclient.New(influxclient.Params{
		ServerURL: serverURL,
		AuthToken: authToken})

	// Get generated client for server API calls
	apiClient := client.APIClient()
	ctx := context.Background()

	// Get a bucket we would like to query using InfluxQL
	b, err := client.BucketsAPI().FindOne(ctx, &influxclient.Filter{Name: bucketName})
	if err != nil {
		panic(err)
	}
	// Get an organization that will own the mapping
	o, err := client.OrganizationAPI().FindOne(ctx, &influxclient.Filter{Name: orgName})
	if err != nil {
		panic(err)
	}

	yes := true
	// Fill required fields of the DBRP struct
	dbrp := model.DBRPCreate{
		BucketID:        *b.Id,
		Database:        b.Name,
		Default:         &yes,
		OrgID:           o.Id,
		RetentionPolicy: "autogen",
	}

	params := &model.PostDBRPAllParams{
		Body: model.PostDBRPJSONRequestBody(dbrp),
	}
	// Call server API
	newDbrp, err := apiClient.PostDBRP(ctx, params)
	if err != nil {
		panic(err)
	}

	// Check generated response
	fmt.Fprintf(os.Stderr, "Created DBRP: %#v\n", newDbrp)

	// Output:
}
