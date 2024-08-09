// Copyright 2020-2021 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package api_test

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/influxdata/influxdb-client-go/v2/api"
	apiHttp "github.com/influxdata/influxdb-client-go/v2/api/http"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
	"github.com/influxdata/influxdb-client-go/v2/domain"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2/internal/examples"
)

func ExampleBucketsAPI() {
	// Create a new client using an InfluxDB server base URL and an authentication token
	client := influxdb2.NewClient("http://localhost:8086", "my-token")

	ctx := context.Background()
	// Get Buckets API client
	bucketsAPI := client.BucketsAPI()

	// Get organization that will own new bucket
	org, err := client.OrganizationsAPI().FindOrganizationByName(ctx, "my-org")
	if err != nil {
		panic(err)
	}
	// Create  a bucket with 1 day retention policy
	bucket, err := bucketsAPI.CreateBucketWithName(ctx, org, "bucket-sensors", domain.RetentionRule{EverySeconds: 3600 * 24})
	if err != nil {
		panic(err)
	}

	// Update description of the bucket
	desc := "Bucket for sensor data"
	bucket.Description = &desc
	bucket, err = bucketsAPI.UpdateBucket(ctx, bucket)
	if err != nil {
		panic(err)
	}

	// Close the client
	client.Close()
}

func ExampleWriteAPIBlocking() {
	// Create a new client using an InfluxDB server base URL and an authentication token
	client := influxdb2.NewClient("http://localhost:8086", "my-token")
	// Get blocking write client
	writeAPI := client.WriteAPIBlocking("my-org", "my-bucket")
	// write some points
	for i := 0; i < 100; i++ {
		// create data point
		p := write.NewPoint(
			"system",
			map[string]string{
				"id":       fmt.Sprintf("rack_%v", i%10),
				"vendor":   "AWS",
				"hostname": fmt.Sprintf("host_%v", i%100),
			},
			map[string]interface{}{
				"temperature": rand.Float64() * 80.0,
				"disk_free":   rand.Float64() * 1000.0,
				"disk_total":  (i/10 + 1) * 1000000,
				"mem_total":   (i/100 + 1) * 10000000,
				"mem_free":    rand.Uint64(),
			},
			time.Now())
		// write synchronously
		err := writeAPI.WritePoint(context.Background(), p)
		if err != nil {
			panic(err)
		}
	}
	// Ensures background processes finishes
	client.Close()
}

func ExampleWriteAPI() {
	// Create a new client using an InfluxDB server base URL and an authentication token
	client := influxdb2.NewClient("http://localhost:8086", "my-token")
	// Get non-blocking write client
	writeAPI := client.WriteAPI("my-org", "my-bucket")
	// write some points
	for i := 0; i < 100; i++ {
		// create point
		p := write.NewPoint(
			"system",
			map[string]string{
				"id":       fmt.Sprintf("rack_%v", i%10),
				"vendor":   "AWS",
				"hostname": fmt.Sprintf("host_%v", i%100),
			},
			map[string]interface{}{
				"temperature": rand.Float64() * 80.0,
				"disk_free":   rand.Float64() * 1000.0,
				"disk_total":  (i/10 + 1) * 1000000,
				"mem_total":   (i/100 + 1) * 10000000,
				"mem_free":    rand.Uint64(),
			},
			time.Now())
		// write asynchronously
		writeAPI.WritePoint(p)
	}
	// Force all unwritten data to be sent
	writeAPI.Flush()
	// Ensures background processes finishes
	client.Close()
}

func ExampleWriteAPI_errors() {
	// Create a new client using an InfluxDB server base URL and an authentication token
	client := influxdb2.NewClient("http://localhost:8086", "my-token")
	// Get non-blocking write client
	writeAPI := client.WriteAPI("my-org", "my-bucket")
	// Get errors channel
	errorsCh := writeAPI.Errors()
	// Create go proc for reading and logging errors
	go func() {
		for err := range errorsCh {
			fmt.Printf("write error: %s\n", err.Error())
			fmt.Printf("trace-id: %s\n", err.(*apiHttp.Error).Header.Get("Trace-ID"))
		}
	}()
	// write some points
	for i := 0; i < 100; i++ {
		// create point
		p := write.NewPointWithMeasurement("stat").
			AddTag("id", fmt.Sprintf("rack_%v", i%10)).
			AddTag("vendor", "AWS").
			AddTag("hostname", fmt.Sprintf("host_%v", i%100)).
			AddField("temperature", rand.Float64()*80.0).
			AddField("disk_free", rand.Float64()*1000.0).
			AddField("disk_total", (i/10+1)*1000000).
			AddField("mem_total", (i/100+1)*10000000).
			AddField("mem_free", rand.Uint64()).
			SetTime(time.Now())
		// write asynchronously
		writeAPI.WritePoint(p)
	}
	// Force all unwritten data to be sent
	writeAPI.Flush()
	// Ensures background processes finishes
	client.Close()
}

func ExampleQueryAPI_query() {
	// Create a new client using an InfluxDB server base URL and an authentication token
	client := influxdb2.NewClient("http://localhost:8086", "my-token")
	// Get query client
	queryAPI := client.QueryAPI("my-org")
	// get QueryTableResult
	result, err := queryAPI.Query(context.Background(), `from(bucket:"my-bucket")|> range(start: -1h) |> filter(fn: (r) => r._measurement == "stat")`)
	if err == nil {
		// Iterate over query response
		for result.Next() {
			// Notice when group key has changed
			if result.TableChanged() {
				fmt.Printf("table: %s\n", result.TableMetadata().String())
			}
			// Access data
			fmt.Printf("value: %v\n", result.Record().Value())
		}
		// check for an error
		if result.Err() != nil {
			fmt.Printf("query parsing error: %s\n", result.Err().Error())
		}
	} else {
		panic(err)
	}
	// Ensures background processes finishes
	client.Close()
}

func ExampleQueryAPI_queryWithParams() {
	// Create a new client using an InfluxDB server base URL and an authentication token
	client := influxdb2.NewClient("http://localhost:8086", "my-token")
	// Get query client
	queryAPI := client.QueryAPI("my-org")
	// Define parameters
	parameters := struct {
		Start string  `json:"start"`
		Field string  `json:"field"`
		Value float64 `json:"value"`
	}{
		"-1h",
		"temperature",
		25,
	}
	// Query with parameters
	query := `from(bucket:"my-bucket")
				|> range(start: duration(params.start)) 
				|> filter(fn: (r) => r._measurement == "stat")
				|> filter(fn: (r) => r._field == params.field)
				|> filter(fn: (r) => r._value > params.value)`

	// Get result
	result, err := queryAPI.QueryWithParams(context.Background(), query, parameters)
	if err == nil {
		// Iterate over query response
		for result.Next() {
			// Notice when group key has changed
			if result.TableChanged() {
				fmt.Printf("table: %s\n", result.TableMetadata().String())
			}
			// Access data
			fmt.Printf("value: %v\n", result.Record().Value())
		}
		// check for an error
		if result.Err() != nil {
			fmt.Printf("query parsing error: %s\n", result.Err().Error())
		}
	} else {
		panic(err)
	}
	// Ensures background processes finishes
	client.Close()
}

func ExampleQueryAPI_queryRaw() {
	// Create a new client using an InfluxDB server base URL and an authentication token
	client := influxdb2.NewClient("http://localhost:8086", "my-token")
	// Get query client
	queryAPI := client.QueryAPI("my-org")
	// Query and get complete result as a string
	// Use default dialect
	result, err := queryAPI.QueryRaw(context.Background(), `from(bucket:"my-bucket")|> range(start: -1h) |> filter(fn: (r) => r._measurement == "stat")`, api.DefaultDialect())
	if err == nil {
		fmt.Println("QueryResult:")
		fmt.Println(result)
	} else {
		panic(err)
	}
	// Ensures background processes finishes
	client.Close()
}

func ExampleOrganizationsAPI() {
	// Create a new client using an InfluxDB server base URL and an authentication token
	client := influxdb2.NewClient("http://localhost:8086", "my-token")

	// Get Organizations API client
	orgAPI := client.OrganizationsAPI()

	// Create new organization
	org, err := orgAPI.CreateOrganizationWithName(context.Background(), "org-2")
	if err != nil {
		panic(err)
	}

	orgDescription := "My second org "
	org.Description = &orgDescription

	org, err = orgAPI.UpdateOrganization(context.Background(), org)
	if err != nil {
		panic(err)
	}

	// Find user to set owner
	user, err := client.UsersAPI().FindUserByName(context.Background(), "user-01")
	if err != nil {
		panic(err)
	}

	// Add another owner (first owner is the one who create organization
	_, err = orgAPI.AddOwner(context.Background(), org, user)
	if err != nil {
		panic(err)
	}

	// Create new user to add to org
	newUser, err := client.UsersAPI().CreateUserWithName(context.Background(), "user-02")
	if err != nil {
		panic(err)
	}

	// Add new user to organization
	_, err = orgAPI.AddMember(context.Background(), org, newUser)
	if err != nil {
		panic(err)
	}
	// Ensures background processes finishes
	client.Close()
}

func ExampleAuthorizationsAPI() {
	// Create a new client using an InfluxDB server base URL and an authentication token
	client := influxdb2.NewClient("http://localhost:8086", "my-token")

	// Find user to grant permission
	user, err := client.UsersAPI().FindUserByName(context.Background(), "user-01")
	if err != nil {
		panic(err)
	}

	// Find organization
	org, err := client.OrganizationsAPI().FindOrganizationByName(context.Background(), "my-org")
	if err != nil {
		panic(err)
	}

	// create write permission for buckets
	permissionWrite := &domain.Permission{
		Action: domain.PermissionActionWrite,
		Resource: domain.Resource{
			Type: domain.ResourceTypeBuckets,
		},
	}

	// create read permission for buckets
	permissionRead := &domain.Permission{
		Action: domain.PermissionActionRead,
		Resource: domain.Resource{
			Type: domain.ResourceTypeBuckets,
		},
	}

	// group permissions
	permissions := []domain.Permission{*permissionWrite, *permissionRead}

	// create authorization object using info above
	auth := &domain.Authorization{
		OrgID:       org.Id,
		Permissions: &permissions,
		UserID:      user.Id,
	}

	// grant permission and create token
	authCreated, err := client.AuthorizationsAPI().CreateAuthorization(context.Background(), auth)
	if err != nil {
		panic(err)
	}
	// Use token
	fmt.Println("Token: ", *authCreated.Token)
	// Ensures background processes finishes
	client.Close()
}

func ExampleUsersAPI() {
	// Create a new client using an InfluxDB server base URL and an authentication token
	client := influxdb2.NewClient("http://localhost:8086", "my-token")

	// Find organization
	org, err := client.OrganizationsAPI().FindOrganizationByName(context.Background(), "my-org")
	if err != nil {
		panic(err)
	}

	// Get users API client
	usersAPI := client.UsersAPI()

	// Create new user
	user, err := usersAPI.CreateUserWithName(context.Background(), "user-01")
	if err != nil {
		panic(err)
	}

	// Set user password
	err = usersAPI.UpdateUserPassword(context.Background(), user, "pass-at-least-8-chars")
	if err != nil {
		panic(err)
	}

	// Add user to organization
	_, err = client.OrganizationsAPI().AddMember(context.Background(), org, user)
	if err != nil {
		panic(err)
	}
	// Ensures background processes finishes
	client.Close()
}

func ExampleUsersAPI_signInOut() {
	// Create a new client using an InfluxDB server base URL and empty token
	client := influxdb2.NewClient("http://localhost:8086", "")
	// Always close client at the end
	defer client.Close()

	ctx := context.Background()

	// The first call must be signIn
	err := client.UsersAPI().SignIn(ctx, "username", "password")
	if err != nil {
		panic(err)
	}

	// Perform some authorized operations
	err = client.WriteAPIBlocking("my-org", "my-bucket").WriteRecord(ctx, "test,a=rock,b=local f=1.2,i=-5i")
	if err != nil {
		panic(err)
	}

	// Sign out at the end
	err = client.UsersAPI().SignOut(ctx)
	if err != nil {
		panic(err)
	}
}

func ExampleLabelsAPI() {
	// Create a new client using an InfluxDB server base URL and an authentication token
	client := influxdb2.NewClient("http://localhost:8086", "my-token")

	ctx := context.Background()
	// Get Labels API client
	labelsAPI := client.LabelsAPI()
	// Get Organizations API client
	orgsAPI := client.OrganizationsAPI()

	// Get organization that will own label
	myorg, err := orgsAPI.FindOrganizationByName(ctx, "my-org")
	if err != nil {
		panic(err)
	}

	labelName := "Active State"
	props := map[string]string{"color": "33ffdd", "description": "Marks org active"}
	label, err := labelsAPI.CreateLabelWithName(ctx, myorg, labelName, props)
	if err != nil {
		panic(err)
	}

	// Change color property
	label.Properties.AdditionalProperties = map[string]string{"color": "ff1122"}
	label, err = labelsAPI.UpdateLabel(ctx, label)
	if err != nil {
		panic(err)
	}

	// Close the client
	client.Close()
}

func ExampleDeleteAPI() {
	// Create a new client using an InfluxDB server base URL and an authentication token
	client := influxdb2.NewClient("http://localhost:8086", "my-token")

	ctx := context.Background()
	// Get Delete API client
	deleteAPI := client.DeleteAPI()
	// Delete last hour data with tag b = static
	err := deleteAPI.DeleteWithName(ctx, "org", "my-bucket", time.Now().Add(-time.Hour), time.Now(), "b=static")
	if err != nil {
		panic(err)
	}

	// Close the client
	client.Close()
}

func ExampleTasksAPI() {
	// Create a new client using an InfluxDB server base URL and an authentication token
	client := influxdb2.NewClient("http://localhost:8086", "my-token")

	ctx := context.Background()
	// Get Delete API client
	tasksAPI := client.TasksAPI()
	// Get organization that will own task
	myorg, err := client.OrganizationsAPI().FindOrganizationByName(ctx, "my-org")
	if err != nil {
		panic(err)
	}
	// task flux script from https://www.influxdata.com/blog/writing-tasks-and-setting-up-alerts-for-influxdb-cloud/
	flux := `fruitCollected = from(bucket: “farming”)
  |> range(start: -task.every)
  |> filter(fn: (r)  => (r._measurement == “totalFruitsCollected))
  |> filter(fn: (r)  => (r._field == “fruits))
  |> group(columns: [“farmName”])
  |> aggregateWindow(fn: sum, every: task.every)
  |> map(fn: (r) => {
    return: _time: r._time,  _stop: r._stop, _start: r._start, _measurement: “fruitCollectionRate”, _field: “fruits”, _value: r._value, farmName: farmName, 
  }
})

fruitCollected 
  |> to(bucket: “farming”)
`
	task, err := tasksAPI.CreateTaskWithEvery(ctx, "fruitCollectedRate", flux, "1h", *myorg.Id)
	if err != nil {
		panic(err)
	}
	// Force running a task
	run, err := tasksAPI.RunManually(ctx, task)
	if err != nil {
		panic(err)
	}

	fmt.Println("Forced run completed on ", *run.FinishedAt, " with status ", *run.Status)

	// Print logs
	logs, err := tasksAPI.FindRunLogs(ctx, run)
	if err != nil {
		panic(err)
	}

	fmt.Println("Log:")
	for _, logEvent := range logs {
		fmt.Println(" Time:", *logEvent.Time, ", Message: ", *logEvent.Message)
	}

	// Close the client
	client.Close()
}
