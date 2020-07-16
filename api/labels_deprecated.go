// Copyright 2020 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package api

import (
	"context"
	"github.com/influxdata/influxdb-client-go/domain"
)

// LabelsApi provides methods for managing labels in a InfluxDB server.
// Deprecated: Use LabelsAPI instead.
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
	labelsAPI LabelsAPI
}

// Deprecated: use NewLabelsAPI instead.
func NewLabelsApi(apiClient *domain.ClientWithResponses) LabelsApi {
	return &labelsApiImpl{
		labelsAPI: NewLabelsAPI(apiClient),
	}
}

func (l *labelsApiImpl) GetLabels(ctx context.Context) (*[]domain.Label, error) {
	return l.labelsAPI.GetLabels(ctx)
}

func (l *labelsApiImpl) FindLabelsByOrg(ctx context.Context, org *domain.Organization) (*[]domain.Label, error) {
	return l.labelsAPI.FindLabelsByOrg(ctx, org)
}

func (l *labelsApiImpl) FindLabelsByOrgId(ctx context.Context, orgID string) (*[]domain.Label, error) {
	return l.labelsAPI.FindLabelsByOrgID(ctx, orgID)
}

func (l *labelsApiImpl) FindLabelById(ctx context.Context, labelID string) (*domain.Label, error) {
	return l.labelsAPI.FindLabelByID(ctx, labelID)
}

func (l *labelsApiImpl) FindLabelByName(ctx context.Context, orgId, labelName string) (*domain.Label, error) {
	return l.labelsAPI.FindLabelByName(ctx, orgId, labelName)
}

func (l *labelsApiImpl) CreateLabelWithName(ctx context.Context, org *domain.Organization, labelName string, properties map[string]string) (*domain.Label, error) {
	return l.labelsAPI.CreateLabelWithName(ctx, org, labelName, properties)
}

func (l *labelsApiImpl) CreateLabelWithNameWithId(ctx context.Context, orgId, labelName string, properties map[string]string) (*domain.Label, error) {
	return l.labelsAPI.CreateLabelWithNameWithID(ctx, orgId, labelName, properties)
}

func (l *labelsApiImpl) CreateLabel(ctx context.Context, label *domain.LabelCreateRequest) (*domain.Label, error) {
	return l.labelsAPI.CreateLabel(ctx, label)
}

func (l *labelsApiImpl) UpdateLabel(ctx context.Context, label *domain.Label) (*domain.Label, error) {
	return l.labelsAPI.UpdateLabel(ctx, label)
}

func (l *labelsApiImpl) DeleteLabel(ctx context.Context, label *domain.Label) error {
	return l.labelsAPI.DeleteLabel(ctx, label)
}

func (l *labelsApiImpl) DeleteLabelWithId(ctx context.Context, labelID string) error {
	return l.labelsAPI.DeleteLabelWithID(ctx, labelID)
}
