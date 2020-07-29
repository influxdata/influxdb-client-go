// +build e2e

// Copyright 2020 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package influxdb2_test

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go"
	"github.com/influxdata/influxdb-client-go/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var authToken string
var serverURL string
var serverV1URL string
var onboardingURL string

func getEnvValue(key, defVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	} else {
		return defVal
	}
}

func init() {
	authToken = getEnvValue("INFLUXDB2_TOKEN", "my-token")
	serverURL = getEnvValue("INFLUXDB2_URL", "http://localhost:9999")
	serverV1URL = getEnvValue("INFLUXDB_URL", "http://localhost:8086")
	onboardingURL = getEnvValue("INFLUXDB2_ONBOARDING_URL", "http://localhost:9990")
}

func TestSetup(t *testing.T) {
	client := influxdb2.NewClientWithOptions(onboardingURL, "", influxdb2.DefaultOptions().SetLogLevel(2))
	response, err := client.Setup(context.Background(), "my-user", "my-password", "my-org", "my-bucket", 0)
	if err != nil {
		t.Error(err)
	}
	require.NotNil(t, response)
	require.NotNil(t, response.Auth)
	require.NotNil(t, response.Auth.Token)

	_, err = client.Setup(context.Background(), "my-user", "my-password", "my-org", "my-bucket", 0)
	require.NotNil(t, err)
	assert.Equal(t, "conflict: onboarding has already been completed", err.Error())
}

func TestReady(t *testing.T) {
	client := influxdb2.NewClient(serverURL, "")

	ok, err := client.Ready(context.Background())
	if err != nil {
		t.Error(err)
	}
	if !ok {
		t.Fail()
	}
}

func TestHealth(t *testing.T) {
	client := influxdb2.NewClient(serverURL, "")

	health, err := client.Health(context.Background())
	if err != nil {
		t.Error(err)
	}
	require.NotNil(t, health)
	assert.Equal(t, domain.HealthCheckStatusPass, health.Status)
}

func TestWrite(t *testing.T) {
	client := influxdb2.NewClientWithOptions(serverURL, authToken, influxdb2.DefaultOptions().SetLogLevel(3))
	writeAPI := client.WriteAPI("my-org", "my-bucket")
	errCh := writeAPI.Errors()
	errorsCount := 0
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		for err := range errCh {
			errorsCount++
			fmt.Println("Error proc: write error: ", err.Error())
		}
		fmt.Println("Error proc: finished ")
		wg.Done()
	}()
	timestamp := time.Now()
	for i, f := 0, 3.3; i < 10; i++ {
		writeAPI.WriteRecord(fmt.Sprintf("test,a=%d,b=local f=%.2f,i=%di %d", i%2, f, i, timestamp.UnixNano()))
		//writeAPI.Flush()
		f += 3.3
		timestamp = timestamp.Add(time.Nanosecond)
	}

	for i, f := int64(10), 33.0; i < 20; i++ {
		p := influxdb2.NewPoint("test",
			map[string]string{"a": strconv.FormatInt(i%2, 10), "b": "static"},
			map[string]interface{}{"f": f, "i": i},
			timestamp)
		writeAPI.WritePoint(p)
		f += 3.3
		timestamp = timestamp.Add(time.Nanosecond)
	}

	err := client.WriteAPIBlocking("my-org", "my-bucket").WritePoint(context.Background(), influxdb2.NewPointWithMeasurement("test").
		AddTag("a", "3").AddField("i", 20).AddField("f", 4.4))
	assert.Nil(t, err)

	client.Close()
	wg.Wait()
	assert.Equal(t, 0, errorsCount)

}

func TestQueryRaw(t *testing.T) {
	client := influxdb2.NewClient(serverURL, authToken)

	queryAPI := client.QueryAPI("my-org")
	res, err := queryAPI.QueryRaw(context.Background(), `from(bucket:"my-bucket")|> range(start: -24h) |> filter(fn: (r) => r._measurement == "test")`, influxdb2.DefaultDialect())
	if err != nil {
		t.Error(err)
	} else {
		fmt.Println("QueryResult:")
		fmt.Println(res)
	}
}

func TestQuery(t *testing.T) {
	client := influxdb2.NewClient(serverURL, authToken)

	queryAPI := client.QueryAPI("my-org")
	fmt.Println("QueryResult")
	result, err := queryAPI.Query(context.Background(), `from(bucket:"my-bucket")|> range(start: -24h) |> filter(fn: (r) => r._measurement == "test")`)
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

func TestHealthV1Compatibility(t *testing.T) {
	client := influxdb2.NewClient(serverV1URL, "")

	health, err := client.Health(context.Background())
	if err != nil {
		t.Error(err)
	}
	require.NotNil(t, health)
	assert.Equal(t, domain.HealthCheckStatusPass, health.Status)
}

func TestWriteV1Compatibility(t *testing.T) {
	client := influxdb2.NewClientWithOptions(serverV1URL, "", influxdb2.DefaultOptions().SetLogLevel(3))
	writeAPI := client.WriteAPI("", "mydb/autogen")
	errCh := writeAPI.Errors()
	errorsCount := 0
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		for err := range errCh {
			errorsCount++
			fmt.Println("Error proc: write error: ", err.Error())
		}
		wg.Done()
	}()
	timestamp := time.Now()
	for i, f := 0, 3.3; i < 10; i++ {
		writeAPI.WriteRecord(fmt.Sprintf("testv1,a=%d,b=local f=%.2f,i=%di %d", i%2, f, i, timestamp.UnixNano()))
		//writeAPI.Flush()
		f += 3.3
		timestamp = timestamp.Add(time.Nanosecond)
	}

	for i, f := int64(10), 33.0; i < 20; i++ {
		p := influxdb2.NewPoint("testv1",
			map[string]string{"a": strconv.FormatInt(i%2, 10), "b": "static"},
			map[string]interface{}{"f": f, "i": i},
			timestamp)
		writeAPI.WritePoint(p)
		f += 3.3
		timestamp = timestamp.Add(time.Nanosecond)
	}

	err := client.WriteAPIBlocking("", "mydb/autogen").WritePoint(context.Background(), influxdb2.NewPointWithMeasurement("testv1").
		AddTag("a", "3").AddField("i", 20).AddField("f", 4.4))
	assert.Nil(t, err)

	client.Close()
	wg.Wait()
	assert.Equal(t, 0, errorsCount)

}

func TestQueryRawV1Compatibility(t *testing.T) {
	client := influxdb2.NewClient(serverV1URL, "")

	queryAPI := client.QueryAPI("")
	res, err := queryAPI.QueryRaw(context.Background(), `from(bucket:"mydb/autogen")|> range(start: -24h) |> filter(fn: (r) => r._measurement == "testv1")`, influxdb2.DefaultDialect())
	if err != nil {
		t.Error(err)
	} else {
		fmt.Println("QueryResult:")
		fmt.Println(res)
	}
}

func TestQueryV1Compatibility(t *testing.T) {
	client := influxdb2.NewClient(serverV1URL, "")

	queryAPI := client.QueryAPI("")
	fmt.Println("QueryResult")
	result, err := queryAPI.Query(context.Background(), `from(bucket:"mydb/autogen")|> range(start: -24h) |> filter(fn: (r) => r._measurement == "testv1")`)
	if err != nil {
		t.Error(err)
	} else {
		rows := 0
		for result.Next() {
			rows++
			if result.TableChanged() {
				fmt.Printf("table: %s\n", result.TableMetadata().String())
			}
			fmt.Printf("row: %sv\n", result.Record().String())
		}
		if result.Err() != nil {
			t.Error(result.Err())
		}
		assert.Equal(t, 42, rows)
	}

}
