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
    - [Writes in Detail](#writes)
    - [Queries in Detail](#queries)
    - [Parametrized Queries](#parametrized-queries)
    - [Concurrency](#concurrency)
    - [Proxy and redirects](#proxy-and-redirects)
    - [Checking Server State](#checking-server-state)
- [InfluxDB 1.8 API compatibility](#influxdb-18-api-compatibility)
- [Contributing](#contributing)
- [License](#license)

## Features

- InfluxDB 2 client
    - Querying data
        - using the Flux language
        - into custom annotated type, map or raw data 
        - [How to queries](#queries)
    - Writing data using
        - [Point](#point)
        - [Custom type](#custom-type)
        - [Line Protocol](#line-protocol)
        - Both [asynchronous](#asynchronous-write) ways and [synchronous](#synchronous-write) ways
        - [How to writes](#writes)
    - InfluxDB 2 API
        - ready, health, ping
        - buckets, authorizations, users, organizations
        - delete, tasks, labels
        - ...

## Documentation

This section contains links to the client library documentation.

- [Product documentation](https://docs.influxdata.com/influxdb/v2.0/tools/client-libraries/), [Getting Started](#how-to-use)
- [Examples](#examples)
- [API Reference](https://pkg.go.dev/github.com/influxdata/influxdb-client-go/v3/influxclient)
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
    go get github.com/influxdata/influxdb-client-go/v3/influxclient
    ```
1. Add import `github.com/influxdata/influxdb-client-go/v3/influxclient` to your source code.
#### GOPATH project
    It is strongly recommended using [Go modules](https://go.dev/blog/using-go-modules). 

    ```sh
    go get github.com/influxdata/influxdb-client-go/influxclient
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
The InfluxDBClient uses set of options to configure behavior. These are available in the [Params](https://pkg.go.dev/github.com/influxdata/influxdb-client-go/v3/influxclient/#Params) object
Creating a client instance using simple
```go
    client, err := influxclient.New(influxclient.Params{
        ServerURL: "http://localhost:8086",
        AuthToken: "my-token",
        Organization: "my-org",
        })
```
will use the default options.

To set different configuration values, e.g. custom batch size and `WriteFailed` callback for `PointsWriter`, get default options
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

### Writes

#### Passing data for writing
Client offers several ways how to provide data for writing:

#### Point
[Point](https://pkg.go.dev/github.com/influxdata/influxdb-client-go/v3/influxclient/#Point) is basic object representation of [InfluxDB line Protocol](https://docs.influxdata.com/influxdb/v2.0/reference/syntax/line-protocol/).
Offers [fluent style](https://en.wikipedia.org/wiki/Fluent_interface) API for short line setting data.

It can be created and filled various ways:
```go
    // Create point using field setting
    p := influxclient.Point{
        Measurement: "stat",
        Timestamp:   time.Now(),
    }
    p.AddTag("unit", "temperature").AddField("avg", 23.2).AddField("max", 45.0)
        
    // Create point using full params constructor
    p = influxclient.NewPoint("stat",
        map[string]string{"unit": "temperature"},
        map[string]interface{}{"avg": 24.5, "max": 45.0},
        time.Now())
    
    // Create point using fluent style
    p = influxclient.NewPointWithMeasurement("stat").
        AddTag("unit", "temperature").
        AddField("avg", 23.2).
        AddField("max", 45.0).
        SetTimestamp(time.Now())	
```

#### Custom type
Write data using a custom annotated type allows smooth integration with an application model.
Client encodes fields of custom type into line protocol. Each custom type must have fields annotated with 
'lp' prefix, value for line protocol part and optional custom name. If custom name is not provided the struct 
fieldname is used:
 - `lp:"measurement"` - mandatory, specifies name of the measurement where to store data
 - `lp:"tag,<tag-name>` - zero or more struct fields of string type. 
 - `lp:"field,<field-name>"` - one or more struct fields of any primitive type
 - `lp:"timestamp"` - zero or one struct field for of `time.Time` type for timestamp
 - `lp:"-"` - use this for skipping field of encoding

```go
    sensorData := struct {
        Table       string    `lp:"measurement"`
        Sensor      string    `lp:"tag,sensor"`
        ID          string    `lp:"tag,device_id"`
        Temperature float64   `lp:"field,temperature"`
        Humidity    int       `lp:"field,humidity"`
        Time        time.Time `lp:"timestamp"`
    }{"air", "SHT31", "1012", 22.3, 55, time.Now()}
```
#### Line protocol
If data are already in the form of [InfluxDB line Protocol](https://docs.influxdata.com/influxdb/v2.0/reference/syntax/line-protocol/) 
it can be directly used with `Write` functions.  

### How to write data
Client has two ways of writing, asynchronous (non-blocking) and synchronous (blocking). Both ways allow writing of [Point](#point), [custom type](#custom-type) or [InfluxDB line protocol](#line-protocol)

### Synchronous write
[Client](https://pkg.go.dev/github.com/influxdata/influxdb-client-go/v3/influxclient/#Client) object offers functions for writing data synchronously. 
It doesn't do implicit batching or retrying. Caller has full control of the result. Batch is created in case of more points are passed.

```go
package main

import (
    "context"
    "fmt"
    "math/rand"
    "time"

    "github.com/influxdata/influxdb-client-go/v3/influxclient"
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
    
    // write some points
    for i := 0; i <100; i++ {
        // create data point
        p := influxclient.NewPoint(
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
        err := client.WritePoints(context.Background(), "my-bucket", p)
        if err != nil {
            panic(err)
        }
    }
    // Ensures background processes finishes
    client.Close()
}
```
### Asynchronous write
Asynchronous (non-blocking) [PointsWriter](https://pkg.go.dev/github.com/influxdata/influxdb-client-go/v3/influxclient/#PointsWriter) uses implicit batching and retrying mechanism. Data are asynchronously
written to the underlying buffer, and they are automatically sent to an InfluxDB server when a condition is met: 
 - the number of points in the buffer is equal `BatchSize`, default 5000
 - the number of bytes represented by points in the buffer reaches `MaxBatchBytes`
 - the `FlushInterval`, default 1s, times out.

Write buffer can also be flushed manually using [Flush()](https://pkg.go.dev/github.com/influxdata/influxdb-client-go/v2/api#WriteAPI.Flush) method.

Write options are configured in [WriteParams](https://pkg.go.dev/github.com/influxdata/influxdb-client-go/v3/influxclient/#WriteParams). 
 
Writes are automatically retried on server back pressure or a connection error.

Note: Always use [Close()](https://pkg.go.dev/github.com/influxdata/influxdb-client-go/v3/influxclient/#PointsWriter.Close) method of the writer to stop all background processes.

Asynchronous writer is recommended for frequent periodic writes.

```go
package main

import (
    "fmt"
    "math/rand"
    "time"

    "github.com/influxdata/influxdb-client-go/v3/influxclient"
)

func main() {
    wp := influxclient.DefaultWriteParams
    // Set batch size to write 100 points in 2 batches
    wp.BatchSize = 50
    // Create client with custom WriteParams
    client, err := influxclient.New(influxclient.Params{
        ServerURL: "http://localhost:8086",
        AuthToken: "my-token",
        Organization: "my-org", // Organization is optional for InfluxDB Cloud
        WriteParams: wp,
    })
    if err != nil {
        panic(err)
    }
    // client.Close() have to be called to clean http connections
    defer client.Close()
    // Get async writer
    writer := client.PointsWriter("my-bucket")
    // writer.Close() MUST be called at the end to ensure completing background operations and cleaning resources
    defer writer.Close()

    // write some points
    for i := 0; i < 100; i++ {
        // create point
        p := influxclient.NewPointWithMeasurement("stat").
            AddTag("id", fmt.Sprintf("rack_%v", i%10)).
            AddTag("vendor", "AWS").
            AddTag("hostname", fmt.Sprintf("host_%v", i%100)).
            AddField("temperature", rand.Float64()*80.0).
            AddField("disk_free", rand.Float64()*1000.0).
            AddField("disk_total", (i/10+1)*1000000).
            AddField("mem_total", (i/100+1)*10000000).
            AddField("mem_free", rand.Uint64()).
            SetTimestamp(time.Now())
        // write asynchronously
        writer.WritePoints(p)
    }
}
```
### Handling of failed async writes
[PointsWriter](https://pkg.go.dev/github.com/influxdata/influxdb-client-go/v3/influxclient/#PointsWriter) continues with retrying of failed writes by default.
Retried are automatically writes that fail on a connection failure or when server returns response HTTP status code >= 429.

Retrying algorithm uses random exponential strategy to set retry time by default.
The delay for the next retry attempt is a random value in the interval _retryInterval * exponentialBase^(attempts)_ and _retryInterval * exponentialBase^(attempts+1)_.
If writes of batch repeatedly fails, WriteAPI continues with retrying until _maxRetries_ is reached or the overall retry time of batch exceeds _maxRetryTime_.


The defaults parameters (part of the WriteOptions) are:
- _retryInterval_=5,000ms
- _exponentialBase_=2
- _maxRetryDelay_=125,000ms
- _maxRetries_=5
- _maxRetryTime_=180,000ms

Retry delays are by default randomly distributed within the ranges:
1. 5,000-10,000
1. 10,000-20,000
1. 20,000-40,000
1. 40,000-80,000
1. 80,000-125,000

#### WriteFailed callback
[WriteFailedCallback](https://pkg.go.dev/github.com/influxdata/influxdb-client-go/v3/influxclient/#WriteParams#WriteFailed) allows advanced controlling of retrying.
Callback synchronously notified when async write fails.
It controls further batch handling by its return value. If it returns `true`, writer continues with retrying of writes of this batch. Returned `false` means the batch should be discarded.

This callback is also called when a point encoding fails or the maximum retry time is reached. Returning value is omitted in these cases.

#### Disabling retry mechanism
Setting _RetryInterval_ to 0 disables retry strategy and any failed write will discard the batch (after calling [WriteFailed](https://pkg.go.dev/github.com/influxdata/influxdb-client-go/v3/influxclient/#WriteParams#WriteFailed) callback.

#### Removed old items 
When RetryBuffer reaches its limit it drops the oldest batches stored for retrying. When this happens, [WriteRetrySkipped](https://pkg.go.dev/github.com/influxdata/influxdb-client-go/v3/influxclient/#WriteParams#WriteRetrySkipped) callback
is called to inform about removed data.

### Queries
[Client.Query()](https://pkg.go.dev/github.com/influxdata/influxdb-client-go/v3/influxclient/#Client.Query) sends the given Flux query to server and returns [QueryResultReader](https://pkg.go.dev/github.com/influxdata/influxdb-client-go/v3/influxclient/#QueryResultReader) for further parsing result.

Use [QueryResultReader.NextSection()](https://pkg.go.dev/github.com/influxdata/influxdb-client-go/v3/influxclient/#QueryResultReader.NextSection) for navigation to the sections in the query result set.
Use [QueryResultReader.NextRow()](https://pkg.go.dev/github.com/influxdata/influxdb-client-go/v3/influxclient/#QueryResultReader.NextRow) for iterating over rows in the section.
Read the row raw data using [QueryResultReader.Row()](https://pkg.go.dev/github.com/influxdata/influxdb-client-go/v3/influxclient/#QueryResultReader.Row)
 or deserialize data into a struct or a slice via [QueryResultReader.Decode](https://pkg.go.dev/github.com/influxdata/influxdb-client-go/v3/annotatedcsv/#Reader.Decode):


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
    // // Create a new client using an InfluxDB server base URL and an authentication token
    client, err := influxclient.New(influxclient.Params{
        ServerURL:    "http://localhost:8086/",
        AuthToken:    "my-token",
        Organization: "my-org",
    })
    if err != nil {
        panic(err)
    }
    //	Build query
    query := `from(bucket: "my-bucket") 
        |> range(start: -20m) 
        |> filter(fn: (r) => r._measurement == "stat")
        |> filter(fn: (r) => r._field == "disk_free")`

    res, err := client.Query(context.Background(), query, nil)
    if err != nil {
        panic(err)
    }

    defer res.Close()

    val := &struct {
        Time     time.Time `csv:"_time"`
        Hostname string    `csv:"hostname"`
        DiskFree float64   `csv:"_value"`
    }{}

    tw := tabwriter.NewWriter(os.Stdout, 10, 4, 2, ' ', 0)
    fmt.Fprintf(tw, "Time\tHostanme\tDisk free\n")

    for res.NextRow() {
        err = res.Decode(val)
        if err != nil {
            fmt.Fprintf(tw, "%v\n", err)
            continue
        }
        fmt.Fprintf(tw, "%s\t%s\t%.1f\n", val.Time.String(), val.Hostname, val.DiskFree)
    }
    tw.Flush()
    if res.Err() != nil {
        panic(res.Err())
    }
}
```

### Parametrized Queries
InfluxDB Cloud supports [Parameterized Queries](https://docs.influxdata.com/influxdb/cloud/query-data/parameterized-queries/)
that let you dynamically change values in a query using the InfluxDB API. Parameterized queries make Flux queries more
reusable and can also be used to help prevent injection attacks.

InfluxDB Cloud inserts the params object into the Flux query as a Flux record named `params`. Use dot or bracket
notation to access parameters in the `params` record in your Flux query. Parameterized Flux queries support only `int`
, `float`, and `string` data types. To convert the supported data types into
other [Flux basic data types, use Flux type conversion functions](https://docs.influxdata.com/influxdb/cloud/query-data/parameterized-queries/#supported-parameter-data-types).

Query parameters can be passed as a struct or map. Param values can be only simple types or `time.Time`.
The name of the parameter represented by a struct field can be specified by JSON annotation.

Parameterized query example:
> :warning: Parameterized Queries are supported only in InfluxDB Cloud. There is no support in InfluxDB OSS currently.
```go
package main

import (
    "context"
    "fmt"

    "github.com/influxdata/influxdb-client-go/v2"
)

func main() {
    // // Create a new client using an InfluxDB Cloud URL and an authentication token
    client, err := influxclient.New(influxclient.Params{
        ServerURL: "https://eu-central-1-1.aws.cloud2.influxdata.com/",
        AuthToken: "my-token",
    })
    if err != nil {
        panic(err)
    }
    // Define query parameters
    params := struct {
        Since string `json:"since"`
        Field string `json:"field"`
    }{
        "-20m",
        "disk_free",
    }
    //	Build query with params
    query := `from(bucket: "my-bucket") 
        |> range(start: duration(v: params.since)) 
        |> filter(fn: (r) => r._measurement == "stat")
        |> filter(fn: (r) => r._field == params.field)`

    res, err := client.Query(context.Background(), query, params)
    if err != nil {
        panic(err)
    }

    defer res.Close()

    val := &struct {
        Time     time.Time `csv:"_time"`
        Hostname string    `csv:"hostname"`
        DiskFree float64   `csv:"_value"`
    }{}

    tw := tabwriter.NewWriter(os.Stdout, 10, 4, 2, ' ', 0)
    fmt.Fprintf(tw, "Time\tHostanme\tDisk free\n")

    for res.NextRow() {
        err = res.Decode(val)
        if err != nil {
            fmt.Fprintf(tw, "%v\n", err)
            continue
        }
        fmt.Fprintf(tw, "%s\t%s\t%.1f\n", val.Time.String(), val.Hostname, val.DiskFree)
    }
    tw.Flush()
    if res.Err() != nil {
        panic(res.Err())
    }
}
```

### Concurrency
InfluxDB Go Client can be used in a concurrent environment. All its functions are thread-safe.

The best practise is to use a single `Client` instance per server URL. This ensures optimized resources usage,
most importantly reusing HTTP connections.

For efficient reuse of HTTP resources among multiple clients, create an HTTP client and set it to all clients:
```go
    // Create HTTP client
    httpClient := &http.Client{
        Timeout: time.Second * time.Duration(60),
        Transport: &http.Transport{
            DialContext: (&net.Dialer{
                Timeout: 5 * time.Second,
            }).DialContext,
            TLSHandshakeTimeout: 5 * time.Second,
            TLSClientConfig: &tls.Config{
                InsecureSkipVerify: true,
            },
            MaxIdleConns:        100,
            MaxIdleConnsPerHost: 100,
            IdleConnTimeout:     90 * time.Second,
        },
    }
    // Client for server 1
    client1, err := influxclient.New(influxclient.Params{
        ServerURL:  "http://server1/",
        AuthToken:  "token1",
        HTTPClient: httpClient,
        })
    if err != nil {
        panic(err)
    }
    // Client for server 2
    client2, err := influxclient.New(influxclient.Params{
        ServerURL:  "http://server2/",
        AuthToken:  "token2",
        HTTPClient: httpClient,
        })
    if err != nil {
        panic(err)
    }
```


## Proxy and redirects
You can configure InfluxDB Go client behind a proxy in two ways:
1. Using environment variable  
   Set environment variable `HTTP_PROXY` (or `HTTPS_PROXY` based on the scheme of your server url).  
   e.g. (linux) `export HTTP_PROXY=http://my-proxy:8080` or in Go code `os.Setenv("HTTP_PROXY","http://my-proxy:8080")`

1. Configure `http.Client` to use proxy<br>
   Create a custom `http.Client` with a proxy configuration:
```go
   proxyUrl, err := url.Parse("http://my-proxy:8080")
   httpClient := &http.Client{
       Transport: &http.Transport{
           Proxy: http.ProxyURL(proxyUrl)
       }
   }
   // Client with custom HTTPClient
   client, err := influxclient.New(influxclient.Params{
        ServerURL:  "http://server/",
        AuthToken:  "token",
        HTTPClient: httpClient,
        })
    if err != nil {
        panic(err)
    }
```

Client automatically follows HTTP redirects. The default redirect policy is to follow up to 10 consecutive requests.
Due to a security reason _Authorization_ header is not forwarded when redirect leads to a different domain.
To overcome this limitation you have to set a custom redirect handler:
```go
    token := "my-token"
    
    httpClient := &http.Client{
        CheckRedirect: func(req *http.Request, via []*http.Request) error {
            req.Header.Add("Authorization","Token " + token)
            return nil
        },
    }
    // Client with custom HTTPClient
    client, err := influxclient.New(influxclient.Params{
        ServerURL:  "http://server/",
        AuthToken:  token,
        HTTPClient: httpClient,
    })
    if err != nil {
        panic(err)
    }
``` 

### Checking Server State
There are three functions for checking whether a server is up and ready for communication:

| Function                                                                                 | Description | Availability |
|:-----------------------------------------------------------------------------------------|:----------|:----------|
| [Health()](https://pkg.go.dev/github.com/influxdata/influxdb-client-go/v3/influxclient/#Client.Health) | Detailed info about the server status, along with version string | OSS |
| [Ready()](https://pkg.go.dev/github.com/influxdata/influxdb-client-go/v3/influxclient/#Client.Ready)   | Server uptime info | OSS |
| [Ping()](https://pkg.go.dev/github.com/influxdata/influxdb-client-go/v3/influxclient/#Client.Ping)     | Whether a server is up | OSS, Cloud |

Only the [Ping()](https://pkg.go.dev/github.com/influxdata/influxdb-client-go/v3/influxclient/#Client.Ping) function works in InfluxDB Cloud server.

## InfluxDB 1.8 API compatibility

[InfluxDB 1.8.0 introduced forward compatibility APIs](https://docs.influxdata.com/influxdb/latest/tools/api/#influxdb-2-0-api-compatibility-endpoints) for InfluxDB 2.0. This allow you to easily move from InfluxDB 1.x to InfluxDB 2.0 Cloud or open source.

Client API usage differences summary:
1. Use the form `username:password` for an **authentication token**. Example: `my-user:my-password`. Use an empty string (`""`) if the server doesn't require authentication.
1. The organization parameter is not used. Use an empty string (`""`) where necessary.
1. Use the form `database/retention-policy` where a **bucket** is required. Skip retention policy if the default retention policy should be used. Examples: `telegraf/autogen`, `telegraf`.  
  
  The following forward compatible APIs are available:

| API                                                                                                                                                                                                                   | Endpoint | Description |
|:----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|:----------|:----------|
| [Write()](https://pkg.go.dev/github.com/influxdata/influxdb-client-go/v3/influxclient/#Client.Write) (also [PointsWriter](https://pkg.go.dev/github.com/influxdata/influxdb-client-go/v3/influxclient/#PointsWriter)) | [/api/v2/write](https://docs.influxdata.com/influxdb/v2.0/write-data/developer-tools/api/) | Write data to InfluxDB 1.8.0+ using the InfluxDB 2.0 API |
| [Query()](https://pkg.go.dev/github.com/influxdata/influxdb-client-go/v3/influxclient/#Client.Query)                                                                                                                  | [/api/v2/query](https://docs.influxdata.com/influxdb/v2.0/query-data/execute-queries/influx-api/) | Query data in InfluxDB 1.8.0+ using the InfluxDB 2.0 API and [Flux](https://docs.influxdata.com/flux/latest/) endpoint should be enabled by the [`flux-enabled` option](https://docs.influxdata.com/influxdb/v1.8/administration/config/#flux-enabled-false)
| [Health()](https://pkg.go.dev/github.com/influxdata/influxdb-client-go/v3/influxclient/#Client.Health)                                                                                                                | [/health](https://docs.influxdata.com/influxdb/v2.0/api/#tag/Health) | Check the health of your InfluxDB instance |    


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
