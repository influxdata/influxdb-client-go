// +build e2e

// Copyright 2020 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package api_test

import (
	"context"
	influxdb2 "github.com/influxdata/influxdb-client-go"
	"github.com/influxdata/influxdb-client-go/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
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
	serverURL = getEnvValue("INFLUXDB2_URL", "http://localhost:9999")
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
