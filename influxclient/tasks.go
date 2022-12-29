// Copyright 2021 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

// Package influxclient provides client for InfluxDB server.
package influxclient

import (
	"context"
	"fmt"

	"github.com/influxdata/influxdb-client-go/influxclient/model"
)

// TasksAPI holds methods related to organization, as found under
// the /tasks endpoint.
type TasksAPI struct {
	client *model.Client
}

// newTasksAPI returns new TasksAPI instance
func newTasksAPI(client *model.Client) *TasksAPI {
	return &TasksAPI{client: client}
}

// Find returns all tasks matching the given filter.
// Supported filters:
//   After
//   Name
//   OrgName
//	 OrgID
//	 UserName
//   Status
//   Limit
func (a *TasksAPI) Find(ctx context.Context, filter *Filter) ([]model.Task, error) {
	return a.getTasks(ctx, filter)
}

// FindOne returns one task matching the given filter.
// Supported filters:
//   After
//   Name
//   OrgName
//	 OrgID
//	 UserName
//   Status
//   Limit
func (a *TasksAPI) FindOne(ctx context.Context, filter *Filter) (*model.Task, error){
	tasks, err := a.getTasks(ctx, filter)
	if err != nil {
		return nil, err
	}
	if len(tasks) > 0 {
		return &(tasks)[0], nil
	}
	return nil, fmt.Errorf("task not found")
}

// Create creates a new task according the the task object.
// Set OrgId, Name, Description, Flux, Status and Every or Cron properties.
// Every and Cron are mutually exclusive. Every has higher priority.
func (a *TasksAPI) Create(ctx context.Context, task *model.Task) (*model.Task, error) {
	if task == nil {
		return nil, fmt.Errorf("task cannot be nil")
	}
	if task.OrgID == "" && task.Org == nil {
		return nil, fmt.Errorf("either orgID or org is required")
	}
	if task.Flux == "" {
		return nil, fmt.Errorf("flux is required")
	}
	var repetition string
	if task.Every != nil {
		repetition = fmt.Sprintf("every: %s", *task.Every)
	} else if task.Cron != nil {
		repetition = fmt.Sprintf(`cron: "%s"`, *task.Cron)
	}
	var flux string
	if repetition != "" {
		flux = fmt.Sprintf("option task = { name: \"%s\", %s }\n\n%s", task.Name, repetition, task.Flux)
	} else {
		flux = task.Flux
	}
	params := &model.PostTasksAllParams{
		Body: model.PostTasksJSONRequestBody{
			Description: task.Description,
			Flux: flux,
			Status: task.Status,
		},
	}
	if task.OrgID != "" {
		params.Body.OrgID = &task.OrgID
	} else {
		params.Body.Org = task.Org
	}
	return a.client.PostTasks(ctx, params)
}

// Update updates a task. The task.ID field must be specified.
// The complete task information is returned.
func (a *TasksAPI) Update(ctx context.Context, task *model.Task) (*model.Task, error) {
	if task == nil {
		return nil, fmt.Errorf("task cannot be nil")
	}
	if task.Id == "" {
		return nil, fmt.Errorf("task ID is required")
	}
	params := &model.PatchTasksIDAllParams{
		TaskID: task.Id,
		Body: model.PatchTasksIDJSONRequestBody{
			Name: &task.Name,
			Description: task.Description,
			Offset: task.Offset,
			Flux: &task.Flux,
			Status: task.Status,
		},
	}
	if task.Every != nil {
		params.Body.Every = task.Every
	} else if task.Cron != nil {
		params.Body.Cron = task.Cron
	}
	return a.client.PatchTasksID(ctx, params)
}

// Delete deletes the task with the given ID.
func (a *TasksAPI) Delete(ctx context.Context, taskID string) error {
	if taskID == "" {
		return fmt.Errorf("taskID is required")
	}
	params := &model.DeleteTasksIDAllParams{
		TaskID: taskID,
	}
	return a.client.DeleteTasksID(ctx, params)
}

// FindRuns returns a task runs according the filter.
// Supported filters:
//   After
//   AfterTime
//   BeforeTime
//   Limit
func (a *TasksAPI) FindRuns(ctx context.Context, taskID string, filter *Filter) ([]model.Run, error){
	if taskID == "" {
		return nil, fmt.Errorf("taskID is required")
	}
	params := &model.GetTasksIDRunsAllParams{
		TaskID: taskID,
	}
	if filter != nil {
		if filter.After != "" {
			params.After = &filter.After
		}
		if !filter.AfterTime.IsZero() {
			params.AfterTime = &filter.AfterTime
		}
		if !filter.BeforeTime.IsZero() {
			params.BeforeTime = &filter.BeforeTime
		}
		if filter.Limit > 0 {
			iLimit := int(filter.Limit)
			params.Limit = &iLimit
		}
	}
	response, err := a.client.GetTasksIDRuns(ctx, params)
	if err != nil {
		return nil, err
	}
	return *response.Runs, nil
}

// FindOneRun returns one task run that matches the given filter.
// Supported filters:
//   ID
// TODO or just pass runID instead of a filter?
func (a *TasksAPI) FindOneRun(ctx context.Context, taskID string, filter *Filter) (*model.Run, error) {
	if taskID == "" {
		return nil, fmt.Errorf("taskID is required")
	}
	if filter == nil {
		return nil, fmt.Errorf("filter cannot be nil")
	}
	if filter.ID == "" {
		return nil, fmt.Errorf("ID is required")
	}
	params := &model.GetTasksIDRunsIDAllParams{
		TaskID: taskID,
		RunID: filter.ID,
	}
	return a.client.GetTasksIDRunsID(ctx, params)
}

// FindRunLogs return all log events for a task run with given ID.
func (a *TasksAPI) FindRunLogs(ctx context.Context, taskID, runID string) ([]model.LogEvent, error) {
	if taskID == "" {
		return nil, fmt.Errorf("taskID is required")
	}
	if runID == "" {
		return nil, fmt.Errorf("runID is required")
	}
	params := &model.GetTasksIDRunsIDLogsAllParams{
		TaskID: taskID,
		RunID: runID,
	}
	response, err := a.client.GetTasksIDRunsIDLogs(ctx, params)
	if err != nil {
		return nil, err
	}
	return *response.Events, nil
}

// RunManually manually start a run of a task with given ID now, overriding the current schedule.
func (a *TasksAPI) RunManually(ctx context.Context, taskID string) (*model.Run, error) {
	if taskID == "" {
		return nil, fmt.Errorf("taskID is required")
	}
	params := &model.PostTasksIDRunsAllParams{
		TaskID: taskID,
		Body: model.PostTasksIDRunsJSONRequestBody{
			// ScheduledFor not set for immediate execution
		},
	}
	return a.client.PostTasksIDRuns(ctx, params)
}

// CancelRun cancels a running task with given ID and given run ID.
func (a *TasksAPI) CancelRun(ctx context.Context, taskID, runID string) error {
	if taskID == "" {
		return fmt.Errorf("taskID is required")
	}
	if runID == "" {
		return fmt.Errorf("runID is required")
	}
	params := &model.DeleteTasksIDRunsIDAllParams{
		TaskID: taskID,
		RunID: runID,
	}
	return a.client.DeleteTasksIDRunsID(ctx, params)
}

// RetryRun retry a run with given ID of a task with given ID.
func (a *TasksAPI) RetryRun(ctx context.Context, taskID, runID string) (*model.Run, error) {
	if taskID == "" {
		return nil, fmt.Errorf("taskID is required")
	}
	if runID == "" {
		return nil, fmt.Errorf("runID is required")
	}
	params := &model.PostTasksIDRunsIDRetryAllParams{
		TaskID: taskID,
		RunID: runID,
	}
	return a.client.PostTasksIDRunsIDRetry(ctx, params)
}

// FindLogs retrieves all logs for a task with given ID.
func (a *TasksAPI) FindLogs(ctx context.Context, taskID string) ([]model.LogEvent, error) {
	if taskID == "" {
		return nil, fmt.Errorf("taskID is required")
	}
	params := &model.GetTasksIDLogsAllParams{
		TaskID: taskID,
	}
	response, err := a.client.GetTasksIDLogs(ctx, params)
	if err != nil {
		return nil, err
	}
	return *response.Events, nil
}

// FindLabels retrieves labels of a task with given ID.
func (a *TasksAPI) FindLabels(ctx context.Context, taskID string) ([]model.Label, error) {
	if taskID == "" {
		return nil, fmt.Errorf("taskID is required")
	}
	params := &model.GetTasksIDLabelsAllParams{
		TaskID: taskID,
	}
	response, err :=  a.client.GetTasksIDLabels(ctx, params)
	if err != nil {
		return nil, err
	}
	return *response.Labels, nil
}

// AddLabel adds a label with given ID to a task with given ID.
func (a *TasksAPI) AddLabel(ctx context.Context, taskID, labelID string) (*model.Label, error) {
	if taskID == "" {
		return nil, fmt.Errorf("taskID is required")
	}
	if labelID == "" {
		return nil, fmt.Errorf("labelID is required")
	}
	params := &model.PostTasksIDLabelsAllParams{
		TaskID: taskID,
		Body: model.PostTasksIDLabelsJSONRequestBody{
			LabelID: &labelID,
		},
	}
	response, err := a.client.PostTasksIDLabels(ctx, params)
	if err != nil {
		return nil, err
	}
	return response.Label, nil
}

// RemoveLabel removes a label with given ID  from a task with given ID.
func (a *TasksAPI) RemoveLabel(ctx context.Context, taskID, labelID string) error {
	if taskID == "" {
		return fmt.Errorf("taskID is required")
	}
	if labelID == "" {
		return fmt.Errorf("labelID is required")
	}
	params := &model.DeleteTasksIDLabelsIDAllParams{
		TaskID: taskID,
		LabelID: labelID,
	}
	return a.client.DeleteTasksIDLabelsID(ctx, params)
}

// Members returns all members of the task with the given ID.
func (a *TasksAPI) Members(ctx context.Context, taskID string) ([]model.ResourceMember, error) {
	if taskID == "" {
		return nil, fmt.Errorf("taskID is required")
	}
	params := &model.GetTasksIDMembersAllParams{
		TaskID: taskID,
	}
	response, err := a.client.GetTasksIDMembers(ctx, params)
	if err != nil {
		return nil, err
	}
	return *response.Users, nil
}

// AddMember adds the user with the given ID to the task with the given ID.
func (a *TasksAPI) AddMember(ctx context.Context, taskID, userID string) error {
	if taskID == "" {
		return fmt.Errorf("taskID is required")
	}
	if userID == "" {
		return fmt.Errorf("userID is required")
	}
	params := &model.PostTasksIDMembersAllParams{
		TaskID: taskID,
		Body: model.PostTasksIDMembersJSONRequestBody{
			Id: userID,
		},
	}
	_, err := a.client.PostTasksIDMembers(ctx, params)
	return err
}

// RemoveMember removes the user with the given ID from the task with the given ID.
func (a *TasksAPI) RemoveMember(ctx context.Context, taskID, userID string) error {
	if taskID == "" {
		return fmt.Errorf("taskID is required")
	}
	if userID == "" {
		return fmt.Errorf("userID is required")
	}
	params := &model.DeleteTasksIDMembersIDAllParams{
		TaskID: taskID,
		UserID: userID,
	}
	return a.client.DeleteTasksIDMembersID(ctx, params)
}

// Owners returns all the owners of the task with the given id.
func (a *TasksAPI) Owners(ctx context.Context, taskID string) ([]model.ResourceOwner, error) {
	if taskID == "" {
		return nil, fmt.Errorf("taskID is required")
	}
	params := &model.GetTasksIDOwnersAllParams{
		TaskID: taskID,
	}
	response, err := a.client.GetTasksIDOwners(ctx, params)
	if err != nil {
		return nil, err
	}
	return *response.Users, nil
}

// AddOwner adds an owner with the given userID to the task with the given id.
func (a *TasksAPI) AddOwner(ctx context.Context, taskID, userID string) error {
	if taskID == "" {
		return fmt.Errorf("taskID is required")
	}
	if userID == "" {
		return fmt.Errorf("userID is required")
	}
	params := &model.PostTasksIDOwnersAllParams{
		TaskID: taskID,
		Body: model.PostTasksIDOwnersJSONRequestBody{
			Id: userID,
		},
	}
	_, err := a.client.PostTasksIDOwners(ctx, params)
	return err
}

// RemoveOwner removes the user with the given userID from the task with the given id.
func (a *TasksAPI) RemoveOwner(ctx context.Context, taskID, userID string) error {
	if taskID == "" {
		return fmt.Errorf("taskID is required")
	}
	if userID == "" {
		return fmt.Errorf("userID is required")
	}
	params := &model.DeleteTasksIDOwnersIDAllParams{
		TaskID: taskID,
		UserID: userID,
	}
	return a.client.DeleteTasksIDOwnersID(ctx, params)
}

// getTasks returns list of tasks matching specified filter.
func (a *TasksAPI) getTasks(ctx context.Context, filter *Filter) ([]model.Task, error) {
	if filter != nil && filter.ID != "" {
		return a.getTasksByID(ctx, filter.ID)
	}
	params := &model.GetTasksParams{}
	if filter != nil {
		if filter.After != "" {
			params.After = &filter.After
		}
		if filter.Name != "" {
			params.Name = &filter.Name
		}
		if filter.UserName != "" {
			params.User = &filter.UserName
		}
		if filter.OrgID != "" {
			params.OrgID = &filter.OrgID
		}
		if filter.OrgName != "" {
			params.Org = &filter.OrgName
		}
		if filter.Status != "" {
			status := model.GetTasksParamsStatus(filter.Status)
			params.Status = &status
		}
		if filter.Limit > 0 {
			iLimit := int(filter.Limit)
			params.Limit = &iLimit
		}
	}
	response, err := a.client.GetTasks(ctx, params)
	if err != nil {
		return nil, err
	}
	return *response.Tasks, nil
}

// getTask returns tasks with matching ID, ie. just one.
func (a *TasksAPI) getTasksByID(ctx context.Context, taskID string) ([]model.Task, error) {
	params := &model.GetTasksIDAllParams{
		TaskID: taskID,
	}
	response, err := a.client.GetTasksID(ctx, params)
	if err != nil {
		return nil, err
	}
	return []model.Task{*response}, nil
}
