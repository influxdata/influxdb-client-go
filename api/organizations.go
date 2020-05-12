// Copyright 2020 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package api

import (
	"context"
	"github.com/influxdata/influxdb-client-go/domain"
)

// OrganizationsApi provides methods for managing Organizations in a InfluxDB server
type OrganizationsApi interface {
	// GetOrganizations returns all organizations
	GetOrganizations(ctx context.Context) (*[]domain.Organization, error)
	// FindOrganizationByName returns an organization found using orgName
	FindOrganizationByName(ctx context.Context, orgName string) (*domain.Organization, error)
	// FindOrganizationById returns an organization found using orgId
	FindOrganizationById(ctx context.Context, orgId string) (*domain.Organization, error)
	// FindOrganizationsByUserId returns organizations an user with userID belongs to
	FindOrganizationsByUserId(ctx context.Context, orgId string) (*[]domain.Organization, error)
	// CreateOrganization creates new organization
	CreateOrganization(ctx context.Context, org *domain.Organization) (*domain.Organization, error)
	// CreateOrganizationWithName creates new organization with orgName and with status active
	CreateOrganizationWithName(ctx context.Context, orgName string) (*domain.Organization, error)
	// UpdateOrganization updates organization
	UpdateOrganization(ctx context.Context, org *domain.Organization) (*domain.Organization, error)
	// DeleteOrganization deletes an organization
	DeleteOrganization(ctx context.Context, org *domain.Organization) error
	// DeleteOrganizationWithId deletes an organization with orgId
	DeleteOrganizationWithId(ctx context.Context, orgId string) error
	// GetMembers returns members of an organization
	GetMembers(ctx context.Context, org *domain.Organization) (*[]domain.ResourceMember, error)
	// GetMembersWithId returns members of an organization with orgId
	GetMembersWithId(ctx context.Context, orgId string) (*[]domain.ResourceMember, error)
	// AddMember add a user to an organization
	AddMember(ctx context.Context, org *domain.Organization, user *domain.User) (*domain.ResourceMember, error)
	// AddMember add a member with id memberId to an organization with orgId
	AddMemberWithId(ctx context.Context, orgId, memberId string) (*domain.ResourceMember, error)
	// RemoveMember removes a user from an organization
	RemoveMember(ctx context.Context, org *domain.Organization, user *domain.User) error
	// RemoveMember removes a member with id memberId from an organization with orgId
	RemoveMemberWithId(ctx context.Context, orgId, memberId string) error
	// GetOwners returns owners of an organization
	GetOwners(ctx context.Context, org *domain.Organization) (*[]domain.ResourceOwner, error)
	// GetOwnersWithId returns owners of an organization with orgId
	GetOwnersWithId(ctx context.Context, orgId string) (*[]domain.ResourceOwner, error)
	// AddOwner add a user to an organization
	AddOwner(ctx context.Context, org *domain.Organization, user *domain.User) (*domain.ResourceOwner, error)
	// AddOwner add an owner with id memberId to an organization with orgId
	AddOwnerWithId(ctx context.Context, orgId, memberId string) (*domain.ResourceOwner, error)
	// RemoveOwner  a user from an organization
	RemoveOwner(ctx context.Context, org *domain.Organization, user *domain.User) error
	// RemoveOwner removes a member with id memberId from an organization with orgId
	RemoveOwnerWithId(ctx context.Context, orgId, memberId string) error
}

type organizationsApiImpl struct {
	apiClient *domain.ClientWithResponses
}

func NewOrganizationsApi(apiClient *domain.ClientWithResponses) OrganizationsApi {
	return &organizationsApiImpl{
		apiClient: apiClient,
	}
}

func (o *organizationsApiImpl) GetOrganizations(ctx context.Context) (*[]domain.Organization, error) {
	params := &domain.GetOrgsParams{}
	response, err := o.apiClient.GetOrgsWithResponse(ctx, params)
	if err != nil {
		return nil, err
	}
	if response.JSONDefault != nil {
		return nil, domain.DomainErrorToError(response.JSONDefault, response.StatusCode())
	}
	return response.JSON200.Orgs, nil
}

func (o *organizationsApiImpl) FindOrganizationByName(ctx context.Context, orgName string) (*domain.Organization, error) {
	params := &domain.GetOrgsParams{Org: &orgName}
	response, err := o.apiClient.GetOrgsWithResponse(ctx, params)
	if err != nil {
		return nil, err
	}
	if response.JSONDefault != nil {
		return nil, domain.DomainErrorToError(response.JSONDefault, response.StatusCode())
	}
	return &(*response.JSON200.Orgs)[0], nil
}

func (o *organizationsApiImpl) FindOrganizationById(ctx context.Context, orgId string) (*domain.Organization, error) {
	params := &domain.GetOrgsParams{OrgID: &orgId}
	response, err := o.apiClient.GetOrgsWithResponse(ctx, params)
	if err != nil {
		return nil, err
	}
	if response.JSONDefault != nil {
		return nil, domain.DomainErrorToError(response.JSONDefault, response.StatusCode())
	}
	return &(*response.JSON200.Orgs)[0], nil
}

func (o *organizationsApiImpl) FindOrganizationsByUserId(ctx context.Context, userID string) (*[]domain.Organization, error) {
	params := &domain.GetOrgsParams{UserID: &userID}
	response, err := o.apiClient.GetOrgsWithResponse(ctx, params)
	if err != nil {
		return nil, err
	}
	if response.JSONDefault != nil {
		return nil, domain.DomainErrorToError(response.JSONDefault, response.StatusCode())
	}
	return response.JSON200.Orgs, nil
}

func (o *organizationsApiImpl) CreateOrganization(ctx context.Context, org *domain.Organization) (*domain.Organization, error) {
	params := &domain.PostOrgsParams{}
	response, err := o.apiClient.PostOrgsWithResponse(ctx, params, domain.PostOrgsJSONRequestBody(*org))
	if err != nil {
		return nil, err
	}
	if response.JSONDefault != nil {
		return nil, domain.DomainErrorToError(response.JSONDefault, response.StatusCode())
	}
	return response.JSON201, nil
}

func (o *organizationsApiImpl) CreateOrganizationWithName(ctx context.Context, orgName string) (*domain.Organization, error) {
	params := &domain.PostOrgsParams{}
	status := domain.OrganizationStatusActive
	org := &domain.Organization{Name: orgName, Status: &status}
	response, err := o.apiClient.PostOrgsWithResponse(ctx, params, domain.PostOrgsJSONRequestBody(*org))
	if err != nil {
		return nil, err
	}
	if response.JSONDefault != nil {
		return nil, domain.DomainErrorToError(response.JSONDefault, response.StatusCode())
	}
	return response.JSON201, nil
}

func (o *organizationsApiImpl) DeleteOrganization(ctx context.Context, org *domain.Organization) error {
	return o.DeleteOrganizationWithId(ctx, *org.Id)
}

func (o *organizationsApiImpl) DeleteOrganizationWithId(ctx context.Context, orgId string) error {
	params := &domain.DeleteOrgsIDParams{}
	response, err := o.apiClient.DeleteOrgsIDWithResponse(ctx, orgId, params)
	if err != nil {
		return err
	}
	if response.JSONDefault != nil {
		return domain.DomainErrorToError(response.JSONDefault, response.StatusCode())
	}
	if response.JSON404 != nil {
		return domain.DomainErrorToError(response.JSON404, response.StatusCode())
	}
	return nil
}

func (o *organizationsApiImpl) UpdateOrganization(ctx context.Context, org *domain.Organization) (*domain.Organization, error) {
	params := &domain.PatchOrgsIDParams{}
	response, err := o.apiClient.PatchOrgsIDWithResponse(ctx, *org.Id, params, domain.PatchOrgsIDJSONRequestBody(*org))
	if err != nil {
		return nil, err
	}
	if response.JSONDefault != nil {
		return nil, domain.DomainErrorToError(response.JSONDefault, response.StatusCode())
	}
	return response.JSON200, nil
}

func (o *organizationsApiImpl) GetMembers(ctx context.Context, org *domain.Organization) (*[]domain.ResourceMember, error) {
	return o.GetMembersWithId(ctx, *org.Id)
}

func (o *organizationsApiImpl) GetMembersWithId(ctx context.Context, orgId string) (*[]domain.ResourceMember, error) {
	params := &domain.GetOrgsIDMembersParams{}
	response, err := o.apiClient.GetOrgsIDMembersWithResponse(ctx, orgId, params)
	if err != nil {
		return nil, err
	}
	if response.JSONDefault != nil {
		return nil, domain.DomainErrorToError(response.JSONDefault, response.StatusCode())
	}
	if response.JSON404 != nil {
		return nil, domain.DomainErrorToError(response.JSON404, response.StatusCode())
	}
	return response.JSON200.Users, nil
}

func (o *organizationsApiImpl) AddMember(ctx context.Context, org *domain.Organization, user *domain.User) (*domain.ResourceMember, error) {
	return o.AddMemberWithId(ctx, *org.Id, *user.Id)
}

func (o *organizationsApiImpl) AddMemberWithId(ctx context.Context, orgId, memberId string) (*domain.ResourceMember, error) {
	params := &domain.PostOrgsIDMembersParams{}
	body := &domain.PostOrgsIDMembersJSONRequestBody{Id: memberId}
	response, err := o.apiClient.PostOrgsIDMembersWithResponse(ctx, orgId, params, *body)
	if err != nil {
		return nil, err
	}
	if response.JSONDefault != nil {
		return nil, domain.DomainErrorToError(response.JSONDefault, response.StatusCode())
	}
	return response.JSON201, nil
}

func (o *organizationsApiImpl) RemoveMember(ctx context.Context, org *domain.Organization, user *domain.User) error {
	return o.RemoveMemberWithId(ctx, *org.Id, *user.Id)
}

func (o *organizationsApiImpl) RemoveMemberWithId(ctx context.Context, orgId, memberId string) error {
	params := &domain.DeleteOrgsIDMembersIDParams{}
	response, err := o.apiClient.DeleteOrgsIDMembersIDWithResponse(ctx, orgId, memberId, params)
	if err != nil {
		return err
	}
	if response.JSONDefault != nil {
		return domain.DomainErrorToError(response.JSONDefault, response.StatusCode())
	}
	return nil
}

func (o *organizationsApiImpl) GetOwners(ctx context.Context, org *domain.Organization) (*[]domain.ResourceOwner, error) {
	return o.GetOwnersWithId(ctx, *org.Id)
}

func (o *organizationsApiImpl) GetOwnersWithId(ctx context.Context, orgId string) (*[]domain.ResourceOwner, error) {
	params := &domain.GetOrgsIDOwnersParams{}
	response, err := o.apiClient.GetOrgsIDOwnersWithResponse(ctx, orgId, params)
	if err != nil {
		return nil, err
	}
	if response.JSONDefault != nil {
		return nil, domain.DomainErrorToError(response.JSONDefault, response.StatusCode())
	}
	if response.JSON404 != nil {
		return nil, domain.DomainErrorToError(response.JSON404, response.StatusCode())
	}
	return response.JSON200.Users, nil
}

func (o *organizationsApiImpl) AddOwner(ctx context.Context, org *domain.Organization, user *domain.User) (*domain.ResourceOwner, error) {
	return o.AddOwnerWithId(ctx, *org.Id, *user.Id)
}

func (o *organizationsApiImpl) AddOwnerWithId(ctx context.Context, orgId, ownerId string) (*domain.ResourceOwner, error) {
	params := &domain.PostOrgsIDOwnersParams{}
	body := &domain.PostOrgsIDOwnersJSONRequestBody{Id: ownerId}
	response, err := o.apiClient.PostOrgsIDOwnersWithResponse(ctx, orgId, params, *body)
	if err != nil {
		return nil, err
	}
	if response.JSONDefault != nil {
		return nil, domain.DomainErrorToError(response.JSONDefault, response.StatusCode())
	}
	return response.JSON201, nil
}

func (o *organizationsApiImpl) RemoveOwner(ctx context.Context, org *domain.Organization, user *domain.User) error {
	return o.RemoveOwnerWithId(ctx, *org.Id, *user.Id)
}

func (o *organizationsApiImpl) RemoveOwnerWithId(ctx context.Context, orgId, memberId string) error {
	params := &domain.DeleteOrgsIDOwnersIDParams{}
	response, err := o.apiClient.DeleteOrgsIDOwnersIDWithResponse(ctx, orgId, memberId, params)
	if err != nil {
		return err
	}
	if response.JSONDefault != nil {
		return domain.DomainErrorToError(response.JSONDefault, response.StatusCode())
	}
	return nil
}
