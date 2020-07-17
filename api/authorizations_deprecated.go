// Copyright 2020 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package api

import (
	"context"
	"github.com/influxdata/influxdb-client-go/domain"
)

//lint:file-ignore ST1003 This is deprecated API to be removed in next release

// AuthorizationsApi provides methods for organizing Authorization in a InfluxDB server
// Deprecated: Use AuthorizationsAPI instead
type AuthorizationsApi interface {
	// GetAuthorizations returns all authorizations
	GetAuthorizations(ctx context.Context) (*[]domain.Authorization, error)
	// FindAuthorizationsByUserName returns all authorizations for given userName
	FindAuthorizationsByUserName(ctx context.Context, userName string) (*[]domain.Authorization, error)
	// FindAuthorizationsByUserId returns all authorizations for given userID
	FindAuthorizationsByUserId(ctx context.Context, userId string) (*[]domain.Authorization, error)
	// FindAuthorizationsByOrgName returns all authorizations for given organization name
	FindAuthorizationsByOrgName(ctx context.Context, orgName string) (*[]domain.Authorization, error)
	// FindAuthorizationsByUserId returns all authorizations for given organization id
	FindAuthorizationsByOrgId(ctx context.Context, orgId string) (*[]domain.Authorization, error)
	// CreateAuthorization creates new authorization
	CreateAuthorization(ctx context.Context, authorization *domain.Authorization) (*domain.Authorization, error)
	// CreateAuthorizationWithOrgId creates new authorization with given permissions scoped to given orgId
	CreateAuthorizationWithOrgId(ctx context.Context, orgId string, permissions []domain.Permission) (*domain.Authorization, error)
	// UpdateAuthorizationStatus updates status of authorization with authId
	UpdateAuthorizationStatus(ctx context.Context, authId string, status domain.AuthorizationUpdateRequestStatus) (*domain.Authorization, error)
	// DeleteAuthorization deletes authorization with authId
	DeleteAuthorization(ctx context.Context, authId string) error
}

type authorizationsApiImpl struct {
	authorizationsAPI AuthorizationsAPI
}

// NewAuthorizationsApi creates instance of AuthorizationsApi
// Deprecated: Use NewAuthorizationsAPI instead.
func NewAuthorizationsApi(apiClient *domain.ClientWithResponses) AuthorizationsApi {
	return &authorizationsApiImpl{
		authorizationsAPI: NewAuthorizationsAPI(apiClient),
	}
}

func (a *authorizationsApiImpl) GetAuthorizations(ctx context.Context) (*[]domain.Authorization, error) {
	return a.authorizationsAPI.GetAuthorizations(ctx)
}

func (a *authorizationsApiImpl) FindAuthorizationsByUserName(ctx context.Context, userName string) (*[]domain.Authorization, error) {
	return a.authorizationsAPI.FindAuthorizationsByUserName(ctx, userName)
}

func (a *authorizationsApiImpl) FindAuthorizationsByUserId(ctx context.Context, userId string) (*[]domain.Authorization, error) {
	return a.authorizationsAPI.FindAuthorizationsByUserID(ctx, userId)
}

func (a *authorizationsApiImpl) FindAuthorizationsByOrgName(ctx context.Context, orgName string) (*[]domain.Authorization, error) {
	return a.authorizationsAPI.FindAuthorizationsByOrgName(ctx, orgName)
}

func (a *authorizationsApiImpl) FindAuthorizationsByOrgId(ctx context.Context, orgId string) (*[]domain.Authorization, error) {
	return a.authorizationsAPI.FindAuthorizationsByOrgID(ctx, orgId)
}

func (a *authorizationsApiImpl) CreateAuthorization(ctx context.Context, authorization *domain.Authorization) (*domain.Authorization, error) {
	return a.authorizationsAPI.CreateAuthorization(ctx, authorization)
}

func (a *authorizationsApiImpl) CreateAuthorizationWithOrgId(ctx context.Context, orgId string, permissions []domain.Permission) (*domain.Authorization, error) {
	return a.authorizationsAPI.CreateAuthorizationWithOrgID(ctx, orgId, permissions)
}

func (a *authorizationsApiImpl) UpdateAuthorizationStatus(ctx context.Context, authId string, status domain.AuthorizationUpdateRequestStatus) (*domain.Authorization, error) {
	return a.authorizationsAPI.UpdateAuthorizationStatusWithID(ctx, authId, status)
}

func (a *authorizationsApiImpl) DeleteAuthorization(ctx context.Context, authId string) error {
	return a.authorizationsAPI.DeleteAuthorizationWithID(ctx, authId)
}
