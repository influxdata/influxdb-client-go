//go:build e2e
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
	"strings"
	"sync"
	"testing"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api/http"
	"github.com/influxdata/influxdb-client-go/v2/domain"
	"github.com/influxdata/influxdb-client-go/v2/internal/test"
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
	response, err := client.Setup(context.Background(), "my-user", "my-password", "my-org", "my-bucket", 24)
	if err != nil {
		t.Error(err)
	}
	require.NotNil(t, response)
	require.NotNil(t, response.Auth)
	require.NotNil(t, response.Auth.Token)
	require.NotNil(t, response.Bucket)
	require.NotNil(t, response.Bucket.RetentionRules)
	require.Len(t, response.Bucket.RetentionRules, 1)
	assert.Equal(t, int64(24*3600), response.Bucket.RetentionRules[0].EverySeconds)

	_, err = client.Setup(context.Background(), "my-user", "my-password", "my-org", "my-bucket", 0)
	require.NotNil(t, err)
	assert.Equal(t, "conflict: onboarding has already been completed", err.Error())
}

func TestReady(t *testing.T) {
	client := influxdb2.NewClient(serverURL, "")

	ready, err := client.Ready(context.Background())
	require.NoError(t, err)
	require.NotNil(t, ready)
	require.NotNil(t, ready.Started)
	assert.True(t, ready.Started.Before(time.Now()))
	dur, err := time.ParseDuration(*ready.Up)
	require.NoError(t, err)
	assert.True(t, dur.Seconds() > 0)
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

func TestPing(t *testing.T) {
	client := influxdb2.NewClient(serverURL, "")

	ok, err := client.Ping(context.Background())
	require.NoError(t, err)
	assert.True(t, ok)
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

func TestPingV1(t *testing.T) {
	client := influxdb2.NewClient(serverV1URL, "")

	ok, err := client.Ping(context.Background())
	require.NoError(t, err)
	assert.True(t, ok)
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
	client := influxdb2.NewClient(serverURL, authToken)
	apiClient := client.APIClient()
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
	params := &domain.PostTasksAllParams{
		Body: domain.PostTasksJSONRequestBody(taskRequest),
	}
	resp, err := apiClient.PostTasks(context.Background(), params)
	if err != nil {
		//return err
		t.Error(err)
	}
	if assert.NotNil(t, resp) {
		assert.Equal(t, "My task", resp.Name)
		deleteParams := &domain.DeleteTasksIDAllParams{
			TaskID: resp.Id,
		}
		err := apiClient.DeleteTasksID(context.Background(), deleteParams)
		if err != nil {
			//return err
			t.Error(err)
		}
	}
}

func TestLogsConcurrent(t *testing.T) {
	var wg sync.WaitGroup
	w := func(loc string, temp float32) {
		client1 := influxdb2.NewClientWithOptions(serverURL, authToken, influxdb2.DefaultOptions().SetLogLevel(log.ErrorLevel))
		for i := 0; i < 10000; i++ {
			client1.WriteAPI("my-org", "my-bucket").WriteRecord(fmt.Sprintf("room,location=%s temp=%f", loc, temp))
		}
		client1.Close()
		wg.Done()
	}
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go w(fmt.Sprintf("T%d", i), 23.3+float32(i))
		<-time.After(time.Nanosecond)
	}
	wg.Wait()
}

func TestWriteCustomBatch(t *testing.T) {
	client := influxdb2.NewClientWithOptions(serverURL, authToken, influxdb2.DefaultOptions().SetLogLevel(0))

	now := time.Now()
	lines := test.GenRecords(10)
	err := client.WriteAPIBlocking("my-org", "my-bucket").WriteRecord(context.Background(), strings.Join(lines, "\n"))
	assert.NoError(t, err)
	result, err := client.QueryAPI("my-org").Query(context.Background(), fmt.Sprintf(`from(bucket:"my-bucket")|> range(start: %s) |> filter(fn: (r) => r._measurement == "test") |> pivot(rowKey:["_time"], columnKey: ["_field"], valueColumn: "_value")`, now.Format(time.RFC3339Nano)))
	assert.NoError(t, err)
	l := 0
	for result.Next() {
		l++
	}
	assert.Equal(t, 10, l)
}

func TestHttpHeadersInError(t *testing.T) {
	client := influxdb2.NewClientWithOptions(serverURL, authToken, influxdb2.DefaultOptions().SetLogLevel(0))
	err := client.WriteAPIBlocking("my-org", "my-bucket").WriteRecord(context.Background(), "asdf")
	assert.Error(t, err)
	assert.Len(t, err.(*http.Error).Header, 6)
	assert.NotEqual(t, err.(*http.Error).Header.Get("Date"), "")
	assert.NotEqual(t, err.(*http.Error).Header.Get("Content-Length"), "")
	assert.NotEqual(t, err.(*http.Error).Header.Get("Content-Type"), "")
	assert.NotEqual(t, err.(*http.Error).Header.Get("X-Platform-Error-Code"), "")
	assert.Contains(t, err.(*http.Error).Header.Get("X-Influxdb-Version"), "v")
	assert.Equal(t, err.(*http.Error).Header.Get("X-Influxdb-Build"), "OSS")
}
