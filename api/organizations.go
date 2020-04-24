package api

import (
	"context"
	"github.com/influxdata/influxdb-client-go/domain"
	ihttp "github.com/influxdata/influxdb-client-go/internal/http"
)

type OrganizationsApi interface {
	// FindOrganizationByName returns organization found using orgNme
	FindOrganizationByName(ctx context.Context, orgName string) (*domain.Organization, error)
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
