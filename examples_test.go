package influxdb2_test

import (
	"github.com/influxdata/influxdb-client-go"
)

func ExampleClient_newClient() {
	// Create client
	client := influxdb2.NewClient("http://localhost:9999", "my-token")

	// always close client at the end
	defer client.Close()
}

func ExampleClient_newClientWithOptions() {
	// Create client and set batch size to 20
	client := influxdb2.NewClientWithOptions("http://localhost:9999", "my-token",
		influxdb2.DefaultOptions().SetBatchSize(20))

	// always close client at the end
	defer client.Close()
}
