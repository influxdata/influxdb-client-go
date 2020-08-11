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

func ExampleBucketsAPI() {
	// Create influxdb client
	client := influxdb2.NewClient("http://localhost:9999", "my-token")

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
	// Create client
	client := influxdb2.NewClient("http://localhost:9999", "my-token")
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
	// Create client
	client := influxdb2.NewClient("http://localhost:9999", "my-token")
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
	// Create client
	client := influxdb2.NewClient("http://localhost:9999", "my-token")
	// Get non-blocking write client
	writeAPI := client.WriteAPI("my-org", "my-bucket")
	// Get errors channel
	errorsCh := writeAPI.Errors()
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
		writeAPI.WritePoint(p)
	}
	// Force all unwritten data to be sent
	writeAPI.Flush()
	// Ensures background processes finishes
	client.Close()
}

func ExampleQueryAPI_query() {
	// Create client
	client := influxdb2.NewClient("http://localhost:9999", "my-token")
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

func ExampleQueryAPI_queryRaw() {
	// Create client
	client := influxdb2.NewClient("http://localhost:9999", "my-token")
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
	// Create influxdb client
	client := influxdb2.NewClient("http://localhost:9999", "my-token")

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
	// Create influxdb client
	client := influxdb2.NewClient("http://localhost:9999", "my-token")

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
		User:        &user.Name,
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
	// Create influxdb client
	client := influxdb2.NewClient("http://localhost:9999", "my-token")

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

func ExampleLabelsAPI() {
	// Create influxdb client
	client := influxdb2.NewClient("http://localhost:9999", "my-token")

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
