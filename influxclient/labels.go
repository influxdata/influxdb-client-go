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

// LabelsAPI provides methods for managing labels in a InfluxDB server.
type LabelsAPI struct {
	client *model.Client
}

// newLabelsAPI returns new LabelsAPI instance
func newLabelsAPI(client *model.Client) *LabelsAPI {
	return &LabelsAPI{client: client}
}

// Find returns labels matching the given filter.
// Supported filters:
//	- OrgID
func (a *LabelsAPI) Find(ctx context.Context, filter *Filter) ([]model.Label, error) {
	return a.getLabels(ctx, filter)
}

// FindOne returns one label matching the given filter.
// Supported filters:
//	- OrgID
func (a *LabelsAPI) FindOne(ctx context.Context, filter *Filter) (*model.Label, error) {
	labels, err := a.getLabels(ctx, filter)
	if err != nil {
		return nil, err
	}
	if len(labels) > 0 {
		return &(labels[0]), nil
	}
	return nil, fmt.Errorf("label not found")
}

// Create creates a new label with the given information.
// The label.Name field must be non-empty.
// The returned Label holds the ID of the new label.
func (a *LabelsAPI) Create(ctx context.Context, label *model.Label) (*model.Label, error) {
	if label == nil {
		return nil, fmt.Errorf("label cannot be nil")
	}
	if label.Name == nil {
		return nil, fmt.Errorf("name is required")
	}
	if label.OrgID == nil {
		return nil, fmt.Errorf("orgId is required")
	}
	params := &model.PostLabelsAllParams{
		Body: model.PostLabelsJSONRequestBody{
			Name:        *(label.Name),
			OrgID:       *(label.OrgID),
		},
	}
	if label.Properties != nil {
		params.Body.Properties = &model.LabelCreateRequest_Properties{
			AdditionalProperties: label.Properties.AdditionalProperties,
		}
	}
	response, err := a.client.PostLabels(ctx, params)
	if err != nil {
		return nil, err
	}
	return response.Label, nil
}

// Update updates the label's name and properties.
// If the name is empty, it won't be changed. If a property isn't mentioned, it won't be changed.
// A property can be removed by using an empty value for that property.
//
// Update returns the fully updated label.
func (a *LabelsAPI) Update(ctx context.Context, label *model.Label) (*model.Label, error) {
	if label == nil {
		return nil, fmt.Errorf("label cannot be nil")
	}
	if label.Id == nil {
		return nil, fmt.Errorf("label ID is required")
	}
	params := &model.PatchLabelsIDAllParams{
		LabelID: *label.Id,
		Body: model.PatchLabelsIDJSONRequestBody{
			Name: label.Name,
		},
	}
	if label.Properties != nil {
		params.Body.Properties = &model.LabelUpdate_Properties{
			AdditionalProperties: label.Properties.AdditionalProperties,
		}
	}
	response, err := a.client.PatchLabelsID(ctx, params)
	if err != nil {
		return nil, err
	}
	return response.Label, nil
}

// Delete deletes the label with the given ID.
func (a *LabelsAPI) Delete(ctx context.Context, labelID string) error {
	params := &model.DeleteLabelsIDAllParams{
		LabelID: labelID,
	}
	return a.client.DeleteLabelsID(ctx, params)
}

// getLabels create request for GET on /labels according to the filter and validates returned structure
func (a *LabelsAPI) getLabels(ctx context.Context, filter *Filter) ([]model.Label, error) {
	params := &model.GetLabelsParams{}
	if filter != nil {
		params.OrgID = &filter.OrgID
	}
	response, err := a.client.GetLabels(ctx, params)
	if err != nil {
		return nil, err
	}
	return *response.Labels, nil
}
