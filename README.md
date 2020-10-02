# InfluxDB Client Go

[![CircleCI](https://circleci.com/gh/influxdata/influxdb-client-go.svg?style=svg)](https://circleci.com/gh/influxdata/influxdb-client-go)
[![codecov](https://codecov.io/gh/influxdata/influxdb-client-go/branch/master/graph/badge.svg)](https://codecov.io/gh/influxdata/influxdb-client-go)
[![License](https://img.shields.io/github/license/influxdata/influxdb-client-go.svg)](https://github.com/influxdata/influxdb-client-go/blob/master/LICENSE)
[![Slack Status](https://img.shields.io/badge/slack-join_chat-white.svg?logo=slack&style=social)](https://www.influxdata.com/slack)

This repository contains the reference Go client for InfluxDB 2.

#### Note: Use this client library with InfluxDB 2.x and InfluxDB 1.8+ ([see details](#influxdb-18-api-compatibility)). For connecting to InfluxDB 1.7 or earlier instances, use the [influxdb1-go](https://github.com/influxdata/influxdb1-client) client library.

- [Features](#features)
- [Documentation](#documentation)
    - [Examples](#examples)
- [How To Use](#how-to-use)
    - [Basic Example](#basic-example)
    - [Writes in Detail](#writes)
    - [Queries in Detail](#queries)
- [InfluxDB 1.8 API compatibility](#influxdb-18-api-compatibility)
- [Contributing](#contributing)
- [License](#license)

## Features

- InfluxDB 2 client 
    - Querying data 
        - using the Flux language
        - into raw data, flux table representation
        - [How to queries](#queries)
    - Writing data using
        - [Line Protocol](https://docs.influxdata.com/influxdb/v1.6/write_protocols/line_protocol_tutorial/) 
        - [Data Point](https://github.com/influxdata/influxdb-client-go/blob/master/point.go)
        - Both [asynchronous](https://github.com/influxdata/influxdb-client-go/blob/master/write.go) or [synchronous](https://github.com/influxdata/influxdb-client-go/blob/master/writeAPIBlocking.go) ways
        - [How to writes](#writes)  
    - InfluxDB 2 API
        - setup, ready, health
        - authotizations, users, organizations
        - buckets, delete
        - ...
     
## Documentation

Go API docs is available at: [https://pkg.go.dev/github.com/influxdata/influxdb-client-go](https://pkg.go.dev/github.com/influxdata/influxdb-client-go)

### Examples

Examples for basic writing and querying data are shown below in this document 

There are also other examples in the API docs:
 - [Client usage](https://pkg.go.dev/github.com/influxdata/influxdb-client-go?tab=doc#pkg-examples) 
 - [Management APIs](https://pkg.go.dev/github.com/influxdata/influxdb-client-go@v1.4.0/api?tab=doc#pkg-examples) 

## How To Use

### Installation
**Go 1.13** or later is required.

1.  Add the client package your to your project dependencies (go.mod).
    ```sh
    go get github.com/influxdata/influxdb-client-go/v2
    ```
1. Add import `github.com/influxdata/influxdb-client-go/v2` to your source code.

### Basic Example
The following example demonstrates how to write data to InfluxDB 2 and read them back using the Flux language:
```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/influxdata/influxdb-client-go/v2"
)

func main() {
    // Create a new client using an InfluxDB server base URL and an authentication token
    client := influxdb2.NewClient("http://localhost:8086", "my-token")
    // Use blocking write client for writes to desired bucket
    writeAPI := client.WriteAPIBlocking("my-org", "my-bucket")
    // Create point using full params constructor 
    p := influxdb2.NewPoint("stat",
        map[string]string{"unit": "temperature"},
        map[string]interface{}{"avg": 24.5, "max": 45.0},
        time.Now())
    // write point immediately 
    writeAPI.WritePoint(context.Background(), p)
    // Create point using fluent style
    p = influxdb2.NewPointWithMeasurement("stat").
        AddTag("unit", "temperature").
        AddField("avg", 23.2).
        AddField("max", 45.0).
        SetTime(time.Now())
    writeAPI.WritePoint(context.Background(), p)
    
    // Or write directly line protocol
    line := fmt.Sprintf("stat,unit=temperature avg=%f,max=%f", 23.5, 45.0)
    writeAPI.WriteRecord(context.Background(), line)

    // Get query client
    queryAPI := client.QueryAPI("my-org")
    // Get parser flux query result
    result, err := queryAPI.Query(context.Background(), `from(bucket:"my-bucket")|> range(start: -1h) |> filter(fn: (r) => r._measurement == "stat")`)
    if err == nil {
        // Use Next() to iterate over query result lines
        for result.Next() {
            // Observe when there is new grouping key producing new table
            if result.TableChanged() {
                fmt.Printf("table: %s\n", result.TableMetadata().String())
            }
            // read result
            fmt.Printf("row: %s\n", result.Record().String())
        }
        if result.Err() != nil {
            fmt.Printf("Query error: %s\n", result.Err().Error())
        }
    }
    // Ensures background processes finishes
    client.Close()
}
```
### Options
The InfluxDBClient uses set of options to configure behavior. These are available in the [Options](https://github.com/influxdata/influxdb-client-go/blob/master/options.go) object
Creating a client instance using
```go
client := influxdb2.NewClient("http://localhost:8086", "my-token")
```
will use the default options.

To set different configuration values, e.g. to set gzip compression and trust all server certificates, get default options 
and change what is needed: 
```go
client := influxdb2.NewClientWithOptions("http://localhost:8086", "my-token", 
    influxdb2.DefaultOptions().
        SetUseGZip(true).
        SetTLSConfig(&tls.Config{
            InsecureSkipVerify: true,
        }))
```
### Writes

Client offers two ways of writing, non-blocking and blocking. 

### Non-blocking write client 
Non-blocking write client uses implicit batching. Data are asynchronously
written to the underlying buffer and they are automatically sent to a server when the size of the write buffer reaches the batch size, default 5000, or the flush interval, default 1s, times out.
Writes are automatically retried on server back pressure.

This write client also offers synchronous blocking method to ensure that write buffer is flushed and all pending writes are finished, 
see [Flush()](https://github.com/influxdata/influxdb-client-go/blob/master/write.go#L24) method.
Always use [Close()](https://github.com/influxdata/influxdb-client-go/blob/master/client.go#L40) method of the client to stop all background processes.
 
Asynchronous write client is recommended for frequent periodic writes.

```go
package main

import (
    "fmt"
    "math/rand"
    "time"
    
    "github.com/influxdata/influxdb-client-go/v2"
)

func main() {
    // Create a new client using an InfluxDB server base URL and an authentication token
    // and set batch size to 20 
    client := influxdb2.NewClientWithOptions("http://localhost:8086", "my-token",
        influxdb2.DefaultOptions().SetBatchSize(20))
    // Get non-blocking write client
    writeAPI := client.WriteAPI("my-org","my-bucket")
    // write some points
    for i := 0; i <100; i++ {
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
        writeAPI.WritePoint(p)
    }
    // Force all unwritten data to be sent
    writeAPI.Flush()
    // Ensures background processes finishes
    client.Close()
}
```

### Reading async errors
[Error()](https://github.com/influxdata/influxdb-client-go/blob/master/write.go#L24) method returns a channel for reading errors which occurs during async writes. This channel is unbuffered and it 
must be read asynchronously otherwise will block write procedure:

```go
package main

import (
    "fmt"
    "math/rand"
    "time"

    "github.com/influxdata/influxdb-client-go/v2"
)

func main() {
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
        }
    }()
    // write some points
    for i := 0; i < 100; i++ {
        // create point
        p := influxdb2.NewPointWithMeasurement("stat").
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
```

### Blocking write client 
Blocking write client writes given point(s) synchronously. It doesn't have implicit batching. Batch is created from given set of points.

```go
package main

import (
    "context"
    "fmt"
    "math/rand"
    "time"

    "github.com/influxdata/influxdb-client-go/v2"
)

func main() {
    // Create a new client using an InfluxDB server base URL and an authentication token
    client := influxdb2.NewClient("http://localhost:8086", "my-token")
    // Get blocking write client
    writeAPI := client.WriteAPIBlocking("my-org","my-bucket")
    // write some points
    for i := 0; i <100; i++ {
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
        err := writeAPI.WritePoint(context.Background(), p)
        if err != nil {
            panic(err)
        }
    }
    // Ensures background processes finishes
    client.Close()
}
```

### Queries
Query client offers two ways of retrieving query results, parsed representation in [QueryTableResult](https://github.com/influxdata/influxdb-client-go/blob/master/query.go#L162) and a raw result string. 

### QueryTableResult 
QueryTableResult offers comfortable way how to deal with flux query CSV response. It parses CSV stream into FluxTableMetaData, FluxColumn and FluxRecord objects
for easy reading the result.

```go
package main

import (
    "context"
    "fmt"

    "github.com/influxdata/influxdb-client-go/v2"
)

func main() {
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
```

### Raw
[QueryRaw()](https://github.com/influxdata/influxdb-client-go/blob/master/query.go#L44) returns raw, unparsed, query result string and process it on your own. Returned csv format  
can be controlled by the third parameter, query dialect.   

```go
package main

import (
    "context"
    "fmt"

    "github.com/influxdata/influxdb-client-go/v2"
)

func main() {
    // Create a new client using an InfluxDB server base URL and an authentication token
    client := influxdb2.NewClient("http://localhost:8086", "my-token")
    // Get query client
    queryAPI := client.QueryAPI("my-org")
    // Query and get complete result as a string
    // Use default dialect
    result, err := queryAPI.QueryRaw(context.Background(), `from(bucket:"my-bucket")|> range(start: -1h) |> filter(fn: (r) => r._measurement == "stat")`, influxdb2.DefaultDialect())
    if err == nil {
        fmt.Println("QueryResult:")
        fmt.Println(result)
    } else {
        panic(err)
    }
    // Ensures background processes finishes
    client.Close()
}    
```


## InfluxDB 1.8 API compatibility
  
  [InfluxDB 1.8.0 introduced forward compatibility APIs](https://docs.influxdata.com/influxdb/latest/tools/api/#influxdb-2-0-api-compatibility-endpoints) for InfluxDB 2.0. This allow you to easily move from InfluxDB 1.x to InfluxDB 2.0 Cloud or open source.
  
  Client API usage differences summary:
 1. Use the form `username:password` for an **authentication token**. Example: `my-user:my-password`. Use an empty string (`""`) if the server doesn't require authentication.
 1. The organization parameter is not used. Use an empty string (`""`) where necessary.
 1. Use the form `database/retention-policy` where a **bucket** is required. Skip retention policy if the default retention policy should be used. Examples: `telegraf/autogen`, `telegraf`.  
  
  The following forward compatible APIs are available:
  
  | API | Endpoint | Description |
  |:----------|:----------|:----------|
  | [WriteAPI](write.go) (also [WriteAPIBlocking](writeAPIBlocking.go))| [/api/v2/write](https://docs.influxdata.com/influxdb/latest/tools/api/#api-v2-write-http-endpoint) | Write data to InfluxDB 1.8.0+ using the InfluxDB 2.0 API |
  | [QueryAPI](query.go) | [/api/v2/query](https://docs.influxdata.com/influxdb/latest/tools/api/#api-v2-query-http-endpoint) | Query data in InfluxDB 1.8.0+ using the InfluxDB 2.0 API and [Flux](https://docs.influxdata.com/flux/latest/) endpoint should be enabled by the [`flux-enabled` option](https://docs.influxdata.com/influxdb/latest/administration/config/#flux-enabled-false)
  | [Health()](client.go#L55) | [/health](https://docs.influxdata.com/influxdb/latest/tools/api/#health-http-endpoint) | Check the health of your InfluxDB instance |    

  
### Example
```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/influxdata/influxdb-client-go/v2"
)

func main() {
    userName := "my-user"
    password := "my-password"
     // Create a new client using an InfluxDB server base URL and an authentication token
    // For authentication token supply a string in the form: "username:password" as a token. Set empty value for an unauthenticated server
    client := influxdb2.NewClient("http://localhost:8086", fmt.Sprintf("%s:%s",userName, password))
    // Get the blocking write client
    // Supply a string in the form database/retention-policy as a bucket. Skip retention policy for the default one, use just a database name (without the slash character)
    // Org name is not used
    writeAPI := client.WriteAPIBlocking("", "test/autogen")
    // create point using full params constructor
    p := influxdb2.NewPoint("stat",
        map[string]string{"unit": "temperature"},
        map[string]interface{}{"avg": 24.5, "max": 45},
        time.Now())
    // Write data
    err := writeAPI.WritePoint(context.Background(), p)
    if err != nil {
        fmt.Printf("Write error: %s\n", err.Error())
    }

    // Get query client. Org name is not used
    queryAPI := client.QueryAPI("")
    // Supply string in a form database/retention-policy as a bucket. Skip retention policy for the default one, use just a database name (without the slash character)
    result, err := queryAPI.Query(context.Background(), `from(bucket:"test")|> range(start: -1h) |> filter(fn: (r) => r._measurement == "stat")`)
    if err == nil {
        for result.Next() {
            if result.TableChanged() {
                fmt.Printf("table: %s\n", result.TableMetadata().String())
            }
            fmt.Printf("row: %s\n", result.Record().String())
        }
        if result.Err() != nil {
            fmt.Printf("Query error: %s\n", result.Err().Error())
        }
    } else {
        fmt.Printf("Query error: %s\n", err.Error())
    }
    // Close client
    client.Close()
}
```
## Contributing

If you would like to contribute code you can do through GitHub by forking the repository and sending a pull request into the `master` branch.

## License

The InfluxDB 2 Go Client is released under the [MIT License](https://opensource.org/licenses/MIT).
