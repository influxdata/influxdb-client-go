//go:build e2e
// +build e2e

// Copyright 2021 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package influxclient_test

import (
	"fmt"
	"testing"
	"time"

	. "github.com/influxdata/influxdb-client-go/influxclient"
	"github.com/influxdata/influxdb-client-go/influxclient/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const taskFluxTemplate = `from(bucket: "%s") |> range(start: -task.every) |> last()`

func TestTasksAPI_CRUDTask(t *testing.T) {
	client, ctx := newClient(t)
	tasksAPI := client.TasksAPI()

	tasks, err := tasksAPI.Find(ctx, nil)
	require.NoError(t, err)
	require.NotNil(t, tasks)
	require.Len(t, tasks, 0)

	org, err := client.OrganizationAPI().FindOne(ctx, &Filter{
		OrgName: orgName,
	})
	require.NoError(t, err)
	require.NotNil(t, org)

	taskDescription := "Example task"
	taskStatus := model.TaskStatusTypeInactive
	taskEvery := "5s"
	taskFlux := fmt.Sprintf(taskFluxTemplate, bucketName)
	newTask := &model.Task{
		Description: &taskDescription,
		Every:       &taskEvery,
		Flux:        taskFlux,
		Name:        "task 01",
		OrgID:       *org.Id,
		Status:      &taskStatus,
	}

	task1, err := tasksAPI.Create(ctx, newTask)
	require.NoError(t, err)
	require.NotNil(t, task1)
	defer tasksAPI.Delete(ctx, safeId(task1.Id))

	assert.Equal(t, "task 01", task1.Name, task1.Name)
	if assert.NotNil(t, task1.Description) {
		assert.Equal(t, taskDescription, *task1.Description)
	}
	if assert.NotNil(t, task1.Every) {
		assert.Equal(t, "5s", *task1.Every)
	}
	if assert.NotNil(t, task1.Status) {
		assert.Equal(t, taskStatus, *task1.Status)
	}
	assert.Equal(t, *org.Id, task1.OrgID, task1.OrgID)

	task2Every := "1h"
	task2, err := tasksAPI.Create(ctx, &model.Task{
		Name:  "task 02",
		Every: &task2Every,
		Flux:  taskFlux,
		OrgID: *org.Id,
	})
	require.NoError(t, err)
	require.NotNil(t, task2)
	defer tasksAPI.Delete(ctx, safeId(task2.Id))

	assert.Equal(t, "task 02", task2.Name)
	assert.Nil(t, task2.Description)
	if assert.NotNil(t, task2.Every) {
		assert.Equal(t, "1h", *task2.Every)
	}
	if assert.NotNil(t, task2.Status) {
		assert.Equal(t, model.TaskStatusTypeActive, *task2.Status)
	}
	assert.Equal(t, *org.Id, task2.OrgID, task2.OrgID)

	task3Cron := "*/1 * * * *"
	task3, err := tasksAPI.Create(ctx, &model.Task{
		Name:  "task 03",
		Cron:  &task3Cron,
		Flux:  taskFlux,
		OrgID: *org.Id,
	})
	require.NoError(t, err)
	require.NotNil(t, task3)
	defer tasksAPI.Delete(ctx, safeId(task3.Id))

	assert.Equal(t, "task 03", task3.Name, task3.Name)
	assert.Nil(t, task3.Description)
	if assert.NotNil(t, task3.Cron) {
		assert.Equal(t, "*/1 * * * *", *task3.Cron)
	}
	if assert.NotNil(t, task3.Status) {
		assert.Equal(t, model.TaskStatusTypeActive, *task3.Status)
	}
	assert.Equal(t, *org.Id, task3.OrgID, task3.OrgID)

	tasks, err = tasksAPI.Find(ctx, nil)
	require.NoError(t, err)
	assert.Len(t, tasks, 3)

	tasks, err = tasksAPI.Find(ctx, &Filter{})
	require.NoError(t, err)
	assert.Len(t, tasks, 3)

	task3.Every = &taskEvery
	task3.Description = &taskDescription
	task3.Status = &taskStatus
	task3.Name = "task 03x"
	task, err := tasksAPI.Update(ctx, task3)
	require.NoError(t, err)
	require.NotNil(t, task)

	assert.Equal(t, "task 03x", task.Name, task.Name)
	if assert.NotNil(t, task.Description) {
		assert.Equal(t, taskDescription, *task.Description)
	}
	if assert.NotNil(t, task.Every) {
		assert.Equal(t, taskEvery, *task3.Every)
	}
	if assert.NotNil(t, task3.Status) {
		assert.Equal(t, taskStatus, *task3.Status)
	}

	flux := fmt.Sprintf(`import "types"
option task = { 
  name: "task 04",
  every: 1h,
}

from(bucket: "%s")
    |> range(start: -task.every)
    |> filter(fn: (r) => r._measurement == "mem" and r.host == "myHost")`, bucketName)
	//task4, err := tasksAPI.CreateTaskByFlux(ctx, flux, *org.Id)
	task4, err := tasksAPI.Create(ctx, &model.Task{
		OrgID: *org.Id,
		Flux:  flux,
	})
	require.NoError(t, err)
	require.NotNil(t, task4)
	defer tasksAPI.Delete(ctx, safeId(task4.Id))

	assert.Equal(t, "task 04", task4.Name, task4.Name)
	assert.Nil(t, task4.Description)
	if assert.NotNil(t, task4.Every) {
		assert.Equal(t, "1h", *task4.Every)
	}
	if assert.NotNil(t, task4.Status) {
		assert.Equal(t, model.TaskStatusTypeActive, *task4.Status)
	}
	assert.Equal(t, *org.Id, task4.OrgID)

	err = tasksAPI.Delete(ctx, task1.Id)
	assert.NoError(t, err)

	err = tasksAPI.Delete(ctx, task2.Id)
	assert.NoError(t, err)

	err = tasksAPI.Delete(ctx, task3.Id)
	assert.NoError(t, err)

	err = tasksAPI.Delete(ctx, task4.Id)
	assert.NoError(t, err)

	tasks, err = tasksAPI.Find(ctx, nil)
	require.NoError(t, err)
	assert.Len(t, tasks, 0)
}

func TestTasksAPI_GetTasks(t *testing.T) {
	client, ctx := newClient(t)
	tasksAPI := client.TasksAPI()

	tasks, err := tasksAPI.Find(ctx, nil)
	require.NoError(t, err)
	require.Len(t, tasks, 0)

	taskOrg, err := client.OrganizationAPI().Create(ctx, &model.Organization{
		Name: "task-org",
	})
	require.NoError(t, err)
	require.NotNil(t, taskOrg)
	defer client.OrganizationAPI().Delete(ctx, safeId(taskOrg.Id))

	taskEvery := "1h"
	taskFlux := fmt.Sprintf(taskFluxTemplate, bucketName)
	newtasks := make([]*model.Task, 30)
	for i := 0; i < 30; i++ {
		newtasks[i], err = tasksAPI.Create(ctx, &model.Task{
			Name:  fmt.Sprintf("task %02d", i+1),
			Every: &taskEvery,
			Flux:  taskFlux,
			OrgID: *taskOrg.Id,
		})
		require.NoError(t, err)
		require.NotNil(t, newtasks[i])
		defer tasksAPI.Delete(ctx, safeId(newtasks[i].Id))
	}

	tasks, err = tasksAPI.Find(ctx, &Filter{
		Limit: 15,
	})
	require.NoError(t, err)
	assert.Len(t, tasks, 15)

	tasks, err = tasksAPI.Find(ctx, &Filter{
		After: tasks[14].Id,
	})
	require.NoError(t, err)
	assert.Len(t, tasks, 15)

	tasks, err = tasksAPI.Find(ctx, &Filter{
		Limit: 100,
	})
	require.NoError(t, err)
	assert.Len(t, tasks, 30)

	for i := 0; i < 30; i++ {
		err = tasksAPI.Delete(ctx, newtasks[i].Id)
		assert.NoError(t, err)
	}

	err = client.OrganizationAPI().Delete(ctx, *taskOrg.Id)
	assert.NoError(t, err)

}

func TestTasksAPI_FindTasks(t *testing.T) {
	client, ctx := newClient(t)
	tasksAPI := client.TasksAPI()

	tasks, err := tasksAPI.Find(ctx, nil)
	require.NoError(t, err)
	require.Len(t, tasks, 0)

	taskOrg, err := client.OrganizationAPI().Create(ctx, &model.Organization{
		Name: "task-org",
	})
	require.NoError(t, err)
	require.NotNil(t, taskOrg)
	defer client.OrganizationAPI().Delete(ctx, safeId(taskOrg.Id))

	user, err := client.UsersAPI().FindOne(ctx, &Filter{
		Name: userName,
	})
	require.NoError(t, err)
	require.NotNil(t, user)

	taskEvery := "1h"
	taskFlux := fmt.Sprintf(taskFluxTemplate, bucketName)
	task1, err := tasksAPI.Create(ctx, &model.Task{
		Name:  "task 01",
		Flux:  taskFlux,
		Every: &taskEvery,
		OrgID: *taskOrg.Id,
	})
	require.NoError(t, err)
	require.NotNil(t, task1)
	defer tasksAPI.Delete(ctx, safeId(task1.Id))

	tasks, err = tasksAPI.Find(ctx, &Filter{
		Name: task1.Name,
	})
	require.NoError(t, err)
	assert.Len(t, tasks, 1)

	task2Every := "1m"
	task2, err := tasksAPI.Create(ctx, &model.Task{
		Name:  "task 01",
		Flux:  taskFlux,
		Every: &task2Every,
		OrgID: *taskOrg.Id,
	})
	require.NoError(t, err)
	require.NotNil(t, task2)
	defer tasksAPI.Delete(ctx, safeId(task2.Id))

	tasks, err = tasksAPI.Find(ctx, &Filter{
		Name: "task 01",
	})
	require.NoError(t, err)
	assert.Len(t, tasks, 2)

	tasks, err = tasksAPI.Find(ctx, &Filter{
		OrgName: "task-org",
	})
	require.NoError(t, err)
	assert.Len(t, tasks, 2)

	tasks, err = tasksAPI.Find(ctx, &Filter{
		OrgName: orgName,
	})
	require.NoError(t, err)
	assert.Len(t, tasks, 0)

	tasks, err = tasksAPI.Find(ctx, &Filter{
		OrgID: *taskOrg.Id,
	})
	require.NoError(t, err)
	assert.Len(t, tasks, 2)

	tasks, err = tasksAPI.Find(ctx, &Filter{
		UserName: *user.Id,
	})
	require.NoError(t, err)
	assert.Len(t, tasks, 2) // TODO how come???

	tasks, err = tasksAPI.Find(ctx, &Filter{
		Status: string(model.TaskStatusTypeActive),
	})
	require.NoError(t, err)
	assert.Len(t, tasks, 2)

	task, err := tasksAPI.FindOne(ctx, &Filter{
		ID: task1.Id,
	})
	require.NoError(t, err)
	assert.NotNil(t, task)

	err = tasksAPI.Delete(ctx, task1.Id)
	assert.NoError(t, err)

	err = tasksAPI.Delete(ctx, task2.Id)
	assert.NoError(t, err)

	err = client.OrganizationAPI().Delete(ctx, *taskOrg.Id)
	assert.NoError(t, err)
}

func TestTasksAPI_MembersOwners(t *testing.T) {
	client, ctx := newClient(t)
	tasksAPI := client.TasksAPI()

	tasks, err := tasksAPI.Find(ctx, nil)
	require.NoError(t, err)
	require.Len(t, tasks, 0)

	org, err := client.OrganizationAPI().FindOne(ctx, &Filter{
		Name: orgName,
	})
	require.NoError(t, err)
	require.NotNil(t, org)

	// Test owners
	userOwner, err := client.UsersAPI().Create(ctx, &model.User{
		Name: "bucket-owner",
	})
	require.Nil(t, err, err)
	require.NotNil(t, userOwner)
	defer client.UsersAPI().Delete(ctx, safeId(userOwner.Id))

	taskEvery := "1h"
	taskFlux := fmt.Sprintf(taskFluxTemplate, bucketName)
	task, err := tasksAPI.Create(ctx, &model.Task{
		Name:  "task 01",
		Flux:  taskFlux,
		Every: &taskEvery,
		OrgID: *org.Id,
	})
	require.NoError(t, err)
	require.NotNil(t, task)
	defer tasksAPI.Delete(ctx, safeId(task.Id))

	owners, err := tasksAPI.Owners(ctx, task.Id)
	require.Nil(t, err, err)
	require.NotNil(t, owners)
	assert.Len(t, owners, 0)

	/*owner, */
	err = tasksAPI.AddOwner(ctx, task.Id, *userOwner.Id)
	require.Nil(t, err, err)
	/*require.NotNil(t, owner)
	assert.Equal(t, *userOwner.Id, *owner.Id)*/

	owners, err = tasksAPI.Owners(ctx, task.Id)
	require.Nil(t, err, err)
	require.NotNil(t, owners)
	assert.Len(t, owners, 1)

	err = tasksAPI.RemoveOwner(ctx, task.Id, *owners[0].Id)
	require.Nil(t, err, err)

	owners, err = tasksAPI.Owners(ctx, task.Id)
	require.Nil(t, err, err)
	require.NotNil(t, owners)
	assert.Len(t, owners, 0)

	// Test members
	userMember, err := client.UsersAPI().Create(ctx, &model.User{
		Name: "bucket-member",
	})
	require.Nil(t, err, err)
	require.NotNil(t, userMember)
	defer client.UsersAPI().Delete(ctx, safeId(userMember.Id))

	members, err := tasksAPI.Members(ctx, task.Id)
	require.Nil(t, err, err)
	require.NotNil(t, members)
	assert.Len(t, members, 0)

	/*member, */
	err = tasksAPI.AddMember(ctx, task.Id, *userMember.Id)
	require.Nil(t, err, err)
	/*require.NotNil(t, member)
	assert.Equal(t, *userMember.Id, *member.Id)*/

	members, err = tasksAPI.Members(ctx, task.Id)
	require.Nil(t, err, err)
	require.NotNil(t, members)
	assert.Len(t, members, 1)

	err = tasksAPI.RemoveMember(ctx, task.Id, *members[0].Id)
	require.Nil(t, err, err)

	members, err = tasksAPI.Members(ctx, task.Id)
	require.Nil(t, err, err)
	require.NotNil(t, members)
	assert.Len(t, members, 0)

	err = tasksAPI.Delete(ctx, task.Id)
	assert.Nil(t, err, err)

	err = client.UsersAPI().Delete(ctx, *userOwner.Id)
	assert.Nil(t, err, err)

	err = client.UsersAPI().Delete(ctx, *userMember.Id)
	assert.Nil(t, err, err)
}

func TestTasksAPI_Labels(t *testing.T) {
	client, ctx := newClient(t)
	tasksAPI := client.TasksAPI()
	labelsAPI := client.LabelsAPI()

	tasks, err := tasksAPI.Find(ctx, nil)
	require.NoError(t, err)
	require.Len(t, tasks, 0)

	org, err := client.OrganizationAPI().FindOne(ctx, &Filter{
		Name: orgName,
	})
	require.NoError(t, err)
	require.NotNil(t, org)

	taskEvery := "1h"
	taskFlux := fmt.Sprintf(taskFluxTemplate, bucketName)
	task, err := tasksAPI.Create(ctx, &model.Task{
		Name:  "task 01",
		Flux:  taskFlux,
		Every: &taskEvery,
		OrgID: *org.Id,
	})
	require.NoError(t, err)
	require.NotNil(t, task)
	defer tasksAPI.Delete(ctx, safeId(task.Id))

	labels, err := tasksAPI.FindLabels(ctx, task.Id)
	require.Nil(t, err, err)
	require.NotNil(t, labels)
	assert.Len(t, labels, 0)

	labelName := "red-label"
	label, err := labelsAPI.Create(ctx, &model.Label{
		OrgID: org.Id,
		Name:  &labelName,
	})
	assert.NoError(t, err)
	assert.NotNil(t, label)
	defer labelsAPI.Delete(ctx, safeId(label.Id))

	labelx, err := tasksAPI.AddLabel(ctx, task.Id, *label.Id)
	require.Nil(t, err, err)
	require.NotNil(t, labelx)

	labels, err = tasksAPI.FindLabels(ctx, task.Id)
	require.Nil(t, err, err)
	require.NotNil(t, labels)
	assert.Len(t, labels, 1)

	err = tasksAPI.RemoveLabel(ctx, task.Id, *label.Id)
	require.Nil(t, err, err)

	labels, err = tasksAPI.FindLabels(ctx, task.Id)
	require.Nil(t, err, err)
	require.NotNil(t, labels)
	assert.Len(t, labels, 0)

	err = labelsAPI.Delete(ctx, *label.Id)
	assert.Nil(t, err, err)

	err = tasksAPI.Delete(ctx, task.Id)
	assert.Nil(t, err, err)
}

func TestTasksAPI_Runs(t *testing.T) {
	client, ctx := newClient(t)
	tasksAPI := client.TasksAPI()

	tasks, err := tasksAPI.Find(ctx, nil)
	require.NoError(t, err)
	require.Len(t, tasks, 0)

	org, err := client.OrganizationAPI().FindOne(ctx, &Filter{
		Name: orgName,
	})
	require.NoError(t, err)
	require.NotNil(t, org)

	start := time.Now()
	taskEvery := "1s"
	taskFlux := fmt.Sprintf(taskFluxTemplate, bucketName)
	task, err := tasksAPI.Create(ctx, &model.Task{
		Name:  "rapid task",
		Flux:  taskFlux,
		Every: &taskEvery,
		OrgID: *org.Id,
	})
	require.NoError(t, err)
	require.NotNil(t, task)
	defer tasksAPI.Delete(ctx, safeId(task.Id))
	// wait for task to run 10x
	<-time.After(time.Second*10 + time.Second*1 /*give it a little more time to finish at least 10 runs*/)

	logs, err := tasksAPI.FindLogs(ctx, task.Id)
	require.NoError(t, err)
	assert.True(t, len(logs) > 0)

	runs, err := tasksAPI.FindRuns(ctx, task.Id, nil)
	require.NoError(t, err)
	runsCount := len(runs)
	assert.True(t, runsCount > 0)

	runs, err = tasksAPI.FindRuns(ctx, task.Id, &Filter{
		Limit: 5,
	})
	require.NoError(t, err)
	assert.Len(t, runs, 5)

	runs, err = tasksAPI.FindRuns(ctx, task.Id, &Filter{
		Limit: 5,
		After: *runs[4].Id,
	})
	require.NoError(t, err)
	require.NotNil(t, runs)
	//assert.Len(t, runs, 5) // https://github.com/influxdata/influxdb/issues/13577

	runs, err = tasksAPI.FindRuns(ctx, task.Id, &Filter{
		AfterTime:  start,
		BeforeTime: start.Add(5 * time.Second),
	})
	require.NoError(t, err)
	//assert.Len(t, runs, 5) // https://github.com/influxdata/influxdb/issues/13577
	assert.True(t, len(runs) > 0)

	runs, err = tasksAPI.FindRuns(ctx, task.Id, &Filter{
		AfterTime: start.Add(5 * time.Second),
	})
	require.NoError(t, err)
	//assert.Len(t, runs, 5) // https://github.com/influxdata/influxdb/issues/13577
	assert.True(t, len(runs) > 0)

	logs, err = tasksAPI.FindRunLogs(ctx, task.Id, *runs[0].Id)
	require.NoError(t, err)
	assert.True(t, len(logs) > 0)

	err = tasksAPI.Delete(ctx, task.Id)
	assert.NoError(t, err)

	task, err = tasksAPI.Create(ctx, &model.Task{
		Name:  "task",
		Flux:  taskFlux,
		Every: &taskEvery,
		OrgID: *org.Id,
	})
	require.NoError(t, err)
	require.NotNil(t, task)
	defer tasksAPI.Delete(ctx, safeId(task.Id))
	//wait for tasks to start and be running
	<-time.After(1500 * time.Millisecond)

	// we should get a running run
	runs, err = tasksAPI.FindRuns(ctx, task.Id, nil)
	require.NoError(t, err)
	if assert.True(t, len(runs) > 0) {
		_ = tasksAPI.CancelRun(ctx, task.Id, *runs[0].Id)
	}

	runm, err := tasksAPI.RunManually(ctx, task.Id)
	require.NoError(t, err)
	require.NotNil(t, runm)

	run, err := tasksAPI.FindOneRun(ctx, *runm.TaskID, &Filter{ID: *runm.Id})
	require.NoError(t, err)
	require.NotNil(t, run)

	run2, err := tasksAPI.RetryRun(ctx, task.Id, *run.Id)
	require.NoError(t, err)
	require.NotNil(t, run2)

	err = tasksAPI.Delete(ctx, task.Id)
	assert.NoError(t, err)
}

func TestTasksAPI_Failures(t *testing.T) {
	client, ctx := newClient(t)
	tasksAPI := client.TasksAPI()

	taskFlux := fmt.Sprintf(taskFluxTemplate, bucketName)

	_, err := tasksAPI.Find(ctx, &Filter{ID: invalidID})
	assert.Error(t, err)

	_, err = tasksAPI.Find(ctx, &Filter{ID: notExistingID})
	assert.Error(t, err)

	_, err = tasksAPI.Find(ctx, &Filter{OrgID: invalidID})
	assert.Error(t, err)

	org, err := client.OrganizationAPI().FindOne(ctx, &Filter{
		Name: orgName,
	})
	require.NoError(t, err)
	require.NotNil(t, org)

	_, err = tasksAPI.Create(ctx, nil)
	assert.Error(t, err)

	_, err = tasksAPI.Create(ctx, &model.Task{})
	assert.Error(t, err)

	// empty org / orgId
	every := "4s"
	_, err = tasksAPI.Create(ctx, &model.Task{Name: "", Flux: taskFlux, Every: &every})
	assert.Error(t, err)

	// empty name
	_, err = tasksAPI.Create(ctx, &model.Task{Name: "", Flux: taskFlux, Every: &every, OrgID: *org.Id})
	assert.Error(t, err)

	// invalid flux
	_, err = tasksAPI.Create(ctx, &model.Task{Name: "Task", Flux: "x := null", Every: &every, OrgID: *org.Id})
	assert.Error(t, err)

	// empty flux
	_, err = tasksAPI.Create(ctx, &model.Task{Name: "Task", Flux: "", Every: &every, OrgID: *org.Id})
	assert.Error(t, err)

	// invalid every
	every = "4g"
	_, err = tasksAPI.Create(ctx, &model.Task{Name: "Task", Flux: taskFlux, Every: &every, OrgID: *org.Id})
	assert.Error(t, err)

	// invalid org
	every = "4s"
	_, err = tasksAPI.Create(ctx, &model.Task{Name: "Task", Flux: taskFlux, Every: &every, OrgID: invalidID})
	assert.Error(t, err)

	// invalid cron
	cron := "0 * *"
	_, err = tasksAPI.Create(ctx, &model.Task{Name: "Task", Flux: taskFlux, Cron: &cron, OrgID: *org.Id})
	assert.Error(t, err)

	_, err = tasksAPI.Update(ctx, nil)
	assert.Error(t, err)

	_, err = tasksAPI.Update(ctx, &model.Task{
		Id:   notExistingID,
		Flux: taskFlux,
		Name: "task 01",
	})
	assert.Error(t, err)

	_, err = tasksAPI.Update(ctx, &model.Task{
		Id:   "",
		Flux: taskFlux,
		Name: "task 01",
	})
	assert.Error(t, err)

	// delete with id
	err = tasksAPI.Delete(ctx, notExistingID)
	assert.Error(t, err)

	err = tasksAPI.Delete(ctx, notInitializedID)
	assert.Error(t, err)

	_, err = tasksAPI.Members(ctx, invalidID)
	assert.Error(t, err)

	_, err = tasksAPI.Members(ctx, notInitializedID)
	assert.Error(t, err)

	/*_,*/
	err = tasksAPI.AddMember(ctx, notExistingID, invalidID)
	assert.Error(t, err)

	/*_,*/
	err = tasksAPI.AddMember(ctx, notExistingID, notInitializedID)
	assert.Error(t, err)

	/*_,*/
	err = tasksAPI.AddMember(ctx, notInitializedID, notExistingID)
	assert.Error(t, err)

	err = tasksAPI.RemoveMember(ctx, notExistingID, invalidID)
	assert.Error(t, err)

	err = tasksAPI.RemoveMember(ctx, notInitializedID, notExistingID)
	assert.Error(t, err)

	err = tasksAPI.RemoveMember(ctx, notExistingID, notInitializedID)
	assert.Error(t, err)

	_, err = tasksAPI.Owners(ctx, invalidID)
	assert.Error(t, err)

	_, err = tasksAPI.Owners(ctx, notInitializedID)
	assert.Error(t, err)

	/*_,*/
	err = tasksAPI.AddOwner(ctx, notExistingID, invalidID)
	assert.Error(t, err)

	/*_,*/
	err = tasksAPI.AddOwner(ctx, notExistingID, notInitializedID)
	assert.Error(t, err)

	/*_,*/
	err = tasksAPI.AddOwner(ctx, notInitializedID, notExistingID)
	assert.Error(t, err)

	err = tasksAPI.RemoveOwner(ctx, notExistingID, invalidID)
	assert.Error(t, err)

	err = tasksAPI.RemoveOwner(ctx, notExistingID, notInitializedID)
	assert.Error(t, err)

	err = tasksAPI.RemoveOwner(ctx, notInitializedID, notExistingID)
	assert.Error(t, err)

	_, err = tasksAPI.FindRuns(ctx, notExistingID, nil)
	assert.Error(t, err)

	_, err = tasksAPI.FindRuns(ctx, notInitializedID, nil)
	assert.Error(t, err)

	_, err = tasksAPI.FindOneRun(ctx, notInitializedID, &Filter{ID: invalidID})
	assert.Error(t, err)

	_, err = tasksAPI.FindOneRun(ctx, notExistingID, nil)
	assert.Error(t, err)

	_, err = tasksAPI.FindOneRun(ctx, notExistingID, &Filter{ID: invalidID})
	assert.Error(t, err)

	_, err = tasksAPI.FindOneRun(ctx, notExistingID, &Filter{ID: notInitializedID})
	assert.Error(t, err)

	_, err = tasksAPI.FindRunLogs(ctx, notExistingID, invalidID)
	assert.Error(t, err)

	_, err = tasksAPI.FindRunLogs(ctx, notInitializedID, invalidID)
	assert.Error(t, err)

	_, err = tasksAPI.FindRunLogs(ctx, invalidID, notInitializedID)
	assert.Error(t, err)

	_, err = tasksAPI.RunManually(ctx, notExistingID)
	assert.Error(t, err)

	_, err = tasksAPI.RunManually(ctx, notInitializedID)
	assert.Error(t, err)

	_, err = tasksAPI.RetryRun(ctx, notExistingID, invalidID)
	assert.Error(t, err)

	_, err = tasksAPI.RetryRun(ctx, notExistingID, notInitializedID)
	assert.Error(t, err)

	_, err = tasksAPI.RetryRun(ctx, notInitializedID, notExistingID)
	assert.Error(t, err)

	err = tasksAPI.CancelRun(ctx, notExistingID, invalidID)
	assert.Error(t, err)

	err = tasksAPI.CancelRun(ctx, notInitializedID, notExistingID)
	assert.Error(t, err)

	err = tasksAPI.CancelRun(ctx, notExistingID, notInitializedID)
	assert.Error(t, err)

	_, err = tasksAPI.FindLogs(ctx, notExistingID)
	assert.Error(t, err)

	_, err = tasksAPI.FindLogs(ctx, notInitializedID)
	assert.Error(t, err)

	_, err = tasksAPI.FindLabels(ctx, notInitializedID)
	assert.Error(t, err)

	_, err = tasksAPI.FindLabels(ctx, invalidID)
	assert.Error(t, err)

	_, err = tasksAPI.AddLabel(ctx, notInitializedID, notExistingID)
	assert.Error(t, err)

	_, err = tasksAPI.AddLabel(ctx, notExistingID, notInitializedID)
	assert.Error(t, err)

	_, err = tasksAPI.AddLabel(ctx, notExistingID, invalidID)
	assert.Error(t, err)

	err = tasksAPI.RemoveLabel(ctx, notExistingID, invalidID)
	assert.Error(t, err)

	err = tasksAPI.RemoveLabel(ctx, notInitializedID, notExistingID)
	assert.Error(t, err)

	err = tasksAPI.RemoveLabel(ctx, notExistingID, notInitializedID)
	assert.Error(t, err)
}
