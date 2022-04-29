//go:build e2e
// +build e2e

// Copyright 2020-2021 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package api_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//
const taskFlux = `from(bucket:"my-bucket") |> range(start: -1h) |> last()`

func TestTasksAPI_CRUDTask(t *testing.T) {
	client := influxdb2.NewClient(serverURL, authToken)
	tasksAPI := client.TasksAPI()
	ctx := context.Background()

	tasks, err := tasksAPI.FindTasks(ctx, nil)
	require.Nil(t, err)
	require.Len(t, tasks, 0)

	org, err := client.OrganizationsAPI().FindOrganizationByName(ctx, "my-org")
	require.Nil(t, err)
	require.NotNil(t, org)

	taskDescription := "Example task"
	taskStatus := domain.TaskStatusTypeInactive
	taskEvery := "5s"
	newTask := &domain.Task{
		Description: &taskDescription,
		Every:       &taskEvery,
		Flux:        taskFlux,
		Name:        "task 01",
		OrgID:       *org.Id,
		Status:      &taskStatus,
	}

	task1, err := tasksAPI.CreateTask(ctx, newTask)
	require.Nil(t, err)
	require.NotNil(t, task1)

	assert.Equal(t, "task 01", task1.Name, task1.Name)
	if assert.NotNil(t, task1.Description) {
		assert.Equal(t, taskDescription, *task1.Description, *task1.Description)
	}
	if assert.NotNil(t, task1.Every) {
		assert.Equal(t, "5s", *task1.Every, *task1.Every)
	}
	if assert.NotNil(t, task1.Status) {
		assert.Equal(t, taskStatus, *task1.Status, *task1.Status)
	}
	assert.Equal(t, *org.Id, task1.OrgID, task1.OrgID)

	task2, err := tasksAPI.CreateTaskWithEvery(ctx, "task 02", taskFlux, "1h", *org.Id)
	require.Nil(t, err)
	require.NotNil(t, task2)

	assert.Equal(t, "task 02", task2.Name, task2.Name)
	assert.Nil(t, task2.Description)
	if assert.NotNil(t, task2.Every) {
		assert.Equal(t, "1h", *task2.Every, *task2.Every)
	}
	if assert.NotNil(t, task2.Status) {
		assert.Equal(t, domain.TaskStatusTypeActive, *task2.Status, *task2.Status)
	}
	assert.Equal(t, *org.Id, task2.OrgID, task2.OrgID)

	task3, err := tasksAPI.CreateTaskWithCron(ctx, "task 03", taskFlux, "*/1 * * * *", *org.Id)
	require.Nil(t, err)
	require.NotNil(t, task3)

	assert.Equal(t, "task 03", task3.Name, task3.Name)
	assert.Nil(t, task3.Description)
	if assert.NotNil(t, task3.Cron) {
		assert.Equal(t, "*/1 * * * *", *task3.Cron, *task3.Cron)
	}
	if assert.NotNil(t, task3.Status) {
		assert.Equal(t, domain.TaskStatusTypeActive, *task3.Status, *task3.Status)
	}
	assert.Equal(t, *org.Id, task3.OrgID, task3.OrgID)

	tasks, err = tasksAPI.FindTasks(ctx, nil)
	require.Nil(t, err)
	assert.Len(t, tasks, 3)

	task3.Every = &taskEvery
	task3.Description = &taskDescription
	task3.Status = &taskStatus
	task, err := tasksAPI.UpdateTask(ctx, task3)
	require.Nil(t, err)
	require.NotNil(t, task)

	assert.Equal(t, "task 03", task.Name, task.Name)
	if assert.NotNil(t, task.Description) {
		assert.Equal(t, taskDescription, *task.Description, *task.Description)
	}
	if assert.NotNil(t, task.Every) {
		assert.Equal(t, taskEvery, *task3.Every, *task3.Every)
	}
	if assert.NotNil(t, task3.Status) {
		assert.Equal(t, taskStatus, *task3.Status, *task3.Status)
	}

	flux := `import "types"
option task = { 
  name: "task 04",
  every: 1h,
}

from(bucket: "my-bucket")
    |> range(start: -task.every)
    |> filter(fn: (r) => r._measurement == "mem" and r.host == "myHost")`
	task4, err := tasksAPI.CreateTaskByFlux(ctx, flux, *org.Id)
	require.Nil(t, err)
	require.NotNil(t, task4)

	assert.Equal(t, "task 04", task4.Name, task4.Name)
	assert.Nil(t, task4.Description)
	if assert.NotNil(t, task4.Every) {
		assert.Equal(t, "1h", *task4.Every, *task4.Every)
	}
	if assert.NotNil(t, task4.Status) {
		assert.Equal(t, domain.TaskStatusTypeActive, *task4.Status, *task4.Status)
	}
	assert.Equal(t, *org.Id, task4.OrgID, task4.OrgID)

	err = tasksAPI.DeleteTask(ctx, task1)
	assert.Nil(t, err)

	err = tasksAPI.DeleteTask(ctx, task2)
	assert.Nil(t, err)

	err = tasksAPI.DeleteTask(ctx, task3)
	assert.Nil(t, err)

	err = tasksAPI.DeleteTask(ctx, task4)
	assert.Nil(t, err)

	tasks, err = tasksAPI.FindTasks(ctx, nil)
	require.Nil(t, err)
	assert.Len(t, tasks, 0)
}
func TestTasksAPI_GetTasks(t *testing.T) {
	client := influxdb2.NewClient(serverURL, authToken)
	tasksAPI := client.TasksAPI()
	ctx := context.Background()

	tasks, err := tasksAPI.FindTasks(ctx, nil)
	require.Nil(t, err)
	require.Len(t, tasks, 0)

	taskOrg, err := client.OrganizationsAPI().CreateOrganizationWithName(ctx, "task-org")
	require.Nil(t, err)
	require.NotNil(t, taskOrg)

	newtasks := make([]*domain.Task, 30)

	for i := 0; i < 30; i++ {
		newtasks[i], err = tasksAPI.CreateTaskWithEvery(ctx, fmt.Sprintf("task %02d", i+1), taskFlux, "1h", *taskOrg.Id)
		require.Nil(t, err)
		require.NotNil(t, newtasks[i])
	}

	tasks, err = tasksAPI.FindTasks(ctx, &api.TaskFilter{Limit: 15})
	require.Nil(t, err)
	require.Len(t, tasks, 15)

	tasks, err = tasksAPI.FindTasks(ctx, &api.TaskFilter{After: tasks[14].Id})
	require.Nil(t, err)
	require.Len(t, tasks, 15)

	tasks, err = tasksAPI.FindTasks(ctx, &api.TaskFilter{Limit: 100})
	require.Nil(t, err)
	require.Len(t, tasks, 30)

	for i := 0; i < 30; i++ {
		err = tasksAPI.DeleteTaskWithID(ctx, newtasks[i].Id)
		assert.Nil(t, err)
	}

	err = client.OrganizationsAPI().DeleteOrganization(ctx, taskOrg)
	assert.Nil(t, err)

}

func TestTasksAPI_FindTasks(t *testing.T) {
	client := influxdb2.NewClient(serverURL, authToken)
	tasksAPI := client.TasksAPI()
	ctx := context.Background()

	tasks, err := tasksAPI.FindTasks(ctx, nil)
	require.Nil(t, err)
	require.Len(t, tasks, 0)

	taskOrg, err := client.OrganizationsAPI().CreateOrganizationWithName(ctx, "task-org")
	require.Nil(t, err)
	require.NotNil(t, taskOrg)

	myuser, err := client.UsersAPI().FindUserByName(ctx, "my-user")
	require.Nil(t, err)
	require.NotNil(t, myuser)

	task1, err := tasksAPI.CreateTaskWithEvery(ctx, "task 01", taskFlux, "1h", *taskOrg.Id)
	require.Nil(t, err)
	require.NotNil(t, task1)

	tasks, err = tasksAPI.FindTasks(ctx, &api.TaskFilter{Name: "task 01"})
	require.Nil(t, err)
	require.Len(t, tasks, 1)

	task2, err := tasksAPI.CreateTaskWithEvery(ctx, "task 01", taskFlux, "1m", *taskOrg.Id)
	require.Nil(t, err)
	require.NotNil(t, task2)

	tasks, err = tasksAPI.FindTasks(ctx, &api.TaskFilter{Name: "task 01"})
	require.Nil(t, err)
	require.Len(t, tasks, 2)

	tasks, err = tasksAPI.FindTasks(ctx, &api.TaskFilter{OrgName: "task-org"})
	require.Nil(t, err)
	require.Len(t, tasks, 2)

	tasks, err = tasksAPI.FindTasks(ctx, &api.TaskFilter{OrgName: "my-org"})
	require.Nil(t, err)
	require.Len(t, tasks, 0)

	tasks, err = tasksAPI.FindTasks(ctx, &api.TaskFilter{OrgID: *taskOrg.Id})
	require.Nil(t, err)
	require.Len(t, tasks, 2)

	tasks, err = tasksAPI.FindTasks(ctx, &api.TaskFilter{User: *myuser.Id})
	require.Nil(t, err)
	require.Len(t, tasks, 2)

	tasks, err = tasksAPI.FindTasks(ctx, &api.TaskFilter{Status: domain.TaskStatusTypeActive})
	require.Nil(t, err)
	require.Len(t, tasks, 2)

	task, err := tasksAPI.GetTask(ctx, task1)
	require.Nil(t, err)
	require.NotNil(t, task)

	err = tasksAPI.DeleteTask(ctx, task1)
	assert.Nil(t, err)

	err = tasksAPI.DeleteTask(ctx, task2)
	assert.Nil(t, err)

	err = client.OrganizationsAPI().DeleteOrganization(ctx, taskOrg)
	assert.Nil(t, err)
}

func TestTasksAPI_MembersOwners(t *testing.T) {
	client := influxdb2.NewClient(serverURL, authToken)
	tasksAPI := client.TasksAPI()
	ctx := context.Background()

	tasks, err := tasksAPI.FindTasks(ctx, nil)
	require.Nil(t, err)
	require.Len(t, tasks, 0)

	org, err := client.OrganizationsAPI().FindOrganizationByName(ctx, "my-org")
	require.Nil(t, err)
	require.NotNil(t, org)

	// Test owners
	userOwner, err := client.UsersAPI().CreateUserWithName(ctx, "bucket-owner")
	require.Nil(t, err, err)
	require.NotNil(t, userOwner)

	task, err := tasksAPI.CreateTaskWithEvery(ctx, "task 01", taskFlux, "1h", *org.Id)
	require.Nil(t, err)
	require.NotNil(t, task)

	owners, err := tasksAPI.FindOwners(ctx, task)
	require.Nil(t, err, err)
	require.NotNil(t, owners)
	assert.Len(t, owners, 0)

	owner, err := tasksAPI.AddOwner(ctx, task, userOwner)
	require.Nil(t, err, err)
	require.NotNil(t, owner)
	assert.Equal(t, *userOwner.Id, *owner.Id)

	owners, err = tasksAPI.FindOwners(ctx, task)
	require.Nil(t, err, err)
	require.NotNil(t, owners)
	assert.Len(t, owners, 1)

	err = tasksAPI.RemoveOwnerWithID(ctx, task.Id, *owners[0].Id)
	require.Nil(t, err, err)

	owners, err = tasksAPI.FindOwners(ctx, task)
	require.Nil(t, err, err)
	require.NotNil(t, owners)
	assert.Len(t, owners, 0)

	// Test members
	userMember, err := client.UsersAPI().CreateUserWithName(ctx, "bucket-member")
	require.Nil(t, err, err)
	require.NotNil(t, userMember)

	members, err := tasksAPI.FindMembers(ctx, task)
	require.Nil(t, err, err)
	require.NotNil(t, members)
	assert.Len(t, members, 0)

	member, err := tasksAPI.AddMember(ctx, task, userMember)
	require.Nil(t, err, err)
	require.NotNil(t, member)
	assert.Equal(t, *userMember.Id, *member.Id)

	members, err = tasksAPI.FindMembers(ctx, task)
	require.Nil(t, err, err)
	require.NotNil(t, members)
	assert.Len(t, members, 1)

	err = tasksAPI.RemoveMemberWithID(ctx, task.Id, *members[0].Id)
	require.Nil(t, err, err)

	members, err = tasksAPI.FindMembers(ctx, task)
	require.Nil(t, err, err)
	require.NotNil(t, members)
	assert.Len(t, members, 0)

	err = tasksAPI.DeleteTask(ctx, task)
	assert.Nil(t, err, err)

	err = client.UsersAPI().DeleteUser(ctx, userOwner)
	assert.Nil(t, err, err)

	err = client.UsersAPI().DeleteUser(ctx, userMember)
	assert.Nil(t, err, err)
}

func TestTasksAPI_Labels(t *testing.T) {
	client := influxdb2.NewClient(serverURL, authToken)
	tasksAPI := client.TasksAPI()
	labelsAPI := client.LabelsAPI()
	ctx := context.Background()

	tasks, err := tasksAPI.FindTasks(ctx, nil)
	require.Nil(t, err)
	require.Len(t, tasks, 0)

	org, err := client.OrganizationsAPI().FindOrganizationByName(ctx, "my-org")
	require.Nil(t, err)
	require.NotNil(t, org)

	task, err := tasksAPI.CreateTaskWithEvery(ctx, "task 01", taskFlux, "1h", *org.Id)
	require.Nil(t, err)
	require.NotNil(t, task)

	labels, err := tasksAPI.FindLabels(ctx, task)
	require.Nil(t, err, err)
	require.NotNil(t, labels)
	assert.Len(t, labels, 0)

	label, err := labelsAPI.CreateLabelWithName(ctx, org, "red-label", nil)
	assert.Nil(t, err)
	assert.NotNil(t, label)

	labelx, err := tasksAPI.AddLabel(ctx, task, label)
	require.Nil(t, err, err)
	require.NotNil(t, labelx)

	labels, err = tasksAPI.FindLabels(ctx, task)
	require.Nil(t, err, err)
	require.NotNil(t, labels)
	assert.Len(t, labels, 1)

	err = tasksAPI.RemoveLabel(ctx, task, label)
	require.Nil(t, err, err)

	labels, err = tasksAPI.FindLabels(ctx, task)
	require.Nil(t, err, err)
	require.NotNil(t, labels)
	assert.Len(t, labels, 0)

	err = labelsAPI.DeleteLabel(ctx, label)
	assert.Nil(t, err, err)

	err = tasksAPI.DeleteTask(ctx, task)
	assert.Nil(t, err, err)
}

func TestTasksAPI_Runs(t *testing.T) {
	client := influxdb2.NewClient(serverURL, authToken)
	tasksAPI := client.TasksAPI()
	ctx := context.Background()

	tasks, err := tasksAPI.FindTasks(ctx, nil)
	require.Nil(t, err)
	require.Len(t, tasks, 0)

	org, err := client.OrganizationsAPI().FindOrganizationByName(ctx, "my-org")
	require.Nil(t, err)
	require.NotNil(t, org)

	start := time.Now()
	task, err := tasksAPI.CreateTaskWithEvery(ctx, "rapid task", taskFlux, "1s", *org.Id)
	require.Nil(t, err)
	require.NotNil(t, task)
	//wait for task to run 10x
	<-time.After(time.Second * 10)

	logs, err := tasksAPI.FindLogs(ctx, task)
	require.Nil(t, err)
	assert.True(t, len(logs) > 0)

	runs, err := tasksAPI.FindRuns(ctx, task, nil)
	require.Nil(t, err)
	runsCount := len(runs)
	require.True(t, runsCount > 0)

	runs, err = tasksAPI.FindRuns(ctx, task, &api.RunFilter{Limit: 5})
	require.Nil(t, err)
	require.Len(t, runs, 5)

	runs, err = tasksAPI.FindRuns(ctx, task, &api.RunFilter{Limit: 5, After: *runs[4].Id})
	require.Nil(t, err)
	require.NotNil(t, runs)
	//assert.Len(t, runs, 5) https://github.com/influxdata/influxdb/issues/13577

	runs, err = tasksAPI.FindRuns(ctx, task, &api.RunFilter{AfterTime: start, BeforeTime: start.Add(5 * time.Second)})
	require.Nil(t, err)
	//assert.Len(t, runs, 5)  https://github.com/influxdata/influxdb/issues/13577
	assert.True(t, len(runs) > 0)

	runs, err = tasksAPI.FindRuns(ctx, task, &api.RunFilter{AfterTime: start.Add(5 * time.Second)})
	require.Nil(t, err)
	//assert.Len(t, runs, 5)  https://github.com/influxdata/influxdb/issues/13577
	assert.True(t, len(runs) > 0)

	logs, err = tasksAPI.FindRunLogs(ctx, &runs[0])
	require.Nil(t, err)
	assert.True(t, len(logs) > 0)

	err = tasksAPI.DeleteTask(ctx, task)
	assert.Nil(t, err)

	task, err = tasksAPI.CreateTaskWithEvery(ctx, "task", taskFlux, "1s", *org.Id)
	require.Nil(t, err)
	require.NotNil(t, task)
	//wait for tasks to start and be running
	<-time.After(1500 * time.Millisecond)

	// we should get a running run
	runs, err = tasksAPI.FindRuns(ctx, task, nil)
	require.Nil(t, err)
	if assert.True(t, len(runs) > 0) {
		_ = tasksAPI.CancelRun(ctx, &runs[0])
	}

	runm, err := tasksAPI.RunManually(ctx, task)
	require.Nil(t, err)
	require.NotNil(t, runm)

	run, err := tasksAPI.GetRunByID(ctx, *runm.TaskID, *runm.Id)
	require.Nil(t, err)
	require.NotNil(t, run)

	run2, err := tasksAPI.RetryRun(ctx, run)
	require.Nil(t, err)
	require.NotNil(t, run2)

	err = tasksAPI.DeleteTask(ctx, task)
	assert.Nil(t, err)
}

func TestTasksAPI_Failures(t *testing.T) {
	client := influxdb2.NewClient(serverURL, authToken)
	invalidID := "000000000000000"
	notExistingID := "1000000000000000"
	tasksAPI := client.TasksAPI()
	ctx := context.Background()

	_, err := tasksAPI.GetTaskByID(ctx, invalidID)
	assert.NotNil(t, err)

	_, err = tasksAPI.GetTaskByID(ctx, notExistingID)
	assert.NotNil(t, err)

	_, err = tasksAPI.FindTasks(ctx, &api.TaskFilter{OrgID: invalidID})
	assert.NotNil(t, err)

	//empty name
	_, err = tasksAPI.CreateTaskWithEvery(ctx, "", taskFlux, "4s", invalidID)
	assert.NotNil(t, err)
	//invalid flux
	_, err = tasksAPI.CreateTaskWithEvery(ctx, "Task", "taskFlux", "4s", invalidID)
	assert.NotNil(t, err)
	//invalid every
	_, err = tasksAPI.CreateTaskWithEvery(ctx, "Task", taskFlux, "4g", invalidID)
	assert.NotNil(t, err)
	//invalid org
	_, err = tasksAPI.CreateTaskWithEvery(ctx, "Task", taskFlux, "4s", invalidID)
	assert.NotNil(t, err)
	//invalid cron
	_, err = tasksAPI.CreateTaskWithCron(ctx, "Task", taskFlux, "0 * *", invalidID)
	assert.NotNil(t, err)
	// delete with id
	err = tasksAPI.DeleteTaskWithID(ctx, notExistingID)
	assert.NotNil(t, err)

	task := &domain.Task{
		Id:   notExistingID,
		Flux: taskFlux,
		Name: "task 01",
	}

	_, err = tasksAPI.UpdateTask(ctx, task)
	assert.NotNil(t, err)

	_, err = tasksAPI.FindMembersWithID(ctx, invalidID)
	assert.NotNil(t, err)

	_, err = tasksAPI.AddMemberWithID(ctx, notExistingID, invalidID)
	assert.NotNil(t, err)

	err = tasksAPI.RemoveMemberWithID(ctx, notExistingID, invalidID)
	assert.NotNil(t, err)

	_, err = tasksAPI.FindOwnersWithID(ctx, invalidID)
	assert.NotNil(t, err)

	_, err = tasksAPI.AddOwnerWithID(ctx, notExistingID, invalidID)
	assert.NotNil(t, err)

	err = tasksAPI.RemoveOwnerWithID(ctx, notExistingID, invalidID)
	assert.NotNil(t, err)

	_, err = tasksAPI.FindRunsWithID(ctx, notExistingID, nil)
	assert.NotNil(t, err)

	_, err = tasksAPI.GetRunByID(ctx, notExistingID, invalidID)
	assert.NotNil(t, err)

	_, err = tasksAPI.FindRunLogsWithID(ctx, notExistingID, invalidID)
	assert.NotNil(t, err)

	_, err = tasksAPI.RunManuallyWithID(ctx, notExistingID)
	assert.NotNil(t, err)

	_, err = tasksAPI.RetryRunWithID(ctx, notExistingID, invalidID)
	assert.NotNil(t, err)

	err = tasksAPI.CancelRunWithID(ctx, notExistingID, invalidID)
	assert.NotNil(t, err)

	_, err = tasksAPI.FindLogsWithID(ctx, notExistingID)
	assert.NotNil(t, err)

	_, err = tasksAPI.FindLabelsWithID(ctx, invalidID)
	assert.NotNil(t, err)

	_, err = tasksAPI.AddLabelWithID(ctx, notExistingID, invalidID)
	assert.NotNil(t, err)

	err = tasksAPI.RemoveLabelWithID(ctx, notExistingID, invalidID)
	assert.NotNil(t, err)

}

func TestTasksAPI_RequestFailures(t *testing.T) {
	client := influxdb2.NewClient("serverURL", authToken)
	invalidID := "000000000000000"
	notExistingID := "1000000000000000"
	tasksAPI := client.TasksAPI()
	ctx := context.Background()

	_, err := tasksAPI.GetTaskByID(ctx, invalidID)
	assert.NotNil(t, err)

	_, err = tasksAPI.GetTaskByID(ctx, notExistingID)
	assert.NotNil(t, err)

	_, err = tasksAPI.FindTasks(ctx, nil)
	assert.NotNil(t, err)

	//empty name
	_, err = tasksAPI.CreateTaskWithEvery(ctx, "", taskFlux, "4s", invalidID)
	assert.NotNil(t, err)
	//invalid flux
	_, err = tasksAPI.CreateTaskWithEvery(ctx, "Task", "taskFlux", "4s", invalidID)
	assert.NotNil(t, err)
	//invalid every
	_, err = tasksAPI.CreateTaskWithEvery(ctx, "Task", taskFlux, "4g", invalidID)
	assert.NotNil(t, err)
	//invalid org
	_, err = tasksAPI.CreateTaskWithEvery(ctx, "Task", taskFlux, "4s", invalidID)
	assert.NotNil(t, err)
	//invalid cron
	_, err = tasksAPI.CreateTaskWithCron(ctx, "Task", taskFlux, "0 * *", invalidID)
	assert.NotNil(t, err)
	// delete with id
	err = tasksAPI.DeleteTaskWithID(ctx, notExistingID)
	assert.NotNil(t, err)

	task := &domain.Task{
		Id:   notExistingID,
		Flux: taskFlux,
		Name: "task 01",
	}

	_, err = tasksAPI.UpdateTask(ctx, task)
	assert.NotNil(t, err)

	_, err = tasksAPI.FindMembersWithID(ctx, invalidID)
	assert.NotNil(t, err)

	_, err = tasksAPI.AddMemberWithID(ctx, notExistingID, invalidID)
	assert.NotNil(t, err)

	err = tasksAPI.RemoveMemberWithID(ctx, notExistingID, invalidID)
	assert.NotNil(t, err)

	_, err = tasksAPI.FindOwnersWithID(ctx, invalidID)
	assert.NotNil(t, err)

	_, err = tasksAPI.AddOwnerWithID(ctx, notExistingID, invalidID)
	assert.NotNil(t, err)

	err = tasksAPI.RemoveOwnerWithID(ctx, notExistingID, invalidID)
	assert.NotNil(t, err)

	_, err = tasksAPI.FindRunsWithID(ctx, notExistingID, nil)
	assert.NotNil(t, err)

	_, err = tasksAPI.GetRunByID(ctx, notExistingID, invalidID)
	assert.NotNil(t, err)

	_, err = tasksAPI.FindRunLogsWithID(ctx, notExistingID, invalidID)
	assert.NotNil(t, err)

	_, err = tasksAPI.RunManuallyWithID(ctx, notExistingID)
	assert.NotNil(t, err)

	_, err = tasksAPI.RetryRunWithID(ctx, notExistingID, invalidID)
	assert.NotNil(t, err)

	err = tasksAPI.CancelRunWithID(ctx, notExistingID, invalidID)
	assert.NotNil(t, err)

	_, err = tasksAPI.FindLogsWithID(ctx, notExistingID)
	assert.NotNil(t, err)

	_, err = tasksAPI.FindLabelsWithID(ctx, invalidID)
	assert.NotNil(t, err)

	_, err = tasksAPI.AddLabelWithID(ctx, notExistingID, invalidID)
	assert.NotNil(t, err)

	err = tasksAPI.RemoveLabelWithID(ctx, notExistingID, invalidID)
	assert.NotNil(t, err)

}
