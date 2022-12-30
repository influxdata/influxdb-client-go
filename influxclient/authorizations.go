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

// AuthorizationsAPI holds methods related to authorization, as found under
// the /authorizations endpoint.
type AuthorizationsAPI struct {
	client *model.Client
}

// newAuthorizationsAPI creates new instance of AuthorizationsAPI
func newAuthorizationsAPI(client *model.Client) *AuthorizationsAPI {
	return &AuthorizationsAPI{client: client}
}

// Find returns all authorizations matching the given filter.
// Supported filters:
//   - OrgName
//   - OrgID
//   - UserName
//   - UserID
func (a *AuthorizationsAPI) Find(ctx context.Context, filter *Filter) ([]model.Authorization, error) {
	return a.getAuthorizations(ctx, filter)
}

// FindOne returns one authorizationsmatching the given filter.
// Supported filters:
//   - OrgName
//   - OrgID
//   - UserName
//   - UserID
func (a *AuthorizationsAPI) FindOne(ctx context.Context, filter *Filter) (*model.Authorization, error) {
	authorizations, err := a.getAuthorizations(ctx, filter)
	if err != nil {
		return nil, err
	}
	if len(authorizations) > 0 {
		return &(authorizations)[0], nil
	}
	return nil, fmt.Errorf("authorization not found")
}

// Create creates a new authorization. The returned Authorization holds the new ID.
func (a *AuthorizationsAPI) Create(ctx context.Context, auth *model.Authorization) (*model.Authorization, error) {
	if auth == nil {
		return nil, fmt.Errorf("auth cannot be nil")
	}
	if auth.Permissions == nil {
		return nil, fmt.Errorf("permissions are required")
	}
	params := &model.PostAuthorizationsAllParams{
		Body: model.PostAuthorizationsJSONRequestBody{
			OrgID:       auth.OrgID,
			UserID:      auth.UserID,
			Permissions: auth.Permissions,
		},
	}
	return a.client.PostAuthorizations(ctx, params)
}

// SetStatus updates authorization status.
func (a *AuthorizationsAPI) SetStatus(ctx context.Context, authID string, status model.AuthorizationUpdateRequestStatus) (*model.Authorization, error) {
	if authID == "" {
		return nil, fmt.Errorf("authID is required")
	}
	params := &model.PatchAuthorizationsIDAllParams{
		AuthID: authID,
		Body: model.PatchAuthorizationsIDJSONRequestBody{
			Status: &status,
		},
	}
	return a.client.PatchAuthorizationsID(ctx, params)
}

// Delete deletes the organization with the given ID.
func (a *AuthorizationsAPI) Delete(ctx context.Context, authID string) error {
	if authID == "" {
		return fmt.Errorf("authID is required")
	}
	params := &model.DeleteAuthorizationsIDAllParams{
		AuthID: authID,
	}
	return a.client.DeleteAuthorizationsID(ctx, params)
}

// getAuthorizations create request for GET on /authorizations  according to the filter and validates returned structure
func (a *AuthorizationsAPI) getAuthorizations(ctx context.Context, filter *Filter) ([]model.Authorization, error) {
	params := &model.GetAuthorizationsParams{}
	if filter != nil {
		if filter.OrgName != "" {
			params.Org = &filter.OrgName
		}
		if filter.OrgID != "" {
			params.OrgID = &filter.OrgID
		}
		if filter.UserName != "" {
			params.User = &filter.UserName
		}
		if filter.UserID != "" {
			params.UserID = &filter.UserID
		}
	}
	response, err := a.client.GetAuthorizations(ctx, params)
	if err != nil {
		return nil, err
	}
	return *response.Authorizations, nil
}
