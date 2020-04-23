package api

import (
	"context"
	"github.com/influxdata/influxdb-client-go/domain"
	ihttp "github.com/influxdata/influxdb-client-go/internal/http"
)

// OrganizationsApi provides methods for managing Organizations in a InfluxDB server
type OrganizationsApi interface {
	// FindOrganizationByName returns all organizations
	FindOrganizations(ctx context.Context) (*[]domain.Organization, error)
	// FindOrganizationByName returns an organization found using orgNme
	FindOrganizationByName(ctx context.Context, orgName string) (*domain.Organization, error)
	// FindOrganizationByName returns an organization found using orgId
	FindOrganizationById(ctx context.Context, orgId string) (*domain.Organization, error)
	// CreateOrganization creates new organization
	CreateOrganization(ctx context.Context, org *domain.Organization) (*domain.Organization, error)
	// CreateOrganizationWithName creates new organization with orgName and with status active
	CreateOrganizationWithName(ctx context.Context, orgName string) (*domain.Organization, error)
	// UpdateOrganization updates organization
	UpdateOrganization(ctx context.Context, org *domain.Organization) (*domain.Organization, error)
	// DeleteOrganization deletes an organization
	DeleteOrganization(ctx context.Context, org *domain.Organization) error
	// DeleteOrganizationWithID deletes an organization with orgId
	DeleteOrganizationWithID(ctx context.Context, orgId string) error
	// GetMembers returns members of an organization
	GetMembers(ctx context.Context, org *domain.Organization) (*[]domain.ResourceMember, error)
	// GetMembersWithID returns members of an organization with orgId
	GetMembersWithID(ctx context.Context, orgId string) (*[]domain.ResourceMember, error)
	// AddMember add a user to an organization
	AddMember(ctx context.Context, org *domain.Organization, user *domain.User) (*domain.ResourceMember, error)
	// AddMember add a member with id memberId to an organization with orgId
	AddMemberWithIDs(ctx context.Context, orgId, memberId string) (*domain.ResourceMember, error)
	// RemoveMember add a user from an organization
	RemoveMember(ctx context.Context, org *domain.Organization, user *domain.User) error
	// RemoveMember removes a member with id memberId from an organization with orgId
	RemoveMemberWithIDs(ctx context.Context, orgId, memberId string) error
}

type organizationsApiImpl struct {
	apiClient *domain.ClientWithResponses
}

func NewOrganizationsApi(service ihttp.Service) OrganizationsApi {

	apiClient := domain.NewClientWithResponses(service)
	return &organizationsApiImpl{
		apiClient: apiClient,
	}
}

func (o *organizationsApiImpl) FindOrganizations(ctx context.Context) (*[]domain.Organization, error) {
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
	return o.DeleteOrganizationWithID(ctx, *org.Id)
}

func (o *organizationsApiImpl) DeleteOrganizationWithID(ctx context.Context, orgId string) error {
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
	return o.GetMembersWithID(ctx, *org.Id)
}

func (o *organizationsApiImpl) GetMembersWithID(ctx context.Context, orgId string) (*[]domain.ResourceMember, error) {
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
	return o.AddMemberWithIDs(ctx, *org.Id, *user.Id)
}

func (o *organizationsApiImpl) AddMemberWithIDs(ctx context.Context, orgId, memberId string) (*domain.ResourceMember, error) {
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
	return o.RemoveMemberWithIDs(ctx, *org.Id, *user.Id)
}

func (o *organizationsApiImpl) RemoveMemberWithIDs(ctx context.Context, orgId, memberId string) error {
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
