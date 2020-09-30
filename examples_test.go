package influxdb2_test

import (
	"context"
	"fmt"

	"github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/domain"
)

func ExampleClient_newClient() {
	// Create a new client using an InfluxDB server base URL and an authentication token
	client := influxdb2.NewClient("http://localhost:8086", "my-token")

	// Always close client at the end
	defer client.Close()
}

func ExampleClient_newClientWithOptions() {
	// Create a new client using an InfluxDB server base URL and an authentication token
	// Create client and set batch size to 20
	client := influxdb2.NewClientWithOptions("http://localhost:8086", "my-token",
		influxdb2.DefaultOptions().SetBatchSize(20))

	// Always close client at the end
	defer client.Close()
}

func ExampleClient_customServerAPICall() {
	// Create a new client using an InfluxDB server base URL and empty token
	client := influxdb2.NewClient("http://localhost:8086", "my-token")
	// Always close client at the end
	defer client.Close()
	// Get generated client for server API calls
	apiClient := domain.NewClientWithResponses(client.HTTPService())
	// Get an organization that will own task
	org, err := client.OrganizationsAPI().FindOrganizationByName(context.Background(), "my-org")
	if err != nil {
		//return err
		panic(err)
	}

	// Basic task properties
	taskDescription := "Example task"
	taskFlux := `option task = {
  name: "My task",
  every: 1h
}

from(bucket:"my-bucket") |> range(start: -1m) |> last()`
	taskStatus := domain.TaskStatusTypeActive

	// Create TaskCreateRequest object
	taskRequest := domain.TaskCreateRequest{
		Org:         &org.Name,
		OrgID:       org.Id,
		Description: &taskDescription,
		Flux:        taskFlux,
		Status:      &taskStatus,
	}

	// Issue an API call
	resp, err := apiClient.PostTasksWithResponse(context.Background(), &domain.PostTasksParams{}, domain.PostTasksJSONRequestBody(taskRequest))
	if err != nil {
		panic(err)
	}

	// Always check generated response errors
	if resp.JSONDefault != nil {
		panic(resp.JSONDefault.Message)
	}

	// Use API call result
	task := resp.JSON201
	fmt.Println("Created task: ", task.Name)
}
