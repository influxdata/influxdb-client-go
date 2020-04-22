// Copyright 2020 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package influxdb2

import (
	"context"
	"flag"
	"fmt"
	"github.com/influxdata/influxdb-client-go/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"strconv"
	"testing"
	"time"
)

var e2e bool

func init() {
	flag.BoolVar(&e2e, "e2e", false, "run the end tests (requires a working influxdb instance on 127.0.0.1)")
	flag.StringVar(&authToken, "token", "", "authentication token")
}

func TestReady(t *testing.T) {
	if !e2e {
		t.Skip("e2e not enabled. Launch InfluxDB 2 on localhost and run test with -e2e")
	}
	client := NewClient("http://localhost:9999", "my-token-123")

	ok, err := client.Ready(context.Background())
	if err != nil {
		t.Error(err)
	}
	if !ok {
		t.Fail()
	}
}

var authToken string

func TestSetup(t *testing.T) {
	if !e2e {
		t.Skip("e2e not enabled. Launch InfluxDB 2 on localhost and run test with -e2e")
	}
	client := NewClientWithOptions("http://localhost:9999", "", DefaultOptions().SetLogLevel(2))
	response, err := client.Setup(context.Background(), "my-user", "my-password", "my-org", "my-bucket", 0)
	if err != nil {
		t.Error(err)
	}
	require.NotNil(t, response)
	authToken = *response.Auth.Token
	fmt.Println("Token:" + authToken)

	_, err = client.Setup(context.Background(), "my-user", "my-password", "my-org", "my-bucket", 0)
	require.NotNil(t, err)
	assert.Equal(t, "conflict: onboarding has already been completed", err.Error())
}
func TestWrite(t *testing.T) {
	if !e2e {
		t.Skip("e2e not enabled. Launch InfluxDB 2 on localhost and run test with -e2e")
	}
	client := NewClientWithOptions("http://localhost:9999", authToken, DefaultOptions().SetLogLevel(3))
	writeApi := client.WriteApi("my-org", "my-bucket")
	errCh := writeApi.Errors()
	errorsCount := 0
	go func() {
		for err := range errCh {
			errorsCount++
			fmt.Println("Write error: ", err.Error())
		}
	}()
	for i, f := 0, 3.3; i < 10; i++ {
		writeApi.WriteRecord(fmt.Sprintf("test,a=%d,b=local f=%.2f,i=%di", i%2, f, i))
		//writeApi.Flush()
		f += 3.3
	}

	for i, f := int64(10), 33.0; i < 20; i++ {
		p := NewPoint("test",
			map[string]string{"a": strconv.FormatInt(i%2, 10), "b": "static"},
			map[string]interface{}{"f": f, "i": i},
			time.Now())
		writeApi.WritePoint(p)
		f += 3.3
	}

	client.Close()
	assert.Equal(t, 0, errorsCount)

}

func TestQueryRaw(t *testing.T) {
	if !e2e {
		t.Skip("e2e not enabled. Launch InfluxDB 2 on localhost and run test with -e2e")
	}
	client := NewClient("http://localhost:9999", authToken)

	queryApi := client.QueryApi("my-org")
	res, err := queryApi.QueryRaw(context.Background(), `from(bucket:"my-bucket")|> range(start: -1h) |> filter(fn: (r) => r._measurement == "test")`, nil)
	if err != nil {
		t.Error(err)
	} else {
		fmt.Println("QueryResult:")
		fmt.Println(res)
	}
}

func TestQuery(t *testing.T) {
	if !e2e {
		t.Skip("e2e not enabled. Launch InfluxDB 2 on localhost and run test with -e2e")
	}
	client := NewClient("http://localhost:9999", authToken)

	queryApi := client.QueryApi("my-org")
	fmt.Println("QueryResult")
	result, err := queryApi.Query(context.Background(), `from(bucket:"my-bucket")|> range(start: -24h) |> filter(fn: (r) => r._measurement == "test")`)
	if err != nil {
		t.Error(err)
	} else {
		for result.Next() {
			if result.TableChanged() {
				fmt.Printf("table: %s\n", result.TableMetadata().String())
			}
			fmt.Printf("row: %sv\n", result.Record().String())
		}
		if result.Err() != nil {
			t.Error(result.Err())
		}
	}
}

func TestAuthorizationsApi(t *testing.T) {
	if !e2e {
		t.Skip("e2e not enabled. Launch InfluxDB 2 on localhost and run test with -e2e")
	}
	client := NewClient("http://localhost:9999", authToken)
	authApi := client.AuthorizationsApi()
	listRes, err := authApi.FindAuthorizations(context.Background())
	require.Nil(t, err)
	require.NotNil(t, listRes)
	assert.Len(t, *listRes, 1)

	orgName := "my-org"
	org, err := client.OrganizationsApi().FindOrganizationByName(context.Background(), orgName)
	require.Nil(t, err)
	require.NotNil(t, org)
	assert.Equal(t, orgName, org.Name)

	permission := &domain.Permission{
		Action: domain.PermissionActionWrite,
		Resource: domain.Resource{
			Type: domain.ResourceTypeBuckets,
		},
	}
	permissions := []domain.Permission{*permission}

	auth, err := authApi.CreateAuthorizationWithOrgID(context.Background(), *org.Id, permissions)
	require.Nil(t, err)
	require.NotNil(t, auth)
	assert.Equal(t, domain.AuthorizationUpdateRequestStatusActive, *auth.Status, *auth.Status)

	listRes, err = authApi.FindAuthorizations(context.Background())
	require.Nil(t, err)
	require.NotNil(t, listRes)
	assert.Len(t, *listRes, 2)

	listRes, err = authApi.FindAuthorizationsByUserName(context.Background(), "my-user")
	require.Nil(t, err)
	require.NotNil(t, listRes)
	assert.Len(t, *listRes, 2)

	listRes, err = authApi.FindAuthorizationsByOrgID(context.Background(), *org.Id)
	require.Nil(t, err)
	require.NotNil(t, listRes)
	assert.Len(t, *listRes, 2)

	listRes, err = authApi.FindAuthorizationsByOrgName(context.Background(), "my-org")
	require.Nil(t, err)
	require.NotNil(t, listRes)
	assert.Len(t, *listRes, 2)

	listRes, err = authApi.FindAuthorizationsByOrgName(context.Background(), "not-existent-org")
	require.Nil(t, listRes)
	require.NotNil(t, err)
	//assert.Len(t, *listRes, 0)

	auth, err = authApi.UpdateAuthorizationStatus(context.Background(), *auth.Id, domain.AuthorizationUpdateRequestStatusInactive)
	require.Nil(t, err)
	require.NotNil(t, auth)
	assert.Equal(t, domain.AuthorizationUpdateRequestStatusInactive, *auth.Status, *auth.Status)

	listRes, err = authApi.FindAuthorizations(context.Background())
	require.Nil(t, err)
	require.NotNil(t, listRes)
	assert.Len(t, *listRes, 2)

	err = authApi.DeleteAuthorization(context.Background(), *auth.Id)
	require.Nil(t, err)

	listRes, err = authApi.FindAuthorizations(context.Background())
	require.Nil(t, err)
	require.NotNil(t, listRes)
	assert.Len(t, *listRes, 1)

}
