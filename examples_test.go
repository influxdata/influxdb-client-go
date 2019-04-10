package client_test

import (
	"compress/gzip"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"time"

	client "github.com/influxdata/influxdb-client-go"
)

var myHTTPClient, myInfluxAddress = func() (*http.Client, string) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reader := r.Body
		if r.Header.Get("Content-Encoding") == "gzip" {
			var err error
			reader, err = gzip.NewReader(reader)
			if err != nil {
				fmt.Println(err)
			}
		}
		buf, err := ioutil.ReadAll(reader)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(string(buf))
		w.WriteHeader(200)
	}))

	return server.Client(), server.URL
}()

func ExampleClient_Write() {
	influx, err := client.NewClient(myHTTPClient, client.WithAddress(myInfluxAddress), client.WithToken("mytoken"))
	if err != nil {
		fmt.Print("foo") // error handling here, normally we wouldn't use fmt, but it works for the example
	}

	// we use client.NewRowMetric for the example because its easy, but if you need extra performance
	// it is fine to manually build the []client.Metric{}
	myMetrics := []client.Metric{
		client.NewRowMetric(
			map[string]interface{}{"memory": 1000, "cpu": 0.93},
			"system-metrics",
			map[string]string{"hostname": "hal9000"},
			time.Date(2018, 3, 4, 5, 6, 7, 8, time.UTC)),
		client.NewRowMetric(
			map[string]interface{}{"memory": 1000, "cpu": 0.93},
			"system-metrics",
			map[string]string{"hostname": "hal9000"},
			time.Date(2018, 3, 4, 5, 6, 7, 9, time.UTC)),
	}

	// The actual write...
	if err := influx.Write(context.Background(), "my-awesome-bucket", "my-very-awesome-org", myMetrics...); err != nil {
		fmt.Println(err) // as above use your own error handling here.
	}
	influx.Close() // closes the client.  After this the client is useless.
	// Output:
	// system-metrics,hostname=hal9000 cpu=0.93,memory=1000i 1520139967000000008
	// system-metrics,hostname=hal9000 cpu=0.93,memory=1000i 1520139967000000009
}
