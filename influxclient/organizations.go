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

// OrganizationAPI holds methods related to organization, as found under
// the /orgs endpoint.
type OrganizationAPI struct {
	client *model.Client
}

// newOrganizationAPI returns new OrganizationAPI instance
func newOrganizationAPI(client *model.Client) *OrganizationAPI {
	return &OrganizationAPI{client: client}
}

// Find returns all organizations matching the given filter.
// Supported filters:
//	- OrgName
//	- OrgID
//	- UserID
func (o *OrganizationAPI) Find(ctx context.Context, filter *Filter) ([]model.Organization, error) {
	return o.getOrganizations(ctx, filter)
}

// getOrganizations create request for GET on /orgs  according to the filter and validates returned structure
func (o *OrganizationAPI) getOrganizations(ctx context.Context, filter *Filter) ([]model.Organization, error) {
	params := &model.GetOrgsParams{}
	if filter != nil {
		if filter.OrgName != "" {
			params.Org = &filter.OrgName
		}
		if filter.OrgID != "" {
			params.OrgID = &filter.OrgID
		}
		if filter.UserID != "" {
			params.UserID = &filter.UserID
		}
		if filter.Limit > 0 {
			limit := model.Limit(filter.Limit)
			params.Limit = &limit
		}
		if filter.Offset > 0 {
			offset := model.Offset(filter.Offset)
			params.Offset = &offset
		}
	}
	response, err := o.client.GetOrgs(ctx, params)
	if err != nil {
		return nil, err
	}
	return *response.Orgs, nil
}

// FindOne returns one organization matching the given filter.
// Supported filters:
//	- OrgName
//	- OrgID
//	- UserID
func (o *OrganizationAPI) FindOne(ctx context.Context, filter *Filter) (*model.Organization, error) {
	organizations, err := o.getOrganizations(ctx, filter)
	if err != nil {
		return nil, err
	}
	if len(organizations) > 0 {
		return &(organizations)[0], nil
	}
	return nil, fmt.Errorf("organization not found")
}

// Create creates a new organization. The returned Organization holds the new ID.
func (o *OrganizationAPI) Create(ctx context.Context, org *model.Organization) (*model.Organization, error) {
	if org == nil {
		return nil, fmt.Errorf("org cannot be nil")
	}
	if org.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	params := &model.PostOrgsAllParams{
		Body: model.PostOrgsJSONRequestBody{
			Name:        org.Name,
			Description: org.Description,
		},
	}
	return o.client.PostOrgs(ctx, params)
}

// Update updates information about the organization. The org.ID field must hold the ID
// of the organization to be changed.
func (o *OrganizationAPI) Update(ctx context.Context, org *model.Organization) (*model.Organization, error) {
	if org == nil {
		return nil, fmt.Errorf("org cannot be nil")
	}
	if org.Id == nil {
		return nil, fmt.Errorf("org ID is required")
	}
	if org.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	params := &model.PatchOrgsIDAllParams{
		OrgID: *org.Id,
		Body: model.PatchOrgsIDJSONRequestBody{
			Name: &org.Name,
			Description: org.Description,
		},
	}
	return o.client.PatchOrgsID(ctx, params)
}

// Delete deletes the organization with the given ID.
func (o *OrganizationAPI) Delete(ctx context.Context, orgID string) error {
	if orgID == "" {
		return fmt.Errorf("orgID is required")
	}
	params := &model.DeleteOrgsIDAllParams{
		OrgID: orgID,
	}
	return o.client.DeleteOrgsID(ctx, params)
}

// Members returns all members of the organization with the given ID.
func (o *OrganizationAPI) Members(ctx context.Context, orgID string) ([]model.ResourceMember, error) {
	if orgID == "" {
		return nil, fmt.Errorf("orgID is required")
	}
	params := &model.GetOrgsIDMembersAllParams{
		OrgID: orgID,
	}
	response,err := o.client.GetOrgsIDMembers(ctx, params)
	if err != nil {
		return nil, err
	}
	return *response.Users, nil
}

// AddMember adds the user with the given ID to the organization with the given ID.
func (o *OrganizationAPI) AddMember(ctx context.Context, orgID, userID string) error {
	if orgID == "" {
		return fmt.Errorf("orgID is required")
	}
	if userID == "" {
		return fmt.Errorf("userID is required")
	}
	params := &model.PostOrgsIDMembersAllParams{
		OrgID: orgID,
		Body: model.PostOrgsIDMembersJSONRequestBody{
			Id: userID,
		},
	}
	_, err := o.client.PostOrgsIDMembers(ctx, params)
	return err
}

// RemoveMember removes the user with the given ID from the organization with the given ID.
func (o *OrganizationAPI) RemoveMember(ctx context.Context, orgID, userID string) error {
	if orgID == "" {
		return fmt.Errorf("orgID is required")
	}
	if userID == "" {
		return fmt.Errorf("userID is required")
	}
	params := &model.DeleteOrgsIDMembersIDAllParams{
		OrgID: orgID,
		UserID: userID,
	}
	return o.client.DeleteOrgsIDMembersID(ctx, params)
}

// Owners returns all the owners of the organization with the given id.
func (o *OrganizationAPI) Owners(ctx context.Context, orgID string) ([]model.ResourceOwner, error) {
	if orgID == "" {
		return nil, fmt.Errorf("orgID is required")
	}
	params := &model.GetOrgsIDOwnersAllParams{
		OrgID: orgID,
	}
	response, err := o.client.GetOrgsIDOwners(ctx, params)
	if err != nil {
		return nil, err
	}
	return *response.Users, nil
}

// AddOwner adds an owner with the given userID to the organization with the given id.
func (o *OrganizationAPI) AddOwner(ctx context.Context, orgID, userID string) error {
	if orgID == "" {
		return fmt.Errorf("orgID is required")
	}
	if userID == "" {
		return fmt.Errorf("userID is required")
	}
	params := &model.PostOrgsIDOwnersAllParams{
		OrgID: orgID,
		Body: model.PostOrgsIDOwnersJSONRequestBody{
			Id: userID,
		},
	}
	_, err := o.client.PostOrgsIDOwners(ctx, params)
	return err
}

// RemoveOwner Remove removes the user with the given userID from the organization with the given id.
func (o *OrganizationAPI) RemoveOwner(ctx context.Context, orgID, userID string) error {
	if orgID == "" {
		return fmt.Errorf("orgID is required")
	}
	if userID == "" {
		return fmt.Errorf("userID is required")
	}
	params := &model.DeleteOrgsIDOwnersIDAllParams{
		OrgID: orgID,
		UserID: userID,
	}
	return o.client.DeleteOrgsIDOwnersID(ctx, params)
}
