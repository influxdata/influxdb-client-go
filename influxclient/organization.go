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
		offset := model.Offset(filter.Offset)
		params.Org = &filter.OrgName
		params.OrgID = &filter.OrgID
		params.UserID = &filter.UserID
		if filter.Limit > 0 {
			limit := model.Limit(filter.Limit)
			params.Limit = &limit
		}
		params.Offset = &offset
	}
	orgs, err := o.client.GetOrgs(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("cannot retrieve organizations: %w", err)
	}
	if orgs == nil {
		return nil, fmt.Errorf("organization not found")
	}
	return *orgs.Orgs, nil
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
	params := &model.PostOrgsAllParams{}
	params.Body = model.PostOrgsJSONRequestBody{
		Name:        org.Name,
		Description: org.Description,
	}
	org, err := o.client.PostOrgs(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("cannot create organization: %w", err)
	}
	return org, nil
}

// Update updates information about the organization. The org.ID field must hold the ID
// of the organization to be changed.
func (o *OrganizationAPI) Update(ctx context.Context, org *model.Organization) (*model.Organization, error) {
	return nil, nil
}

// Delete deletes the organization with the given ID.
func (o *OrganizationAPI) Delete(ctx context.Context, orgID string) error {
	return nil
}

// Members returns all members of the organization with the given ID.
func (o *OrganizationAPI) Members(ctx context.Context, orgID string) ([]model.ResourceMember, error) {
	return nil, nil
}

// AddMember adds the user with the given ID to the organization with the given ID.
func (o *OrganizationAPI) AddMember(ctx context.Context, orgID, userID string) error {
	return nil
}

// RemoveMember AddMember removes the user with the given ID from the organization with the given ID.
func (o *OrganizationAPI) RemoveMember(ctx context.Context, orgID, userID string) error {
	return nil
}

// Owners returns all the owners of the organization with the given id.
func (o *OrganizationAPI) Owners(ctx context.Context, orgID string) ([]model.ResourceOwner, error) {
	return nil, nil
}

// AddOwner adds an owner with the given userID to the organization with the given id.
func (o *OrganizationAPI) AddOwner(ctx context.Context, orgID, userID string) error {
	return nil
}

// RemoveOwner Remove removes the user with the given userID from the organization with the given id.
func (o *OrganizationAPI) RemoveOwner(ctx context.Context, orgID, userID string) error {
	return nil
}
