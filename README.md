# influxdb-client-go
A home for InfluxDB’s 2.x's golang client. 

## Pre-Alpha Warning!
This is not even close to ready for prod use.
Keep an eye on this repo if you are interested in the client.


## Example:
```
influx, err := client.NewClient(http.DefaultClient, WithAddress("http:/127.0.0.1:9999"), WithToken("mytoken"))
if err!=nil{
    panic(err)
}
var myMetrics []LMetric

myMetrics = append(myMetrics, client.NewMetric(
    map[string]interface{}{"memory":1000,"cpu":0.93},
    "system-metrics",
    map[string]string{"hostname":"hal9000"},
    time.Now())

influx.Write(context.Background(), "my-awesome-bucket", "my-very-awesome-org", myMetrics...)

```

## Releases
We will be using git-flow style releases, the current stable release will be listed in the master readme.