// +build e2e

// Copyright 2020-2021 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package api_test

import (
	"context"
	"os"
	"testing"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var authToken string
var serverURL string

func getEnvValue(key, defVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	} else {
		return defVal
	}
}

func init() {
	authToken = getEnvValue("INFLUXDB2_TOKEN", "my-token")
	serverURL = getEnvValue("INFLUXDB2_URL", "http://localhost:8086")
}

func TestAuthorizationsAPI(t *testing.T) {
	client := influxdb2.NewClient(serverURL, authToken)
	authAPI := client.AuthorizationsAPI()
	listRes, err := authAPI.GetAuthorizations(context.Background())
	require.Nil(t, err)
	require.NotNil(t, listRes)
	assert.Len(t, *listRes, 1)

	orgName := "my-org"
	org, err := client.OrganizationsAPI().FindOrganizationByName(context.Background(), orgName)
	require.Nil(t, err)
	require.NotNil(t, org)
	assert.Equal(t, orgName, org.Name)

	permission := &domain.Permission{
		Action: domain.PermissionActionWrite,
		Resource: domain.Resource{
			Type: domain.ResourceTypeBuckets,
		},
	}
	permissions := []domain.Permission{*permission}

	auth, err := authAPI.CreateAuthorizationWithOrgID(context.Background(), *org.Id, permissions)
	require.Nil(t, err)
	require.NotNil(t, auth)
	assert.Equal(t, domain.AuthorizationUpdateRequestStatusActive, *auth.Status, *auth.Status)

	listRes, err = authAPI.GetAuthorizations(context.Background())
	require.Nil(t, err)
	require.NotNil(t, listRes)
	assert.Len(t, *listRes, 2)

	listRes, err = authAPI.FindAuthorizationsByUserName(context.Background(), "my-user")
	require.Nil(t, err)
	require.NotNil(t, listRes)
	assert.Len(t, *listRes, 2)

	listRes, err = authAPI.FindAuthorizationsByOrgID(context.Background(), *org.Id)
	require.Nil(t, err)
	require.NotNil(t, listRes)
	assert.Len(t, *listRes, 2)

	listRes, err = authAPI.FindAuthorizationsByOrgName(context.Background(), "my-org")
	require.Nil(t, err)
	require.NotNil(t, listRes)
	assert.Len(t, *listRes, 2)

	user, err := client.UsersAPI().FindUserByName(context.Background(), "my-user")
	require.Nil(t, err)
	require.NotNil(t, user)

	listRes, err = authAPI.FindAuthorizationsByUserID(context.Background(), *user.Id)
	require.Nil(t, err)
	require.NotNil(t, listRes)
	assert.Len(t, *listRes, 2)

	listRes, err = authAPI.FindAuthorizationsByOrgName(context.Background(), "not-existent-org")
	require.Nil(t, listRes)
	require.NotNil(t, err)
	//assert.Len(t, *listRes, 0)

	auth, err = authAPI.UpdateAuthorizationStatus(context.Background(), auth, domain.AuthorizationUpdateRequestStatusInactive)
	require.Nil(t, err)
	require.NotNil(t, auth)
	assert.Equal(t, domain.AuthorizationUpdateRequestStatusInactive, *auth.Status, *auth.Status)

	listRes, err = authAPI.GetAuthorizations(context.Background())
	require.Nil(t, err)
	require.NotNil(t, listRes)
	assert.Len(t, *listRes, 2)

	err = authAPI.DeleteAuthorization(context.Background(), auth)
	require.Nil(t, err)

	listRes, err = authAPI.GetAuthorizations(context.Background())
	require.Nil(t, err)
	require.NotNil(t, listRes)
	assert.Len(t, *listRes, 1)

}

func TestAuthorizationsAPI_Failing(t *testing.T) {
	client := influxdb2.NewClient(serverURL, authToken)
	authAPI := client.AuthorizationsAPI()
	invalidID := "xcv"
	notExistentID := "100000000000000"

	listRes, err := authAPI.FindAuthorizationsByUserName(context.Background(), "invalid-user")
	assert.NotNil(t, err)
	assert.Nil(t, listRes)

	listRes, err = authAPI.FindAuthorizationsByUserID(context.Background(), invalidID)
	assert.NotNil(t, err)
	assert.Nil(t, listRes)

	listRes, err = authAPI.FindAuthorizationsByOrgID(context.Background(), notExistentID)
	assert.NotNil(t, err)
	assert.Nil(t, listRes)

	listRes, err = authAPI.FindAuthorizationsByOrgName(context.Background(), "not-existing-org")
	assert.NotNil(t, err)
	assert.Nil(t, listRes)

	org, err := client.OrganizationsAPI().FindOrganizationByName(context.Background(), "my-org")
	require.Nil(t, err)
	require.NotNil(t, org)

	auth, err := authAPI.CreateAuthorizationWithOrgID(context.Background(), *org.Id, nil)
	assert.NotNil(t, err)
	assert.Nil(t, auth)

	permission := &domain.Permission{
		Action: domain.PermissionActionWrite,
		Resource: domain.Resource{
			Type: domain.ResourceTypeBuckets,
		},
	}
	permissions := []domain.Permission{*permission}

	auth, err = authAPI.CreateAuthorizationWithOrgID(context.Background(), notExistentID, permissions)
	assert.NotNil(t, err)
	assert.Nil(t, auth)

	auth, err = authAPI.UpdateAuthorizationStatusWithID(context.Background(), notExistentID, domain.AuthorizationUpdateRequestStatusInactive)
	assert.NotNil(t, err)
	assert.Nil(t, auth)

	err = authAPI.DeleteAuthorizationWithID(context.Background(), notExistentID)
	assert.NotNil(t, err)
}

func TestAuthorizationsAPI_requestFailing(t *testing.T) {

	client := influxdb2.NewClientWithOptions("htp://localhost:9910", authToken, influxdb2.DefaultOptions().SetHTTPRequestTimeout(1))
	authAPI := client.AuthorizationsAPI()

	listRes, err := authAPI.GetAuthorizations(context.Background())
	assert.NotNil(t, err)
	assert.Nil(t, listRes)

	permission := &domain.Permission{
		Action: domain.PermissionActionWrite,
		Resource: domain.Resource{
			Type: domain.ResourceTypeBuckets,
		},
	}
	permissions := []domain.Permission{*permission}

	auth, err := authAPI.CreateAuthorizationWithOrgID(context.Background(), "1000000000000000000", permissions)
	assert.NotNil(t, err)
	assert.Nil(t, auth)

	auth, err = authAPI.UpdateAuthorizationStatusWithID(context.Background(), "1000000000000000000", domain.AuthorizationUpdateRequestStatusInactive)
	assert.NotNil(t, err)
	assert.Nil(t, auth)

	err = authAPI.DeleteAuthorizationWithID(context.Background(), "1000000000000000000")
	assert.NotNil(t, err)
}
