package client

import (
	"fmt"
	"runtime"
)

func ua() string {
	return fmt.Sprintf("InfluxDBClient/v0.0.1  (%s; %s; %s)", "golang", runtime.GOOS, runtime.GOARCH)
}
