package influxdb2_test

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/influxdata/influxdb-client-go"
)

func ExampleWriteApiBlocking() {
	// Create client
	client := influxdb2.NewClient("http://localhost:9999", "my-token")
	// Get blocking write client
	writeApi := client.WriteApiBlocking("my-org", "my-bucket")
	// write some points
	for i := 0; i < 100; i++ {
		// create data point
		p := influxdb2.NewPoint(
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
	// Create client and set batch size to 20
	client := influxdb2.NewClientWithOptions("http://localhost:9999", "my-token",
		influxdb2.DefaultOptions().SetBatchSize(20))
	// Get non-blocking write client
	writeApi := client.WriteApi("my-org", "my-bucket")
	// write some points
	for i := 0; i < 100; i++ {
		// create point
		p := influxdb2.NewPoint(
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
	result, err := queryApi.QueryRaw(context.Background(), `from(bucket:"my-bucket")|> range(start: -1h) |> filter(fn: (r) => r._measurement == "stat")`, influxdb2.DefaultDialect())
	if err == nil {
		fmt.Println("QueryResult:")
		fmt.Println(result)
	} else {
		panic(err)
	}
	// Ensures background processes finishes
	client.Close()
}
