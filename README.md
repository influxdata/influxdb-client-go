# InfluxDB Client Go

[![CircleCI](https://circleci.com/gh/influxdata/influxdb-client-go/tree/v3.svg?style=svg)](https://circleci.com/gh/influxdata/influxdb-client-go/tree/v3)
[![codecov](https://codecov.io/gh/influxdata/influxdb-client-go/branch/v3/graph/badge.svg)](https://app.codecov.io/gh/influxdata/influxdb-client-go/branch/v3)
[![License](https://img.shields.io/github/license/influxdata/influxdb-client-go.svg)](https://github.com/influxdata/influxdb-client-go/blob/v3/LICENSE)
[![Slack Status](https://img.shields.io/badge/slack-join_chat-white.svg?logo=slack&style=social)](https://www.influxdata.com/slack)


InfluxDB 2 Client Go v3. 

This repository contains the reference Go client for InfluxDB 2.

#### Note: Use this client library with InfluxDB 2.x and InfluxDB 1.8+ ([see details](#influxdb-18-api-compatibility)). For connecting to InfluxDB 1.7 or earlier instances, use the [influxdb1-go](https://github.com/influxdata/influxdb1-client) client library.

- [Features](#features)
- [Documentation](#documentation)
    - [Examples](#examples)
- [How To Use](#how-to-use)
    - [Installation](#installation)
    - [Basic Example](#basic-example)
    - [Checking Server State](#checking-server-state)
- [InfluxDB 1.8 API compatibility](#influxdb-18-api-compatibility)
- [Contributing](#contributing)
- [License](#license)

## Features

- InfluxDB 2 client
    - Querying data
        - using the Flux language
    - Writing data using
        - [Line Protocol](https://docs.influxdata.com/influxdb/v2.0/reference/syntax/line-protocol/)
        - [Data Point](https://pkg.go.dev/github.com/influxdata/influxdb-client-go/v3/Point)
        - [Custom type](https://pkg.go.dev/github.com/influxdata/influxdb-client-go/v3/Client.WriteData)
        - Both [asynchronous](https://pkg.go.dev/github.com/influxdata/influxdb-client-go/v3/PointsWriter) or [synchronous](https://pkg.go.dev/github.com/influxdata/influxdb-client-go/v3/Client.WritePoints) ways
    - InfluxDB 2 API
        - setup, ready, health
        - authorizations, users, organizations
        - buckets, delete
        - ...

## Documentation

This section contains links to the client library documentation.

- [Product documentation](https://docs.influxdata.com/influxdb/v2.0/tools/client-libraries/), [Getting Started](#how-to-use)
- [Examples](#examples)
- [API Reference](https://pkg.go.dev/github.com/influxdata/influxdb-client-go/v3)
- [Changelog](CHANGELOG.md)

### Examples

Examples for basic writing and querying data are shown below in this document

There are also other examples in the API docs:
- [Examples](https://pkg.go.dev/github.com/influxdata/influxdb-client-go/v3?tab=doc#pkg-examples)

## How To Use

### Installation
**Go 1.17** or later is required.

#### Go mod project
1.  Add the latest version of the client package to your project dependencies (go.mod).
    ```sh
    go get github.com/influxdata/influxdb-client-go/v3
    ```
1. Add import `github.com/influxdata/influxdb-client-go/v3/influxclient` to your source code.
#### GOPATH project
    ```sh
    go get github.com/influxdata/influxdb-client-go
    ```
Note: To have _go get_ in the GOPATH mode, the environment variable `GO111MODULE` must have the `off` value.

### Basic Example
The following example demonstrates how to write data to InfluxDB 2 and read them back using the Flux language:
```go
package main

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

    "github.com/influxdata/influxdb-client-go/v3"
)

func main() {
    // Create a new client using an InfluxDB server base URL and an authentication token
	client, err := influxclient.New(influxclient.Params{
		ServerURL: "http://localhost:8086",
		AuthToken: "my-token",
		Organization: "my-org", // Organization is optional for InfluxDB Cloud
	})
	
	if err != nil {
		panic(err)
	}
	// Close client at the end
	defer client.Close()
   
    // Create point using full params constructor 
    p := influxclient.NewPoint("stat",
        map[string]string{"unit": "temperature"},
        map[string]interface{}{"avg": 24.5, "max": 45.0},
        time.Now())
    // write point synchronously 
	err = client.WritePoints(context.Background(), "my-bucket", p)
	if err != nil {
		panic(err)
	}
    // Create point using fluent style
    p = influxclient.NewPointWithMeasurement("stat").
        AddTag("unit", "temperature").
        AddField("avg", 23.2).
        AddField("max", 45.0).
		SetTimestamp(time.Now())
	// write point synchronously 
    err = client.WritePoints(context.Background(), "my-bucket", p)
	if err != nil {
		panic(err)
	}
	// Prepare custom type
	sensorData := struct {
		Table       string    `lp:"measurement"`
		Unit        string    `lp:"tag,unit"`
		Avg         float64   `lp:"field,avg"`
		Max         float64   `lp:"field,max"`
		Time        time.Time `lp:"timestamp"`
	}{"stat", "temperature", 22.3, 40.3, time.Now()}
	// Write point
	err = client.WriteData(context.Background(), "my-bucket", sensorData)
	if err != nil {
		panic(err)
	}
    // Or write directly line protocol
    line := fmt.Sprintf("stat,unit=temperature avg=%f,max=%f", 23.5, 45.0)
    err = client.Write(context.Background(), "my-bucket", []byte(line))
	if err != nil {
		panic(err)
	}
	
    // Query using the Flux language and get parser for the flux query result
    res, err := client.Query(context.Background(), `from(bucket:"my-bucket")|> range(start: -1m) |> filter(fn: (r) => r._measurement == "stat") |> filter(fn: (r) => r._field == "avg")`, nil)
	if err != nil {
		panic(err)
	}
	// Declare custom type for data
	val := &struct {
		Time   time.Time `csv:"_time"`
		Avg    float64   `csv:"_value"`
		Unit   string    `csv:"unit"`
	}{}

	tw := tabwriter.NewWriter(os.Stdout, 10, 4, 2, ' ', 0)
	fmt.Fprintf(tw, "Time\tAvg\tUnit\n")

	// Iterate over result set
	for res.NextSection() {
		for res.NextRow() {
			err = res.Decode(val)
			if err != nil {
				fmt.Fprintf(tw, "%v\n", err)
				continue
			}
			fmt.Fprintf(tw, "%s\t%.2f\t%s\n", val.Time.String(), val.Avg, val.Unit)
		}
	}
	tw.Flush()
	if res.Err() != nil {
		panic(res.Err())
	}
    // Ensures background processes finishes
    client.Close()
}
```
### Options
The InfluxDBClient uses set of options to configure behavior. These are available in the [Params](https://pkg.go.dev/github.com/influxdata/influxdb-client-go/v3#Params) object
Creating a client instance using simple
```go
client, err := influxclient.New(influxclient.Params{
    ServerURL: "http://localhost:8086",
    AuthToken: "my-token",
    Organization: "my-org",
    })
```
will use the default options.

To set different configuration values, e.g. custom batch size and `WriteFailed` callback for `PointsWriter` , get default options
and change what is needed:
```go

wp := influxclient.DefaultWriteParams
// Set batch size to write 100 points in 2 batches
wp.BatchSize = 50
// Set callback for failed writes
wp.WriteFailed = func(err error, lines []byte, attempt int, expires time.Time) bool {
    fmt.Println("Write failed", err)
    return true
}
// Create client with custom WriteParams
client, err := influxclient.New(influxclient.Params{
    ServerURL: "https://eu-central-1-1.aws.cloud2.influxdata.com/",
    AuthToken: "my-token",
    //		Organization: "my-org", // Organization is optional for InfluxDB Cloud
    WriteParams: wp,
})
```

### Checking Server State
There are three functions for checking whether a server is up and ready for communication:

| Function                                                                                 | Description | Availability |
|:-----------------------------------------------------------------------------------------|:----------|:----------|
| [Health()](https://pkg.go.dev/github.com/influxdata/influxdb-client-go/v3#Client.Health) | Detailed info about the server status, along with version string | OSS |
| [Ready()](https://pkg.go.dev/github.com/influxdata/influxdb-client-go/v3#Client.Ready)   | Server uptime info | OSS |
| [Ping()](https://pkg.go.dev/github.com/influxdata/influxdb-client-go/v3#Client.Ping)     | Whether a server is up | OSS, Cloud |

Only the [Ping()](https://pkg.go.dev/github.com/influxdata/influxdb-client-go/v2#Client.Ping) function works in InfluxDB Cloud server.

## InfluxDB 1.8 API compatibility

[InfluxDB 1.8.0 introduced forward compatibility APIs](https://docs.influxdata.com/influxdb/latest/tools/api/#influxdb-2-0-api-compatibility-endpoints) for InfluxDB 2.0. This allow you to easily move from InfluxDB 1.x to InfluxDB 2.0 Cloud or open source.

Client API usage differences summary:
1. Use the form `username:password` for an **authentication token**. Example: `my-user:my-password`. Use an empty string (`""`) if the server doesn't require authentication.
1. The organization parameter is not used. Use an empty string (`""`) where necessary.
1. Use the form `database/retention-policy` where a **bucket** is required. Skip retention policy if the default retention policy should be used. Examples: `telegraf/autogen`, `telegraf`.  
  
  The following forward compatible APIs are available:

| API                                                                                                                                                                                       | Endpoint | Description |
--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|:----------|:----------|:----------|
| [Write()](https://pkg.go.dev/github.com/influxdata/influxdb-client-go/v3/Client.Write) (also [PointsWriter](https://pkg.go.dev/github.com/influxdata/influxdb-client-go/v3/PointsWriter)) | [/api/v2/write](https://docs.influxdata.com/influxdb/v2.0/write-data/developer-tools/api/) | Write data to InfluxDB 1.8.0+ using the InfluxDB 2.0 API |
| [Query()](https://pkg.go.dev/github.com/influxdata/influxdb-client-go/v3/Client.Query)                                                                                                    | [/api/v2/query](https://docs.influxdata.com/influxdb/v2.0/query-data/execute-queries/influx-api/) | Query data in InfluxDB 1.8.0+ using the InfluxDB 2.0 API and [Flux](https://docs.influxdata.com/flux/latest/) endpoint should be enabled by the [`flux-enabled` option](https://docs.influxdata.com/influxdb/v1.8/administration/config/#flux-enabled-false)
| [Health()](https://pkg.go.dev/github.com/influxdata/influxdb-client-go/v3#Client.Health)                                                                                                  | [/health](https://docs.influxdata.com/influxdb/v2.0/api/#tag/Health) | Check the health of your InfluxDB instance |    


### Example
```go
package main

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/influxdata/influxdb-client-go/v3/influxclient"
)

func main() {
    userName := "my-user"
    password := "my-password"
     // Create a new client using an InfluxDB server base URL and an authentication token
    // For authentication token supply a string in the form: "username:password" as a token. 
	// Set empty value for an unauthenticated server
	client, err := influxclient.New(influxclient.Params{
		ServerURL: "http://localhost:8086",
		AuthToken: fmt.Sprintf("%s:%s",userName, password),
	})

	if err != nil {
		panic(err)
	}
	// Close client at the end
	defer client.Close()

	// Create point using full params constructor 
	p := influxclient.NewPoint("stat",
		map[string]string{"unit": "temperature"},
		map[string]interface{}{"avg": 24.5, "max": 45.0},
		time.Now())
	// write point synchronously 
	// Supply a string in the form database/retention-policy as a bucket. Skip retention policy for the default one, use just a database name (without the slash character)
	err = client.WritePoints(context.Background(), "test/autogen", p)
	if err != nil {
		panic(err)
	}
	
	// Query using the Flux language and get parser for the flux query result
	res, err := client.Query(context.Background(), `from(bucket:"test/autogen")|> range(start: -1m) |> filter(fn: (r) => r._measurement == "stat") |> filter(fn: (r) => r._field == "avg")`, nil)
	if err != nil {
		panic(err)
	}
	// Declare custom type for data
	val := &struct {
		Time   time.Time `csv:"_time"`
		Avg    float64   `csv:"_value"`
		Unit   string    `csv:"unit"`
	}{}

	tw := tabwriter.NewWriter(os.Stdout, 10, 4, 2, ' ', 0)
	fmt.Fprintf(tw, "Time\tAvg\tUnit\n")

	// Iterate over result set
	for res.NextSection() {
		for res.NextRow() {
			err = res.Decode(val)
			if err != nil {
				fmt.Fprintf(tw, "%v\n", err)
				continue
			}
			fmt.Fprintf(tw, "%s\t%.2f\t%s\n", val.Time.String(), val.Avg, val.Unit)
		}
	}
	tw.Flush()
	if res.Err() != nil {
		panic(res.Err())
	}
	// Ensures background processes finishes
	client.Close()
}
```

## Contributing

If you would like to contribute code you can do through GitHub by forking the repository and sending a pull request into the `master` branch.

## License

The InfluxDB 2 Go Client is released under the [MIT License](https://opensource.org/licenses/MIT).
