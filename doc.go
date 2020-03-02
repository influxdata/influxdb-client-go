/*
Package influxdb implements InfluxDB’s 2.x's golang client. This client is not compatible with InfluxDB 1.x -- if you are looking for the 1.x golang client you can find it at https://github.com/influxdata/influxdb1-client.

Example:

	influx, err := influxdb.New(
		myHTTPInfluxAddress,
		myToken,
		influxdb.WithHTTPClient(myHTTPClient))

	if err != nil {
		return err
	}
	defer influx.Close() // closes the client.  After this the client is useless.

	// we use client.NewRowMetric for the example because it's easy, but if you
	// need extra performance it is fine to manually build the []client.Metric{}.
	metrics := []influxdb.Metric{
		influxdb.NewRowMetric(
			map[string]interface{}{"memory": 1000, "cpu": 0.93},
			"system-metrics",
			map[string]string{"hostname": "hal9000"},
			time.Date(2018, 3, 4, 5, 6, 7, 8, time.UTC)),
		influxdb.NewRowMetric(
			map[string]interface{}{"memory": 1000, "cpu": 0.93},
			"system-metrics",
			map[string]string{"hostname": "hal9000"},
			time.Date(2018, 3, 4, 5, 6, 7, 9, time.UTC)),
	}

	// the actual write. this method can be called concurrently.
	if _, err := influx.write(context.background(), "my-awesome-bucket", "my-very-awesome-org", metrics...); err != nil {
		return err
	}
*/
package influxdb
