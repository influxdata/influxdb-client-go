# InfluxDB Client Go

[![CircleCI](https://circleci.com/gh/influxdata/influxdb-client-go.svg?style=svg)](https://circleci.com/gh/influxdata/influxdb-client-go)
[![codecov](https://codecov.io/gh/influxdata/influxdb-client-go/branch/master/graph/badge.svg)](https://codecov.io/gh/influxdata/influxdb-client-go)
[![License](https://img.shields.io/influxdata/license/influxdata/influxdb-client-go.svg)](https://github.com/influxdata/influxdb-client-go/blob/master/LICENSE)

This repository contains the reference Go client for InfluxDB 2.

## Features

- InfluxDB 2 client
    - Querying data 
        - using the Flux language
        - into raw data, flux table representation
        - [How to queries](#queries)
    - Writing data using
        - [Line Protocol](https://docs.influxdata.com/influxdb/v1.6/write_protocols/line_protocol_tutorial/) 
        - [Data Point](https://github.com/influxdata/influxdb-client-go/blob/master/point.go)
        - both [asynchronous](https://github.com/influxdata/influxdb-client-go/blob/master/write.go) or [synchronous](https://github.com/influxdata/influxdb-client-go/blob/master/writeApiBlocking.go) manners
        - [How to writes](#writes)  
    - InfluxDB 2 API
        - setup
        - ready
     
## Installation
**Go 1.3** or later is required.

Add import `github.com/influxdata/influxdb-client-go` to your source code and sync dependencies or directly edit go.mod.

## Usage
Basic example with blocking write and flux query:
```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/influxdata/influxdb-client-go"
)

func main() {
    // create new client with default option for server url authenticate by token
    client := influxdb2.NewClient("http://localhost:9999", "my-token")
    // user blocking write client for writes to desired bucket
    writeApi := client.WriteApiBlocking("my-org", "my-bucket")
    // create point using full params constructor 
    p := influxdb2.NewPoint("stat",
        map[string]string{"unit": "temperature"},
        map[string]interface{}{"avg": 24.5, "max": 45},
        time.Now())
    // write point immediately 
    writeApi.WritePoint(context.Background(), p)
    // create point using fluent style
    p = influxdb2.NewPointWithMeasurement("stat").
        AddTag("unit", "temperature").
        AddField("avg", 23.2).
        AddField("max", 45).
        SetTime(time.Now())
    writeApi.WritePoint(context.Background(), p)
    
    // Or write directly line protocol
    line := fmt.Sprintf("stat,unit=temperature avg=%f,max=%f", 23.5, 45.0)
    writeApi.WriteRecord(context.Background(), line)

    // get query client
    queryApi := client.QueryApi("my-org")
    // get parser flux query result
    result, err := queryApi.Query(context.Background(), `from(bucket:"my-bucket")|> range(start: -1h) |> filter(fn: (r) => r._measurement == "stat")`)
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
    
}
```
### Options

Client uses set of options to configure behavior. These are available in the [Options](https://github.com/influxdata/influxdb-client-go/blob/master/options.go) object
Creating client using:
```go
client := influxdb2.NewClient("http://localhost:9999", "my-token")
```

To set different configuration values, e.g. to set gzip compression and trust all server certificates, get default options 
and change what needed: 
```go
    client := influxdb2.NewClientWithOptions("http://localhost:9999", "my-token", 
        influxdb2.DefaultOptions().
            SetUseGZip(true).
            SetTlsConfig(&tls.Config{
                InsecureSkipVerify: true,
            }))
```
### Writes

Client offers two ways of writing, non-blocking and blocking. 

### Non-blocking write client 
Non-blocking write client uses implicit batching. Data are asynchronously
written to the underlying buffer and are automatically sent to server when size of the write buffer reaches the batch size, default 1000, or flush interval, default 1s, times out.
Writes are automatically retried on server back pressure.

This client also offers synchronous blocking method to ensure that write buffer is flushed and all pending writes are finished, 
see [Flush()](https://github.com/influxdata/influxdb-client-go/blob/master/write.go#L24) method.
Always use [Close()](https://github.com/influxdata/influxdb-client-go/blob/master/write.go#L26) method of the client to stop all background processes.
 
This write client recommended for frequent periodic writes.

```go
package main

import (
    "fmt"
    "github.com/influxdata/influxdb-client-go"
    "math/rand"
    "time"
)

func main() {
    // Create client and set batch size to 20 
    client := influxdb2.NewClientWithOptions("http://localhost:9999", "my-token",
        influxdb2.DefaultOptions().SetBatchSize(20))
    // Get non-blocking write client
    writeApi := client.WriteApi("my-org","my-bucket")
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
        writeApi.WritePoint(p)
    }
    // Force all unwritten data to be sent
    writeApi.Flush()
    // Ensures background processes finishes
    client.Close()
}
```
### Blocking write client 
Blocking write client writes given point(s) synchronously. No implicit batching. Batch is created from given set of points

```go
package main

import (
    "context"
    "fmt"
    "github.com/influxdata/influxdb-client-go"
    "math/rand"
    "time"
)

func main() {
    // Create client
    client := influxdb2.NewClient("http://localhost:9999", "my-token")
    // Get non-blocking write client
    writeApi := client.WriteApiBlocking("my-org","my-bucket")
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
        err := writeApi.WritePoint(context.Background(), p)
        if err != nil {
            panic(err)
        }
    }
}
```

### Queries
Query client offer two ways of retrieving query results, parsed representation in [QueryTableResult](https://github.com/influxdata/influxdb-client-go/blob/master/query.go#L162) and a raw result string. 
which parses response stream into FluxTableMetaData, FluxColumn and FluxRecord objects.

### QueryTableResult 
QueryTableResult offers comfortable way how to deal with flux query CSV response. It parses CSV stream into FluxTableMetaData, FluxColumn and FluxRecord objects
for easy reading the result.

```go
package main

import (
    "context"
    "fmt"
    "github.com/influxdata/influxdb-client-go"
)

func main() {
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
}
```

### Raw
[QueryRaw()](https://github.com/influxdata/influxdb-client-go/blob/master/query.go#L44) returns raw, unparsed, query result string and process it on your own. Returned csv format  
can controlled by the third parameter, query dialect.   

```go
package main

import (
    "context"
    "fmt"
    "github.com/influxdata/influxdb-client-go"
)

func main() {
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
}    
```

## Contributing

If you would like to contribute code you can do through GitHub by forking the repository and sending a pull request into the `master` branch.

## License

The InfluxDB 2 Go Client is released under the [MIT License](https://opensource.org/licenses/MIT).
