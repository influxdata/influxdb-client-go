// Copyright 2021 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package influxclient

import (
	"runtime"
)

// version defines current version
const version = "3.0.0alpha1"

// userAgent header value
const userAgent = "influxdb-client-go/" + version + " (" + runtime.GOOS + "; " + runtime.GOARCH + ")"
