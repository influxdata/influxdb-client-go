// Copyright 2020 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package api

import (
	"context"
	"github.com/influxdata/influxdb-client-go/domain"
)

//lint:file-ignore ST1003 This is deprecated API to be removed in next release.

// OrganizationsApi provides methods for managing Organizations in a InfluxDB server.
// Deprecated: Use OrganizationsAPI instead.
type OrganizationsApi interface {
	// GetOrganizations returns all organizations.
	GetOrganizations(ctx context.Context) (*[]domain.Organization, error)
	// FindOrganizationByName returns an organization found using orgName.
	FindOrganizationByName(ctx context.Context, orgName string) (*domain.Organization, error)
	// FindOrganizationById returns an organization found using orgId.
	FindOrganizationById(ctx context.Context, orgId string) (*domain.Organization, error)
	// FindOrganizationsByUserId returns organizations an user with userID belongs to.
	FindOrganizationsByUserId(ctx context.Context, userID string) (*[]domain.Organization, error)
	// CreateOrganization creates new organization.
	CreateOrganization(ctx context.Context, org *domain.Organization) (*domain.Organization, error)
	// CreateOrganizationWithName creates new organization with orgName and with status active.
	CreateOrganizationWithName(ctx context.Context, orgName string) (*domain.Organization, error)
	// UpdateOrganization updates organization.
	UpdateOrganization(ctx context.Context, org *domain.Organization) (*domain.Organization, error)
	// DeleteOrganization deletes an organization.
	DeleteOrganization(ctx context.Context, org *domain.Organization) error
	// DeleteOrganizationWithId deletes an organization with orgId.
	DeleteOrganizationWithId(ctx context.Context, orgId string) error
	// GetMembers returns members of an organization.
	GetMembers(ctx context.Context, org *domain.Organization) (*[]domain.ResourceMember, error)
	// GetMembersWithId returns members of an organization with orgId.
	GetMembersWithId(ctx context.Context, orgId string) (*[]domain.ResourceMember, error)
	// AddMember adds a member to an organization.
	AddMember(ctx context.Context, org *domain.Organization, user *domain.User) (*domain.ResourceMember, error)
	// AddMemberWithId adds a member with id memberId to an organization with orgId.
	AddMemberWithId(ctx context.Context, orgId, memberId string) (*domain.ResourceMember, error)
	// RemoveMember removes a member from an organization.
	RemoveMember(ctx context.Context, org *domain.Organization, user *domain.User) error
	// RemoveMemberWithId removes a member with id memberId from an organization with orgId.
	RemoveMemberWithId(ctx context.Context, orgId, memberId string) error
	// GetOwners returns owners of an organization.
	GetOwners(ctx context.Context, org *domain.Organization) (*[]domain.ResourceOwner, error)
	// GetOwnersWithId returns owners of an organization with orgId.
	GetOwnersWithId(ctx context.Context, orgId string) (*[]domain.ResourceOwner, error)
	// AddOwner adds an owner to an organization.
	AddOwner(ctx context.Context, org *domain.Organization, user *domain.User) (*domain.ResourceOwner, error)
	// AddOwnerWithId adds an owner with id memberId to an organization with orgId.
	AddOwnerWithId(ctx context.Context, orgId, memberId string) (*domain.ResourceOwner, error)
	// RemoveOwner removes an owner from an organization.
	RemoveOwner(ctx context.Context, org *domain.Organization, user *domain.User) error
	// RemoveOwnerWithId removes an owner with id memberId from an organization with orgId.
	RemoveOwnerWithId(ctx context.Context, orgId, memberId string) error
}

type organizationsApiImpl struct {
	organizationsAPI OrganizationsAPI
}

// NewOrganizationsApi creates instance of OrganizationsApi
// Deprecated: use NewOrganizationsAPI instead.
func NewOrganizationsApi(apiClient *domain.ClientWithResponses) OrganizationsApi {
	return &organizationsApiImpl{
		organizationsAPI: NewOrganizationsAPI(apiClient),
	}
}

func (o *organizationsApiImpl) GetOrganizations(ctx context.Context) (*[]domain.Organization, error) {
	return o.organizationsAPI.GetOrganizations(ctx)
}

func (o *organizationsApiImpl) FindOrganizationByName(ctx context.Context, orgName string) (*domain.Organization, error) {
	return o.organizationsAPI.FindOrganizationByName(ctx, orgName)
}

func (o *organizationsApiImpl) FindOrganizationById(ctx context.Context, orgId string) (*domain.Organization, error) {
	return o.organizationsAPI.FindOrganizationByID(ctx, orgId)
}

func (o *organizationsApiImpl) FindOrganizationsByUserId(ctx context.Context, userID string) (*[]domain.Organization, error) {
	return o.organizationsAPI.FindOrganizationsByUserID(ctx, userID)
}

func (o *organizationsApiImpl) CreateOrganization(ctx context.Context, org *domain.Organization) (*domain.Organization, error) {
	return o.organizationsAPI.CreateOrganization(ctx, org)
}

func (o *organizationsApiImpl) CreateOrganizationWithName(ctx context.Context, orgName string) (*domain.Organization, error) {
	return o.organizationsAPI.CreateOrganizationWithName(ctx, orgName)
}

func (o *organizationsApiImpl) DeleteOrganization(ctx context.Context, org *domain.Organization) error {
	return o.organizationsAPI.DeleteOrganization(ctx, org)
}

func (o *organizationsApiImpl) DeleteOrganizationWithId(ctx context.Context, orgId string) error {
	return o.organizationsAPI.DeleteOrganizationWithID(ctx, orgId)
}

func (o *organizationsApiImpl) UpdateOrganization(ctx context.Context, org *domain.Organization) (*domain.Organization, error) {
	return o.organizationsAPI.UpdateOrganization(ctx, org)
}

func (o *organizationsApiImpl) GetMembers(ctx context.Context, org *domain.Organization) (*[]domain.ResourceMember, error) {
	return o.organizationsAPI.GetMembers(ctx, org)
}

func (o *organizationsApiImpl) GetMembersWithId(ctx context.Context, orgId string) (*[]domain.ResourceMember, error) {
	return o.organizationsAPI.GetMembersWithID(ctx, orgId)
}

func (o *organizationsApiImpl) AddMember(ctx context.Context, org *domain.Organization, user *domain.User) (*domain.ResourceMember, error) {
	return o.organizationsAPI.AddMember(ctx, org, user)
}

func (o *organizationsApiImpl) AddMemberWithId(ctx context.Context, orgId, memberId string) (*domain.ResourceMember, error) {
	return o.organizationsAPI.AddMemberWithID(ctx, orgId, memberId)
}

func (o *organizationsApiImpl) RemoveMember(ctx context.Context, org *domain.Organization, user *domain.User) error {
	return o.organizationsAPI.RemoveMember(ctx, org, user)
}

func (o *organizationsApiImpl) RemoveMemberWithId(ctx context.Context, orgId, memberId string) error {
	return o.organizationsAPI.RemoveMemberWithID(ctx, orgId, memberId)
}

func (o *organizationsApiImpl) GetOwners(ctx context.Context, org *domain.Organization) (*[]domain.ResourceOwner, error) {
	return o.organizationsAPI.GetOwners(ctx, org)
}

func (o *organizationsApiImpl) GetOwnersWithId(ctx context.Context, orgId string) (*[]domain.ResourceOwner, error) {
	return o.organizationsAPI.GetOwnersWithID(ctx, orgId)
}

func (o *organizationsApiImpl) AddOwner(ctx context.Context, org *domain.Organization, user *domain.User) (*domain.ResourceOwner, error) {
	return o.organizationsAPI.AddOwner(ctx, org, user)
}

func (o *organizationsApiImpl) AddOwnerWithId(ctx context.Context, orgId, ownerId string) (*domain.ResourceOwner, error) {
	return o.organizationsAPI.AddOwnerWithID(ctx, orgId, ownerId)
}

func (o *organizationsApiImpl) RemoveOwner(ctx context.Context, org *domain.Organization, user *domain.User) error {
	return o.organizationsAPI.RemoveOwner(ctx, org, user)
}

func (o *organizationsApiImpl) RemoveOwnerWithId(ctx context.Context, orgId, memberId string) error {
	return o.organizationsAPI.RemoveMemberWithID(ctx, orgId, memberId)
}
