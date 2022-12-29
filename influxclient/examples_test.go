// +build e2e

// Copyright 2020-2021 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package influxclient_test

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

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
	bucket, err := client.BucketsAPI().FindOne(ctx, &influxclient.Filter{Name: bucketName})
	if err != nil {
		panic(err)
	}

	// Get an organization that will own the mapping
	org, err := client.OrganizationAPI().FindOne(ctx, &influxclient.Filter{Name: orgName})
	if err != nil {
		panic(err)
	}

	yes := true
	// Fill required fields of the DBRP struct
	dbrp := model.DBRPCreate{
		BucketID:        *bucket.Id,
		Database:        bucket.Name,
		Default:         &yes,
		OrgID:           org.Id,
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
	defer apiClient.DeleteDBRPID(ctx, &model.DeleteDBRPIDAllParams{
		DeleteDBRPIDParams: model.DeleteDBRPIDParams{OrgID: org.Id}, DbrpID: safeId(newDbrp.Id),
	}) // only for E2E tests

	// Check generated response
	fmt.Fprintf(os.Stderr, "\tCreated DBRP: %#v\n", newDbrp)

	// Output:
}

func ExampleClient_Query() {
	// Create client
	client, err := influxclient.New(influxclient.Params{
		ServerURL:    "https://eu-central-1-1.aws.cloud2.influxdata.com/",
		AuthToken:    "my-token",
		Organization: "my-org",
	})

	// Define query parameters
	params := struct {
		Since       string  `json:"since"`
		GreaterThan float64 `json:"greaterThan"`
	}{
		"-10m",
		23.0,
	}
	// Prepare a query
	query := `from(bucket: "iot_center") 
		|> range(start: duration(v: params.since)) 
		|> filter(fn: (r) => r._measurement == "environment")
		|> filter(fn: (r) => r._field == "Temperature")
		|> filter(fn: (r) => r._value > params.greaterThan)`

	// Execute query
	res, err := client.Query(context.Background(), query, params)
	if err != nil {
		panic(err)
	}

	// Make sure query result is always closed
	defer res.Close()

	// Declare custom type for data
	val := &struct {
		Time   time.Time `csv:"_time"`
		Temp   float64   `csv:"_value"`
		Sensor string    `csv:"sensor"`
	}{}

	tw := tabwriter.NewWriter(os.Stdout, 10, 4, 2, ' ', 0)
	fmt.Fprintf(tw, "Time\tTemp\tSensor\n")

	// Iterate over result set
	for res.NextSection() {
		for res.NextRow() {
			err = res.Decode(val)
			if err != nil {
				fmt.Fprintf(tw, "%v\n", err)
				continue
			}
			fmt.Fprintf(tw, "%s\t%.2f\t%s\n", val.Time.String(), val.Temp, val.Sensor)
		}
	}
	tw.Flush()
	if res.Err() != nil {
		panic(res.Err())
	}
}

func ExampleBucketsAPI() {
	// Create a new client using an InfluxDB server base URL and an authentication token
	client, err := influxclient.New(influxclient.Params{
		ServerURL: serverURL,
		AuthToken: authToken})

	// Get Buckets API client
	bucketsAPI := client.BucketsAPI()
	ctx := context.Background()

	// Get organization that will own new bucket
	org, err := client.OrganizationAPI().FindOne(ctx, &influxclient.Filter{Name: orgName})
	if err != nil {
		panic(err)
	}

	// Create bucket with 1 day retention policy
	bucket, err := bucketsAPI.Create(ctx, &model.Bucket{
		OrgID: org.Id,
		Name: "bucket-sensors",
		RetentionRules: []model.RetentionRule{
			{
				EverySeconds: 3600 * 24,
			},
		},
	})
	if err != nil {
		panic(err)
	}
	defer bucketsAPI.Delete(ctx, safeId(bucket.Id)) // only for E2E tests

	// Update description of the bucket
	desc := "Bucket for sensor data"
	bucket.Description = &desc
	bucket, err = bucketsAPI.Update(ctx, bucket)
	if err != nil {
		panic(err)
	}

	// Output:
}

func ExampleOrganizationsAPI() {
	// Create a new client using an InfluxDB server base URL and an authentication token
	client, err := influxclient.New(influxclient.Params{
		ServerURL: serverURL,
		AuthToken: authToken})

	// Get Organizations API client
	orgAPI := client.OrganizationAPI()
	ctx := context.Background()

	// Create new organization
	org, err := orgAPI.Create(ctx, &model.Organization{Name: orgName+"-2"})
	if err != nil {
		panic(err)
	}
	defer orgAPI.Delete(ctx, safeId(org.Id)) // only for E2E tests

	orgDescription := "My second org"
	org.Description = &orgDescription
	org, err = orgAPI.Update(ctx, org)
	if err != nil {
		panic(err)
	}

	// Create new user to add to org
	newUser, err := client.UsersAPI().Create(ctx, &model.User{Name: userName+"-2"})
	if err != nil {
		panic(err)
	}
	defer client.UsersAPI().Delete(ctx, safeId(newUser.Id)) // only for E2E tests

	// Add new user to organization
	err = orgAPI.AddMember(ctx, *org.Id, *newUser.Id)
	if err != nil {
		panic(err)
	}

	// Output:
}

func ExampleAuthorizationsAPI() {
	// Create a new client using an InfluxDB server base URL and an authentication token
	client, err := influxclient.New(influxclient.Params{
		ServerURL: serverURL,
		AuthToken: authToken})
	ctx := context.Background()

	// Find user to grant permission
	user, err := client.UsersAPI().FindOne(ctx, &influxclient.Filter{Name: userName})
	if err != nil {
		panic(err)
	}

	// Find organization
	org, err := client.OrganizationAPI().FindOne(ctx, &influxclient.Filter{Name: "my-org"})
	if err != nil {
		panic(err)
	}

	// group permissions
	permissions := []model.Permission{
		{
			Action: model.PermissionActionWrite,
			Resource: model.Resource{
				Type: model.ResourceTypeBuckets,
			},
		},
		{
			Action: model.PermissionActionRead,
			Resource: model.Resource{
				Type: model.ResourceTypeBuckets,
			},
		},
	}

	// create authorization object using info above
	auth := &model.Authorization{
		OrgID:       org.Id,
		Permissions: &permissions,
		UserID:      user.Id,
	}

	// grant permission and create token
	authCreated, err := client.AuthorizationsAPI().Create(ctx, auth)
	if err != nil {
		panic(err)
	}
	defer client.AuthorizationsAPI().Delete(ctx, safeId(authCreated.Id)) // only for E2E tests

	// Use token
	fmt.Fprintf(os.Stderr, "\tToken: %v\n", *authCreated.Token)

	// Output:
}

func ExampleUsersAPI() {
	// Create a new client using an InfluxDB server base URL and an authentication token
	client, err := influxclient.New(influxclient.Params{
		ServerURL: serverURL,
		AuthToken: authToken})
	ctx := context.Background()

	// Find organization
	org, err := client.OrganizationAPI().FindOne(ctx, &influxclient.Filter{Name: orgName})
	if err != nil {
		panic(err)
	}

	// Get users API client
	usersAPI := client.UsersAPI()

	// Create new user
	user, err := usersAPI.Create(ctx, &model.User{Name: "user-01"})
	if err != nil {
		panic(err)
	}
	defer usersAPI.Delete(ctx, safeId(user.Id)) // only for E2E tests

	// Set user password
	err = usersAPI.SetPassword(ctx, *user.Id, "pass-at-least-8-chars")
	if err != nil {
		panic(err)
	}

	// Add user to organization
	err = client.OrganizationAPI().AddMember(ctx, *org.Id, *user.Id)
	if err != nil {
		panic(err)
	}

	// Output:
}

func ExampleLabelsAPI() {
	// Create a new client using an InfluxDB server base URL and an authentication token
	client, err := influxclient.New(influxclient.Params{
		ServerURL: serverURL,
		AuthToken: authToken})

	// Get Labels API client
	labelsAPI := client.LabelsAPI()
	ctx := context.Background()

	// Get organization that will own label
	org, err := client.OrganizationAPI().FindOne(ctx, &influxclient.Filter{Name: orgName})
	if err != nil {
		panic(err)
	}

	labelName := "Active State"
	props := map[string]string{"color": "33ffdd", "description": "Marks org active"}
	label, err := labelsAPI.Create(ctx, &model.Label{
		OrgID: org.Id,
		Name:  &labelName,
		Properties: &model.Label_Properties{
			AdditionalProperties: props,
		},
	})
	if err != nil {
		panic(err)
	}
	defer labelsAPI.Delete(ctx, safeId(label.Id)) // only for E2E tests

	// Change color property
	label.Properties.AdditionalProperties = map[string]string{"color": "ff1122"}
	label, err = labelsAPI.Update(ctx, label)
	if err != nil {
		panic(err)
	}

	// Output:
}

func ExampleTasksAPI() {
	// Create a new client using an InfluxDB server base URL and an authentication token
	client, err := influxclient.New(influxclient.Params{
		ServerURL: serverURL,
		AuthToken: authToken})

	// Get Delete API client
	tasksAPI := client.TasksAPI()
	ctx := context.Background()

	// Get organization that will own task
	org, err := client.OrganizationAPI().FindOne(ctx, &influxclient.Filter{Name: orgName})
	if err != nil {
		panic(err)
	}

	// task flux script from https://www.influxdata.com/blog/writing-tasks-and-setting-up-alerts-for-influxdb-cloud/
	flux := `fruitCollected = from(bucket: "farming")
  |> range(start: -task.every)
  |> filter(fn: (r)  => r["_measurement"] == "totalFruitsCollected")
  |> filter(fn: (r)  => r["_field"] == "fruits")
  |> group(columns: ["farmName"])
  |> aggregateWindow(fn: sum, every: task.every)
  |> map(fn: (r) => ({
     _time: r._time,  _stop: r._stop, _start: r._start, _measurement: "fruitCollectionRate", _field: "fruits", _value: r._value, farmName: r.farmName,
  }))

fruitCollected 
  |> to(bucket: "farming")
`
	every := "1h"
	task, err := tasksAPI.Create(ctx, &model.Task{
		OrgID: *org.Id,
		Name:  "fruitCollectedRate",
		Flux:  flux,
		Every: &every,
	})
	if err != nil {
		panic(err)
	}
	defer tasksAPI.Delete(ctx, safeId(task.Id)) // only for E2E tests

	// Force running a task
	run, err := tasksAPI.RunManually(ctx, task.Id)
	if err != nil {
		panic(err)
	}

	// Print run info
	fmt.Fprint(os.Stderr, "\tForced run scheduled for ", *run.ScheduledFor, " with status ", *run.Status, "\n")
	//wait for tasks to start and be running
	<-time.After(1500 * time.Millisecond)

	// Find logs
	logs, err := tasksAPI.FindRunLogs(ctx, task.Id, *run.Id)
	if err != nil {
		panic(err)
	}

	// Print logs
	fmt.Fprintln(os.Stderr, "\tLogs:")
	for _, logEvent := range logs {
		fmt.Fprint(os.Stderr,"\t Time:", *logEvent.Time, ", Message: ", *logEvent.Message, "\n")
	}

	// Output:
}
