package api_test

import (
	"context"
	"fmt"
	"github.com/influxdata/influxdb-client-go/api"
	"github.com/influxdata/influxdb-client-go/api/write"
	"github.com/influxdata/influxdb-client-go/domain"
	influxdb2 "github.com/influxdata/influxdb-client-go/internal/examples"
	"math/rand"
	"time"
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

func ExampleWriteApiBlocking() {
	// Create client
	client := influxdb2.NewClient("http://localhost:9999", "my-token")
	// Get blocking write client
	writeApi := client.WriteApiBlocking("my-org", "my-bucket")
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
		err := writeApi.WritePoint(context.Background(), p)
		if err != nil {
			panic(err)
		}
	}
	// Ensures background processes finishes
	client.Close()
}

func ExampleWriteApi() {
	// Create client
	client := influxdb2.NewClient("http://localhost:9999", "my-token")
	// Get non-blocking write client
	writeApi := client.WriteApi("my-org", "my-bucket")
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
		writeApi.WritePoint(p)
	}
	// Force all unwritten data to be sent
	writeApi.Flush()
	// Ensures background processes finishes
	client.Close()
}

func ExampleWriteApi_errors() {
	// Create client
	client := influxdb2.NewClient("http://localhost:9999", "my-token")
	// Get non-blocking write client
	writeApi := client.WriteApi("my-org", "my-bucket")
	// Get errors channel
	errorsCh := writeApi.Errors()
	// Create go proc for reading and logging errors
	go func() {
		for err := range errorsCh {
			fmt.Printf("write error: %s\n", err.Error())
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
		writeApi.WritePoint(p)
	}
	// Force all unwritten data to be sent
	writeApi.Flush()
	// Ensures background processes finishes
	client.Close()
}

func ExampleQueryApi_query() {
	// Create client
	client := influxdb2.NewClient("http://localhost:9999", "my-token")
	// Get query client
	queryApi := client.QueryApi("my-org")
	// get QueryTableResult
	result, err := queryApi.Query(context.Background(), `from(bucket:"my-bucket")|> range(start: -1h) |> filter(fn: (r) => r._measurement == "stat")`)
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

func ExampleQueryApi_queryRaw() {
	// Create client
	client := influxdb2.NewClient("http://localhost:9999", "my-token")
	// Get query client
	queryApi := client.QueryApi("my-org")
	// Query and get complete result as a string
	// Use default dialect
	result, err := queryApi.QueryRaw(context.Background(), `from(bucket:"my-bucket")|> range(start: -1h) |> filter(fn: (r) => r._measurement == "stat")`, api.DefaultDialect())
	if err == nil {
		fmt.Println("QueryResult:")
		fmt.Println(result)
	} else {
		panic(err)
	}
	// Ensures background processes finishes
	client.Close()
}

func ExampleOrganizationsApi() {
	// Create influxdb client
	client := influxdb2.NewClient("http://localhost:9999", "my-token")

	// Get Organizations API client
	orgApi := client.OrganizationsApi()

	// Create new organization
	org, err := orgApi.CreateOrganizationWithName(context.Background(), "org-2")
	if err != nil {
		panic(err)
	}

	orgDescription := "My second org "
	org.Description = &orgDescription

	org, err = orgApi.UpdateOrganization(context.Background(), org)
	if err != nil {
		panic(err)
	}

	// Find user to set owner
	user, err := client.UsersApi().FindUserByName(context.Background(), "user-01")
	if err != nil {
		panic(err)
	}

	// Add another owner (first owner is the one who create organization
	_, err = orgApi.AddOwner(context.Background(), org, user)
	if err != nil {
		panic(err)
	}

	// Create new user to add to org
	newUser, err := client.UsersApi().CreateUserWithName(context.Background(), "user-02")
	if err != nil {
		panic(err)
	}

	// Add new user to organization
	_, err = orgApi.AddMember(context.Background(), org, newUser)
	if err != nil {
		panic(err)
	}
	// Ensures background processes finishes
	client.Close()
}

func ExampleAuthorizationsApi() {
	// Create influxdb client
	client := influxdb2.NewClient("http://localhost:9999", "my-token")

	// Find user to grant permission
	user, err := client.UsersApi().FindUserByName(context.Background(), "user-01")
	if err != nil {
		panic(err)
	}

	// Find organization
	org, err := client.OrganizationsApi().FindOrganizationByName(context.Background(), "my-org")
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
		User:        &user.Name,
		UserID:      user.Id,
	}

	// grant permission and create token
	authCreated, err := client.AuthorizationsApi().CreateAuthorization(context.Background(), auth)
	if err != nil {
		panic(err)
	}
	// Use token
	fmt.Println("Token: ", *authCreated.Token)
	// Ensures background processes finishes
	client.Close()
}

func ExampleUsersApi() {
	// Create influxdb client
	client := influxdb2.NewClient("http://localhost:9999", "my-token")

	// Find organization
	org, err := client.OrganizationsApi().FindOrganizationByName(context.Background(), "my-org")
	if err != nil {
		panic(err)
	}

	// Get users API client
	usersApi := client.UsersApi()

	// Create new user
	user, err := usersApi.CreateUserWithName(context.Background(), "user-01")
	if err != nil {
		panic(err)
	}

	// Set user password
	err = usersApi.UpdateUserPassword(context.Background(), user, "pass-at-least-8-chars")
	if err != nil {
		panic(err)
	}

	// Add user to organization
	_, err = client.OrganizationsApi().AddMember(context.Background(), org, user)
	if err != nil {
		panic(err)
	}
	// Ensures background processes finishes
	client.Close()
}

func ExampleLabelsApi() {
	// Create influxdb client
	client := influxdb2.NewClient("http://localhost:9999", "my-token")

	ctx := context.Background()
	// Get Labels API client
	labelsApi := client.LabelsApi()
	// Get Organizations API client
	orgsApi := client.OrganizationsApi()

	// Get organization that will own label
	myorg, err := orgsApi.FindOrganizationByName(ctx, "my-org")
	if err != nil {
		panic(err)
	}

	labelName := "Active State"
	props := map[string]string{"color": "33ffdd", "description": "Marks org active"}
	label, err := labelsApi.CreateLabelWithName(ctx, myorg, labelName, props)
	if err != nil {
		panic(err)
	}

	// Get organization that will have the label
	org, err := orgsApi.FindOrganizationByName(ctx, "IT")
	if err != nil {
		panic(err)
	}

	// Add label to org
	_, err = orgsApi.AddLabel(ctx, org, label)
	if err != nil {
		panic(err)
	}

	// Change color property
	label.Properties.AdditionalProperties = map[string]string{"color": "ff1122"}
	label, err = labelsApi.UpdateLabel(ctx, label)
	if err != nil {
		panic(err)
	}

	// Close the client
	client.Close()
}
