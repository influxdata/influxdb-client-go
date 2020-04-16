package api

import (
	"context"
	"github.com/influxdata/influxdb-client-go/domain"
	ihttp "github.com/influxdata/influxdb-client-go/internal/http"
)

type AuthorizationsApi interface {
	ListAuthorizations(ctx context.Context, query *domain.GetAuthorizationsParams) (*domain.Authorizations, error)
	CreateAuthorization(ctx context.Context, authorization *domain.Authorization) (*domain.Authorization, error)
}

type authorizationsApiImpl struct {
	apiClient *domain.ClientWithResponses
	service   ihttp.Service
}

func NewAuthorizationApi(service ihttp.Service) AuthorizationsApi {

	apiClient := domain.NewClientWithResponses(service)
	return &authorizationsApiImpl{
		apiClient: apiClient,
		service:   service,
	}
}

func (a *authorizationsApiImpl) ListAuthorizations(ctx context.Context, query *domain.GetAuthorizationsParams) (*domain.Authorizations, error) {
	if query == nil {
		query = &domain.GetAuthorizationsParams{}
	}
	response, err := a.apiClient.GetAuthorizationsWithResponse(ctx, query)
	if err != nil {
		return nil, err
	}
	if response.JSONDefault != nil {
		return nil, &ihttp.Error{
			StatusCode: response.HTTPResponse.StatusCode,
			Code:       string(response.JSONDefault.Code),
			Message:    response.JSONDefault.Message,
		}
	}
	return response.JSON200, nil
}

func (a *authorizationsApiImpl) CreateAuthorization(ctx context.Context, authorization *domain.Authorization) (*domain.Authorization, error) {
	params := &domain.PostAuthorizationsParams{}
	response, err := a.apiClient.PostAuthorizationsWithResponse(ctx, params, domain.PostAuthorizationsJSONRequestBody(*authorization))
	if err != nil {
		return nil, err
	}
	if response.JSONDefault != nil {
		return nil, &ihttp.Error{
			StatusCode: response.HTTPResponse.StatusCode,
			Code:       string(response.JSONDefault.Code),
			Message:    response.JSONDefault.Message,
		}
	}
	if response.JSON400 != nil {
		return nil, &ihttp.Error{
			StatusCode: response.HTTPResponse.StatusCode,
			Code:       string(response.JSON400.Code),
			Message:    response.JSON400.Message,
		}
	}
	return response.JSON201, nil
}
