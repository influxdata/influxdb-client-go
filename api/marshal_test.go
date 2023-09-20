package api

import "time"

type InfluxTestType struct {
	Name      string    `influxdb:"name"`
	Title     string    `influxdb:"title"`
	Timestamp time.Time `influxdb:"timestamp"`
}
