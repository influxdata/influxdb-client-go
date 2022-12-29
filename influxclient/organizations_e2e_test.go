// +build e2e

// Copyright 2020-2021 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package influxclient_test

import (
	"fmt"
	"testing"

	. "github.com/influxdata/influxdb-client-go/influxclient"
	"github.com/influxdata/influxdb-client-go/influxclient/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrganizationsAPI(t *testing.T) {
	client, ctx := newClient(t)
	orgsAPI := client.OrganizationAPI()
	usersAPI := client.UsersAPI()

	// find onboarded orgs
	orgs, err := orgsAPI.Find(ctx, nil)
	require.NoError(t, err)
	require.NotNil(t, orgs)
	assert.Len(t, orgs, 1)

	// find onboarded org
	org, err := orgsAPI.FindOne(ctx, nil)
	require.NoError(t, err)
	require.NotNil(t, orgs)

	user2Name := "user-01"
	org2Name := "my-org-2"
	org2Description := "my-org 2 description"

	// create new org
	org2, err := orgsAPI.Create(ctx, &model.Organization{
		Name: org2Name,
	})
	require.NoError(t, err)
	require.NotNil(t, org2)
	defer orgsAPI.Delete(ctx, safeId(org2.Id))
	assert.Equal(t, org2Name, org2.Name)

	// attempt to to create org with existing name
	_, err = orgsAPI.Create(ctx, &model.Organization{
		Name: org2.Name,
	})
	assert.Error(t, err)

	// update
	org2.Description = &org2Description
	org2, err = orgsAPI.Update(ctx, org2)
	require.NoError(t, err)
	require.NotNil(t, org2)
	assert.Equal(t, org2Description, *org2.Description)

	// find more
	orgs, err = orgsAPI.Find(ctx, nil)
	require.NoError(t, err)
	require.NotNil(t, orgs)
	assert.Len(t, orgs, 2)

	// create authorization for new org
	permissions := []model.Permission{
		{
			Action: model.PermissionActionWrite,
			Resource: model.Resource{
				Type: model.ResourceTypeBuckets,
			},
		},
	}
	auth2, err := client.AuthorizationsAPI().Create(ctx, &model.Authorization{
		OrgID: org2.Id,
		Permissions: &permissions,
	})
	require.NoError(t, err)
	require.NotNil(t, auth2)
	defer client.AuthorizationsAPI().Delete(ctx, safeId(auth2.Id))

	// create client with new auth token without permission
	clientOrg2, err := New(Params{ ServerURL: serverURL, AuthToken: *auth2.Token})
	require.NoError(t, err)

	orgs2, err := clientOrg2.OrganizationAPI().Find(ctx, nil)
	require.NoError(t, err)
	require.NotNil(t, orgs2)
	assert.Len(t, orgs2, 0)

	// find org using token without org permission
	orgx, err := clientOrg2.OrganizationAPI().FindOne(ctx, &Filter{
		Name: org.Name,
	})
	assert.Error(t, err)
	assert.Nil(t, orgx)

	err = client.AuthorizationsAPI().Delete(ctx, *auth2.Id)
	require.NoError(t, err)

	// members
	members, err := orgsAPI.Members(ctx, *org2.Id)
	require.NoError(t, err)
	require.NotNil(t, members)
	require.Len(t, members, 0)

	user, err := usersAPI.Create(ctx, &model.User{
		Name: user2Name,
	})
	require.NoError(t, err)
	require.NotNil(t, user)
	defer usersAPI.Delete(ctx, safeId(user.Id))

	err = orgsAPI.AddMember(ctx, *org2.Id, *user.Id)
	require.NoError(t, err)

	members, err = orgsAPI.Members(ctx, *org2.Id)
	require.NoError(t, err)
	require.NotNil(t, members)
	require.Len(t, members, 1)

	org, err = orgsAPI.FindOne(ctx, &Filter{
		OrgID: *org2.Id,
	})
	require.NoError(t, err)
	require.NotNil(t, org)
	assert.Equal(t, org2.Name, org.Name)

	orgs, err = orgsAPI.Find(ctx, &Filter{
		UserID: *user.Id,
	})
	require.NoError(t, err)
	require.NotNil(t, orgs)
	require.Len(t, orgs, 1)
	assert.Equal(t, org2.Name, orgs[0].Name)

	org3Name := "my-org-3"
	org3, err := orgsAPI.Create(ctx, &model.Organization{
		Name: org3Name,
	})
	require.NoError(t, err)
	require.NotNil(t, org3)
	defer orgsAPI.Delete(ctx, safeId(org3.Id))
	assert.Equal(t, org3Name, org3.Name)

	orgs, err = orgsAPI.Find(ctx, nil)
	require.NoError(t, err)
	require.NotNil(t, orgs)
	assert.Len(t, orgs, 3)

	owners, err := orgsAPI.Owners(ctx, *org3.Id)
	assert.NoError(t, err)
	assert.NotNil(t, owners)
	assert.Len(t, owners, 1)

	err = orgsAPI.AddOwner(ctx, *org3.Id, *user.Id)
	assert.NoError(t, err)

	owners, err = orgsAPI.Owners(ctx, *org3.Id)
	require.NoError(t, err)
	require.NotNil(t, owners)
	assert.Len(t, owners, 2)

	u, err := usersAPI.FindOne(ctx, &Filter{
		Name: user2Name,
	})
	require.NoError(t, err)
	require.NotNil(t, u)

	err = orgsAPI.RemoveOwner(ctx, *org3.Id, *u.Id)
	require.NoError(t, err)

	owners, err = orgsAPI.Owners(ctx, *org3.Id)
	require.NoError(t, err)
	require.NotNil(t, owners)
	assert.Len(t, owners, 1)

	orgs, err = orgsAPI.Find(ctx, &Filter{
		UserID: *user.Id,
	})
	require.NoError(t, err)
	require.NotNil(t, orgs)
	require.Len(t, orgs, 1)

	err = orgsAPI.RemoveMember(ctx, *org2.Id, *user.Id)
	require.NoError(t, err)

	members, err = orgsAPI.Members(ctx, *org2.Id)
	require.NoError(t, err)
	require.NotNil(t, members)
	require.Len(t, members, 0)

	err = usersAPI.Delete(ctx, *user.Id)
	require.NoError(t, err)

	err = orgsAPI.Delete(ctx, *org3.Id)
	require.NoError(t, err)

	err = orgsAPI.Delete(ctx, *org2.Id)
	assert.NoError(t, err)

	orgs, err = orgsAPI.Find(ctx, nil)
	require.NoError(t, err)
	require.NotNil(t, orgs)
	assert.Len(t, orgs, 1)
}

func TestOrganizationAPI_pagination(t *testing.T) {
	client, ctx := newClient(t)
	orgsAPI := client.OrganizationAPI()

	for i := 0; i < 50; i++ {
		org, err := orgsAPI.Create(ctx, &model.Organization{
			Name: fmt.Sprintf("org-%02d", i+1),
		})
		require.NoError(t, err)
		require.NotNil(t, org)
		defer orgsAPI.Delete(ctx, safeId(org.Id))
	}

	orgs, err := orgsAPI.Find(ctx, nil)
	require.NoError(t, err)
	require.NotNil(t, orgs)
	require.Len(t, orgs, 20)

	orgs, err = orgsAPI.Find(ctx, &Filter{
		Offset: 20,
	})
	require.NoError(t, err)
	require.NotNil(t, orgs)
	require.Len(t, orgs, 20)

	orgs, err = orgsAPI.Find(ctx, &Filter{
		Offset: 40,
	})
	require.NoError(t, err)
	require.NotNil(t, orgs)
	require.Len(t, orgs, 11)

	orgs, err = orgsAPI.Find(ctx, &Filter{
		Limit: 100,
	})
	require.NoError(t, err)
	require.NotNil(t, orgs)
	require.Len(t, orgs, 51)
}

func TestOrganizationAPI_failing(t *testing.T) {
	client, ctx := newClient(t)
	orgsAPI := client.OrganizationAPI()

	// try nil input
	org, err := orgsAPI.Create(ctx, nil)
	assert.Error(t, err)
	require.Nil(t, org)

	// try empty org name
	org, err = orgsAPI.Create(ctx, &model.Organization{})
	assert.Error(t, err)
	require.Nil(t, org)

	org, err = orgsAPI.Create(ctx, &model.Organization{
		Name: "failing",
	})
	require.NoError(t, err)
	assert.NotNil(t, org)
	defer orgsAPI.Delete(ctx, safeId(org.Id))

	err = orgsAPI.AddMember(ctx, *org.Id, notExistingID)
	assert.Error(t, err)

	err = orgsAPI.AddMember(ctx, *org.Id, notInitializedID)
	assert.Error(t, err)

	// get members with invalid org id
	members, err := orgsAPI.Members(ctx, invalidID)
	assert.Error(t, err)
	assert.Nil(t, members)

	// get member with not existing id
	members, err = orgsAPI.Members(ctx, notExistingID)
	assert.Error(t, err)
	assert.Nil(t, members)

	members, err = orgsAPI.Members(ctx, notInitializedID)
	assert.Error(t, err)
	assert.Nil(t, members)

	//get owners with invalid id
	owners, err := orgsAPI.Owners(ctx, invalidID)
	assert.Error(t, err)
	assert.Nil(t, owners)

	owners, err = orgsAPI.Owners(ctx, notInitializedID)
	assert.Error(t, err)
	assert.Nil(t, owners)

	//get owners with not existing id
	owners, err = orgsAPI.Owners(ctx, notExistingID)
	assert.Error(t, err)
	assert.Nil(t, owners)

	// add owner with invalid ID
	err = orgsAPI.AddOwner(ctx, *org.Id, notExistingID)
	assert.Error(t, err)

	err = orgsAPI.AddOwner(ctx, notInitializedID, notExistingID)
	assert.Error(t, err)

	err = orgsAPI.AddOwner(ctx, *org.Id, notInitializedID)
	assert.Error(t, err)

	// update with nil input
	_, err = orgsAPI.Update(ctx, nil)
	assert.Error(t, err)

	// update with nil org ID
	_, err = orgsAPI.Update(ctx, &model.Organization{
		Id: nil,
		Name: org.Name,
	})
	assert.Error(t, err)

	// update with empty name
	_, err = orgsAPI.Update(ctx, &model.Organization{
		Id: org.Id,
		Name: "",
	})
	assert.Error(t, err)

	// update with not existing id
	_, err = orgsAPI.Update(ctx, &model.Organization{
		Id: &notExistingID,
		Name: org.Name,
	})
	assert.Error(t, err)

	// remove owner with invalid ID
	err = orgsAPI.RemoveOwner(ctx, notExistingID, notExistingID)
	assert.Error(t, err)

	err = orgsAPI.RemoveOwner(ctx, notInitializedID, notExistingID)
	assert.Error(t, err)

	err = orgsAPI.RemoveOwner(ctx, notExistingID, notInitializedID)
	assert.Error(t, err)

	// remove invalid memberID
	err = orgsAPI.RemoveMember(ctx, notExistingID, notExistingID)
	assert.Error(t, err)

	err = orgsAPI.RemoveMember(ctx, notInitializedID, notExistingID)
	assert.Error(t, err)

	err = orgsAPI.RemoveMember(ctx, notExistingID, notInitializedID)
	assert.Error(t, err)

	// delete not existent org
	err = orgsAPI.Delete(ctx, notExistingID)
	assert.Error(t, err)

	// delete invalid org
	err = orgsAPI.Delete(ctx, invalidID)
	assert.Error(t, err)

	// delete invalid org
	err = orgsAPI.Delete(ctx, notInitializedID)
	assert.Error(t, err)
}

func TestOrganizationAPI_skip(t *testing.T) {
	t.Skip("Should fail but doesn't")

	// https://github.com/influxdata/influxdb/issues/19110

	client, ctx := newClient(t)
	orgsAPI := client.OrganizationAPI()

	// find by not existing org ID
	o, err := orgsAPI.FindOne(ctx, &Filter{
		ID: notExistingID,
	})
	assert.NotNil(t, err, "Should fail when filtering by non-existent ID")
	assert.Nil(t, o, "Should be nil when filtering by non-existent ID")

	// find by not existing user ID
	orgs, err := orgsAPI.Find(ctx, &Filter{
		UserID: notExistingID,
	})
	assert.NotNil(t, err, "Should fail when filtering by non-existent user ID")
	assert.Nil(t, o, "Should be nil when filtering by non-existent user ID")
	assert.Len(t, orgs, 0)
}
