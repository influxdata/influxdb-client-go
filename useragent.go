package influxdb

import (
	"fmt"
	"runtime"
)

func userAgent() string {
	return fmt.Sprintf("InfluxDBClient/0.0.1  (%s; %s; %s)", "golang", runtime.GOOS, runtime.GOARCH)
}
