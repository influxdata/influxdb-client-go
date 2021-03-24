// +build e2e

// Copyright 2020-2021 InfluxData, Inc. All rights reserved.
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

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/domain"
	"github.com/influxdata/influxdb-client-go/v2/log"
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
	serverURL = getEnvValue("INFLUXDB2_URL", "http://localhost:8086")
	serverV1URL = getEnvValue("INFLUXDB_URL", "http://localhost:8087")
	onboardingURL = getEnvValue("INFLUXDB2_ONBOARDING_URL", "http://localhost:8089")
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
	assert.NoError(t, err)

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
	client := influxdb2.NewClientWithOptions(serverV1URL, "", influxdb2.DefaultOptions().SetLogLevel(log.DebugLevel))
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
	assert.NoError(t, err)

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
		assert.True(t, rows > 0)
	}
}

func TestV2APIAgainstV1Server(t *testing.T) {
	client := influxdb2.NewClient(serverV1URL, "")
	ctx := context.Background()
	_, err := client.AuthorizationsAPI().GetAuthorizations(ctx)
	require.Error(t, err)
	_, err = client.UsersAPI().GetUsers(ctx)
	require.Error(t, err)
	_, err = client.OrganizationsAPI().GetOrganizations(ctx)
	require.Error(t, err)
	_, err = client.TasksAPI().FindTasks(ctx, nil)
	require.Error(t, err)
	_, err = client.LabelsAPI().GetLabels(ctx)
	require.Error(t, err)
	_, err = client.BucketsAPI().GetBuckets(ctx)
	require.Error(t, err)
	err = client.DeleteAPI().DeleteWithName(ctx, "org", "bucket", time.Now(), time.Now(), "")
	require.Error(t, err)
}

func TestHTTPService(t *testing.T) {
	client := influxdb2.NewClient("http://localhost:8086", "my-token")
	apiClient := domain.NewClientWithResponses(client.HTTPService())
	org, err := client.OrganizationsAPI().FindOrganizationByName(context.Background(), "my-org")
	if err != nil {
		//return err
		t.Fatal(err)
	}
	taskDescription := "Example task"
	taskFlux := `option task = {
  name: "My task",
  every: 1h
}

from(bucket:"my-bucket") |> range(start: -1m) |> last()`
	taskStatus := domain.TaskStatusTypeActive
	taskRequest := domain.TaskCreateRequest{
		Org:         &org.Name,
		OrgID:       org.Id,
		Description: &taskDescription,
		Flux:        taskFlux,
		Status:      &taskStatus,
	}
	resp, err := apiClient.PostTasksWithResponse(context.Background(), &domain.PostTasksParams{}, domain.PostTasksJSONRequestBody(taskRequest))
	if err != nil {
		//return err
		t.Error(err)
	}
	if resp.JSONDefault != nil {
		t.Error(resp.JSONDefault.Message)
	}
	if assert.NotNil(t, resp.JSON201) {
		assert.Equal(t, "My task", resp.JSON201.Name)
		_, err := apiClient.DeleteTasksID(context.Background(), resp.JSON201.Id, &domain.DeleteTasksIDParams{})
		if err != nil {
			//return err
			t.Error(err)
		}
	}
}
