package influxdb2_test

import (
	"context"
	"fmt"

	"github.com/influxdata/influxdb-client-go"
	"github.com/influxdata/influxdb-client-go/domain"
)

func ExampleClient_newClient() {
	// Create client
	client := influxdb2.NewClient("http://localhost:9999", "my-token")

	// always close client at the end
	defer client.Close()
}

func ExampleClient_newClientWithOptions() {
	// Create client and set batch size to 20
	client := influxdb2.NewClientWithOptions("http://localhost:9999", "my-token",
		influxdb2.DefaultOptions().SetBatchSize(20))

	// always close client at the end
	defer client.Close()
}

func ExampleClient_customServerAPICall() {
	client := influxdb2.NewClient("http://localhost:9999", "my-token")
	apiClient := domain.NewClientWithResponses(client.HTTPService())
	org, err := client.OrganizationsAPI().FindOrganizationByName(context.Background(), "my-org")
	if err != nil {
		//return err
		panic(err)
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
		panic(err)
	}
	if resp.JSONDefault != nil {
		panic(resp.JSONDefault.Message)
	}
	task := resp.JSON201
	fmt.Println("Created task: ", task.Name)
}
