// +build e2e

// Copyright 2021 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package influxclient_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	. "github.com/influxdata/influxdb-client-go/influxclient"
	"github.com/stretchr/testify/require"
)

var authToken string
var serverURL string
var orgName string
var bucketName string
var userName string

func getEnvValue(key, defVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	} else {
		return defVal
	}
}

const envPrefix = "INFLUXDB2"

var notInitializedID string
var invalidID = "#1"
var notExistingID = "100000000000000"

func init() {
	authToken = getEnvValue(envPrefix + "_TOKEN", "my-token")
	serverURL = getEnvValue(envPrefix + "_URL", "http://localhost:9999")
	orgName = getEnvValue(envPrefix + "_ORG", "my-org")
	bucketName = getEnvValue(envPrefix + "_BUCKET", "my-bucket")
	userName = getEnvValue(envPrefix + "_USER", "my-user")
	fmt.Printf("E2E testing values:\n  server:  %s\n  token :  %s\n  org   :  %s\n  user  :  %s\n  bucket:  %s\n",
		serverURL, authToken, orgName, userName, bucketName)
}

func newClient(t *testing.T) (*Client, context.Context) {
	c, err := New(Params{
		ServerURL: serverURL,
		AuthToken: authToken,
	})
	require.NoError(t, err)
	require.NotNil(t, c)
	return c, context.Background()
}

func safeId(ID interface{}) string {
	switch v := ID.(type) {
	case string:
		return fmt.Sprintf("%s", v)
	case *string:
		return fmt.Sprintf("%s", *v)
	default: panic("unsupported type")
	}
}