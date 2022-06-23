//go:build e2e
// +build e2e

// Copyright 2021 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package influxclient_test

import (
	"testing"

	. "github.com/influxdata/influxdb-client-go/influxclient"
	"github.com/influxdata/influxdb-client-go/influxclient/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthorizationsAPI(t *testing.T) {
	client, ctx := newClient(t)
	authAPI := client.AuthorizationsAPI()

	auths, err := authAPI.Find(ctx, nil)
	require.NoError(t, err)
	require.NotNil(t, auths)
	assert.Len(t, auths, 1)

	auths, err = authAPI.Find(ctx, &Filter{})
	require.NoError(t, err)
	require.NotNil(t, auths)
	assert.Len(t, auths, 1) // only oboarded should exist

	authone, err := authAPI.FindOne(ctx, nil)
	require.NoError(t, err)
	require.NotNil(t, authone) // oboarded should exist

	org, err := client.OrganizationAPI().FindOne(ctx, &Filter{
		Name: orgName,
	})
	require.NoError(t, err)
	require.NotNil(t, org)
	assert.Equal(t, orgName, org.Name)

	permissions := []model.Permission{
		{
			Action: model.PermissionActionWrite,
			Resource: model.Resource{
				Type: model.ResourceTypeBuckets,
			},
		},
	}

	auth, err := authAPI.Create(ctx, &model.Authorization{
		OrgID:       org.Id,
		Permissions: &permissions,
	})
	require.NoError(t, err)
	require.NotNil(t, auth)
	defer authAPI.Delete(ctx, safeId(auth.Id))
	assert.Equal(t, model.AuthorizationUpdateRequestStatusActive, *auth.Status)

	auths, err = authAPI.Find(ctx, nil)
	require.NoError(t, err)
	require.NotNil(t, auths)
	assert.Len(t, auths, 2)

	auths, err = authAPI.Find(ctx, &Filter{
		UserName: userName,
	})
	require.NoError(t, err)
	require.NotNil(t, auths)
	assert.Len(t, auths, 2)

	auths, err = authAPI.Find(ctx, &Filter{
		OrgID: *org.Id,
	})
	require.NoError(t, err)
	require.NotNil(t, auths)
	assert.Len(t, auths, 2)

	auths, err = authAPI.Find(ctx, &Filter{
		OrgName: orgName,
	})
	require.NoError(t, err)
	require.NotNil(t, auths)
	assert.Len(t, auths, 2)

	user, err := client.UsersAPI().FindOne(ctx, &Filter{
		Name: userName,
	})
	require.NoError(t, err)
	require.NotNil(t, user)

	auths, err = authAPI.Find(ctx, &Filter{
		UserID: *user.Id,
	})
	require.NoError(t, err)
	require.NotNil(t, auths)
	assert.Len(t, auths, 2)

	auths, err = authAPI.Find(ctx, &Filter{
		OrgName: "not-existent-org",
	})
	require.Error(t, err)
	require.Nil(t, auths)

	auth, err = authAPI.SetStatus(ctx, *auth.Id, model.AuthorizationUpdateRequestStatusInactive) // default is active
	require.NoError(t, err)
	require.NotNil(t, auth)
	assert.Equal(t, model.AuthorizationUpdateRequestStatusInactive, *auth.Status)

	auths, err = authAPI.Find(ctx, nil)
	require.NoError(t, err)
	require.NotNil(t, auths)
	assert.Len(t, auths, 2)

	err = authAPI.Delete(ctx, *auth.Id)
	require.NoError(t, err)

	auths, err = authAPI.Find(ctx, nil)
	require.NoError(t, err)
	require.NotNil(t, auths)
	assert.Len(t, auths, 1) // only oboarded should exist

}

func TestAuthorizationsAPI_failing(t *testing.T) {
	client, ctx := newClient(t)
	authAPI := client.AuthorizationsAPI()

	auths, err := authAPI.Find(ctx, &Filter{
		UserName: "not-existing-user",
	})
	assert.Error(t, err)
	assert.Nil(t, auths)

	auths, err = authAPI.Find(ctx, &Filter{
		UserID: invalidID,
	})
	assert.Error(t, err)
	assert.Nil(t, auths)

	auths, err = authAPI.Find(ctx, &Filter{
		OrgID: notExistingID,
	})
	assert.Error(t, err)
	assert.Nil(t, auths)

	auths, err = authAPI.Find(ctx, &Filter{
		OrgName: "not-existing-org",
	})
	assert.Error(t, err)
	assert.Nil(t, auths)

	authone, err := authAPI.FindOne(ctx, &Filter{
		OrgName: "not-existing-org",
	})
	assert.Error(t, err)
	assert.Nil(t, authone)

	org, err := client.OrganizationAPI().FindOne(ctx, &Filter{
		Name: orgName,
	})
	require.NoError(t, err)
	require.NotNil(t, org)

	auth, err := authAPI.Create(ctx, nil)
	assert.Error(t, err)
	assert.Nil(t, auth)

	auth, err = authAPI.Create(ctx, &model.Authorization{
		OrgID: org.Id,
	})
	assert.Error(t, err)
	assert.Nil(t, auth)

	permissions := []model.Permission{
		{
			Action: model.PermissionActionWrite,
			Resource: model.Resource{
				Type: model.ResourceTypeBuckets,
			},
		},
	}

	auth, err = authAPI.Create(ctx, &model.Authorization{
		OrgID:       &notExistingID,
		Permissions: &permissions,
	})
	assert.Error(t, err)
	assert.Nil(t, auth)

	auth, err = authAPI.SetStatus(ctx, notInitializedID, model.AuthorizationUpdateRequestStatusInactive)
	assert.Error(t, err)
	assert.Nil(t, auth)

	auth, err = authAPI.SetStatus(ctx, notExistingID, model.AuthorizationUpdateRequestStatusInactive)
	assert.Error(t, err)
	assert.Nil(t, auth)

	err = authAPI.Delete(ctx, notInitializedID)
	assert.Error(t, err)

	err = authAPI.Delete(ctx, notExistingID)
	assert.Error(t, err)
}
