// Copyright 2022 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package influxdb2_test

import (
	"context"
	"fmt"
	"net/http"

	"github.com/influxdata/influxdb-client-go/v2"
	ihttp "github.com/influxdata/influxdb-client-go/v2/api/http"
)

// UserAgentSetter is the implementation of Doer interface for setting User-Agent header
type UserAgentSetter struct {
	UserAgent   string
	RequestDoer ihttp.Doer
}

// Do fulfills the Doer interface
func (u *UserAgentSetter) Do(req *http.Request) (*http.Response, error) {
	// Set User-Agent header to request
	req.Header.Set("User-Agent", u.UserAgent)
	// Call original Doer to proceed with request
	return u.RequestDoer.Do(req)
}

func ExampleClient_customUserAgentHeader() {
	// Set custom Doer to HTTPOptions
	opts := influxdb2.DefaultOptions()
	opts.HTTPOptions().SetHTTPDoer(&UserAgentSetter{
		UserAgent:   "NetMonitor/1.1",
		RequestDoer: http.DefaultClient,
	})

	//Create client with customized options
	client := influxdb2.NewClientWithOptions("http://localhost:8086", "my-token", opts)

	// Always close client at the end
	defer client.Close()

	// Issue a call with custom User-Agent header
	resp, err := client.Ping(context.Background())
	if err != nil {
		panic(err)
	}
	if resp {
		fmt.Println("Server is up")
	} else {
		fmt.Println("Server is down")
	}
}
