// Copyright 2020 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package api

import (
	"context"
	"fmt"
	"github.com/influxdata/influxdb-client-go/domain"
)

// LabelsApi provides methods for managing labels in a InfluxDB server.
type LabelsApi interface {
	// GetLabels returns all labels.
	GetLabels(ctx context.Context) (*[]domain.Label, error)
	// FindLabelsByOrg returns labels belonging to organization org.
	FindLabelsByOrg(ctx context.Context, org *domain.Organization) (*[]domain.Label, error)
	// FindLabelsByOrgId returns labels belonging to organization with id orgId.
	FindLabelsByOrgId(ctx context.Context, orgID string) (*[]domain.Label, error)
	// FindLabelById returns a label with labelID.
	FindLabelById(ctx context.Context, labelID string) (*domain.Label, error)
	// FindLabelByName returns a label with name labelName under an organization orgId.
	FindLabelByName(ctx context.Context, orgId, labelName string) (*domain.Label, error)
	// CreateLabel creates a new label.
	CreateLabel(ctx context.Context, label *domain.LabelCreateRequest) (*domain.Label, error)
	// CreateLabelWithName creates a new label with label labelName and properties, under the organization org.
	// Properties example: {"color": "ffb3b3", "description": "this is a description"}.
	CreateLabelWithName(ctx context.Context, org *domain.Organization, labelName string, properties map[string]string) (*domain.Label, error)
	// CreateLabelWithName creates a new label with label labelName and properties, under the organization with id orgId.
	// Properties example: {"color": "ffb3b3", "description": "this is a description"}.
	CreateLabelWithNameWithId(ctx context.Context, orgId, labelName string, properties map[string]string) (*domain.Label, error)
	// UpdateLabel updates the label.
	// Properties can be removed by sending an update with an empty value.
	UpdateLabel(ctx context.Context, label *domain.Label) (*domain.Label, error)
	// DeleteLabelWithId deletes a label with labelId.
	DeleteLabelWithId(ctx context.Context, labelID string) error
	// DeleteLabel deletes a label.
	DeleteLabel(ctx context.Context, label *domain.Label) error
}

type labelsApiImpl struct {
	apiClient *domain.ClientWithResponses
}

func NewLabelsApi(apiClient *domain.ClientWithResponses) LabelsApi {
	return &labelsApiImpl{
		apiClient: apiClient,
	}
}

func (u *labelsApiImpl) GetLabels(ctx context.Context) (*[]domain.Label, error) {
	params := &domain.GetLabelsParams{}
	return u.getLabels(ctx, params)
}

func (u *labelsApiImpl) getLabels(ctx context.Context, params *domain.GetLabelsParams) (*[]domain.Label, error) {
	response, err := u.apiClient.GetLabelsWithResponse(ctx, params)
	if err != nil {
		return nil, err
	}
	if response.JSONDefault != nil {
		return nil, domain.DomainErrorToError(response.JSONDefault, response.StatusCode())
	}
	return (*[]domain.Label)(response.JSON200.Labels), nil
}

func (u *labelsApiImpl) FindLabelsByOrg(ctx context.Context, org *domain.Organization) (*[]domain.Label, error) {
	return u.FindLabelsByOrgId(ctx, *org.Id)
}

func (u *labelsApiImpl) FindLabelsByOrgId(ctx context.Context, orgID string) (*[]domain.Label, error) {
	params := &domain.GetLabelsParams{OrgID: &orgID}
	return u.getLabels(ctx, params)
}

func (u *labelsApiImpl) FindLabelById(ctx context.Context, labelID string) (*domain.Label, error) {
	params := &domain.GetLabelsIDParams{}
	response, err := u.apiClient.GetLabelsIDWithResponse(ctx, labelID, params)
	if err != nil {
		return nil, err
	}
	if response.JSONDefault != nil {
		return nil, domain.DomainErrorToError(response.JSONDefault, response.StatusCode())
	}
	return response.JSON200.Label, nil
}

func (u *labelsApiImpl) FindLabelByName(ctx context.Context, orgId, labelName string) (*domain.Label, error) {
	labels, err := u.FindLabelsByOrgId(ctx, orgId)
	if err != nil {
		return nil, err
	}
	var label *domain.Label
	for _, u := range *labels {
		if *u.Name == labelName {
			label = &u
			break
		}
	}
	if label == nil {
		return nil, fmt.Errorf("label '%s' not found", labelName)
	}
	return label, nil
}

func (u *labelsApiImpl) CreateLabelWithName(ctx context.Context, org *domain.Organization, labelName string, properties map[string]string) (*domain.Label, error) {
	return u.CreateLabelWithNameWithId(ctx, *org.Id, labelName, properties)
}

func (u *labelsApiImpl) CreateLabelWithNameWithId(ctx context.Context, orgId, labelName string, properties map[string]string) (*domain.Label, error) {
	props := &domain.LabelCreateRequest_Properties{AdditionalProperties: properties}
	label := &domain.LabelCreateRequest{Name: &labelName, OrgID: orgId, Properties: props}
	return u.CreateLabel(ctx, label)
}

func (u *labelsApiImpl) CreateLabel(ctx context.Context, label *domain.LabelCreateRequest) (*domain.Label, error) {
	response, err := u.apiClient.PostLabelsWithResponse(ctx, domain.PostLabelsJSONRequestBody(*label))
	if err != nil {
		return nil, err
	}
	if response.JSONDefault != nil {
		return nil, domain.DomainErrorToError(response.JSONDefault, response.StatusCode())
	}
	return response.JSON201.Label, nil
}

func (u *labelsApiImpl) UpdateLabel(ctx context.Context, label *domain.Label) (*domain.Label, error) {
	var props *domain.LabelUpdate_Properties
	params := &domain.PatchLabelsIDParams{}
	if label.Properties != nil {
		props = &domain.LabelUpdate_Properties{AdditionalProperties: label.Properties.AdditionalProperties}
	}
	body := &domain.LabelUpdate{
		Name:       label.Name,
		Properties: props,
	}
	response, err := u.apiClient.PatchLabelsIDWithResponse(ctx, *label.Id, params, domain.PatchLabelsIDJSONRequestBody(*body))
	if err != nil {
		return nil, err
	}
	if response.JSON404 != nil {
		return nil, domain.DomainErrorToError(response.JSON404, response.StatusCode())
	}
	if response.JSONDefault != nil {
		return nil, domain.DomainErrorToError(response.JSONDefault, response.StatusCode())
	}
	return response.JSON200.Label, nil
}

func (u *labelsApiImpl) DeleteLabel(ctx context.Context, label *domain.Label) error {
	return u.DeleteLabelWithId(ctx, *label.Id)
}

func (u *labelsApiImpl) DeleteLabelWithId(ctx context.Context, labelID string) error {
	params := &domain.DeleteLabelsIDParams{}
	response, err := u.apiClient.DeleteLabelsIDWithResponse(ctx, labelID, params)
	if err != nil {
		return err
	}
	if response.JSON404 != nil {
		return domain.DomainErrorToError(response.JSON404, response.StatusCode())
	}
	if response.JSONDefault != nil {
		return domain.DomainErrorToError(response.JSONDefault, response.StatusCode())
	}
	return nil
}
