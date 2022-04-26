package influxclient_test

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/influxdata/influxdb-client-go/influxclient"
)

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
