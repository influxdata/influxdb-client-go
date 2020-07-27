// +build e2e

// Copyright 2020 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package api_test

import (
	"context"
	"testing"

	influxdb2 "github.com/influxdata/influxdb-client-go"
	"github.com/influxdata/influxdb-client-go/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrganizationsAPI(t *testing.T) {
	client := influxdb2.NewClient(serverURL, authToken)
	orgsAPI := client.OrganizationsAPI()
	usersAPI := client.UsersAPI()
	orgName := "my-org-2"
	orgDescription := "my-org 2 description"
	ctx := context.Background()
	invalidID := "aaaaaaaaaaaaaaaa"

	orgList, err := orgsAPI.GetOrganizations(ctx)
	require.Nil(t, err)
	require.NotNil(t, orgList)
	assert.Len(t, *orgList, 1)

	//test error
	org, err := orgsAPI.CreateOrganizationWithName(ctx, "")
	assert.NotNil(t, err)
	require.Nil(t, org)

	org, err = orgsAPI.CreateOrganizationWithName(ctx, orgName)
	require.Nil(t, err)
	require.NotNil(t, org)
	assert.Equal(t, orgName, org.Name)

	//test duplicit org
	_, err = orgsAPI.CreateOrganizationWithName(ctx, orgName)
	require.NotNil(t, err)

	org.Description = &orgDescription

	org, err = orgsAPI.UpdateOrganization(ctx, org)
	require.Nil(t, err)
	require.NotNil(t, org)
	assert.Equal(t, orgDescription, *org.Description)

	orgList, err = orgsAPI.GetOrganizations(ctx)

	require.Nil(t, err)
	require.NotNil(t, orgList)
	assert.Len(t, *orgList, 2)

	permission := &domain.Permission{
		Action: domain.PermissionActionWrite,
		Resource: domain.Resource{
			Type: domain.ResourceTypeBuckets,
		},
	}
	permissions := []domain.Permission{*permission}

	//create authorization for new org
	auth, err := client.AuthorizationsAPI().CreateAuthorizationWithOrgID(context.Background(), *org.Id, permissions)
	require.Nil(t, err)
	require.NotNil(t, auth)

	// create client with new auth token without permission
	clientOrg2 := influxdb2.NewClient(serverURL, *auth.Token)

	orgList, err = clientOrg2.OrganizationsAPI().GetOrganizations(ctx)
	require.Nil(t, err)
	require.NotNil(t, orgList)
	assert.Len(t, *orgList, 0)

	org2, err := orgsAPI.FindOrganizationByName(ctx, orgName)
	require.Nil(t, err)
	require.NotNil(t, org2)

	//find unknown org
	org2, err = orgsAPI.FindOrganizationByName(ctx, "not-existetn-org")
	assert.NotNil(t, err)
	assert.Nil(t, org2)

	//find org using token without org permission
	org2, err = clientOrg2.OrganizationsAPI().FindOrganizationByName(ctx, org.Name)
	assert.NotNil(t, err)
	assert.Nil(t, org2)

	err = client.AuthorizationsAPI().DeleteAuthorization(ctx, auth)
	require.Nil(t, err)

	members, err := orgsAPI.GetMembers(ctx, org)
	require.Nil(t, err)
	require.NotNil(t, members)
	require.Len(t, *members, 0)

	user, err := usersAPI.CreateUserWithName(ctx, "user-01")
	require.Nil(t, err)
	require.NotNil(t, user)

	member, err := orgsAPI.AddMember(ctx, org, user)
	require.Nil(t, err)
	require.NotNil(t, member)
	assert.Equal(t, *user.Id, *member.Id)
	assert.Equal(t, user.Name, member.Name)

	// Add member with invalid id
	member, err = orgsAPI.AddMemberWithID(ctx, *org.Id, invalidID)
	assert.NotNil(t, err)
	assert.Nil(t, member)

	members, err = orgsAPI.GetMembers(ctx, org)
	require.Nil(t, err)
	require.NotNil(t, members)
	require.Len(t, *members, 1)

	// get member with invalid id
	members, err = orgsAPI.GetMembersWithID(ctx, invalidID)
	assert.NotNil(t, err)
	assert.Nil(t, members)

	org2, err = orgsAPI.FindOrganizationByID(ctx, *org.Id)
	require.Nil(t, err)
	require.NotNil(t, org2)
	assert.Equal(t, org.Name, org2.Name)

	// find invalid id
	org2, err = orgsAPI.FindOrganizationByID(ctx, invalidID)
	assert.NotNil(t, err)
	assert.Nil(t, org2)

	orgs, err := orgsAPI.FindOrganizationsByUserID(ctx, *user.Id)
	require.Nil(t, err)
	require.NotNil(t, orgs)
	require.Len(t, *orgs, 1)
	assert.Equal(t, org.Name, (*orgs)[0].Name)

	// look for not existent
	orgs, err = orgsAPI.FindOrganizationsByUserID(ctx, invalidID)
	assert.Nil(t, err)
	assert.NotNil(t, orgs)
	assert.Len(t, *orgs, 0)

	orgName2 := "my-org-3"

	org2, err = orgsAPI.CreateOrganizationWithName(ctx, orgName2)
	require.Nil(t, err)
	require.NotNil(t, org2)
	assert.Equal(t, orgName2, org2.Name)

	orgList, err = orgsAPI.GetOrganizations(ctx)
	require.Nil(t, err)
	require.NotNil(t, orgList)
	assert.Len(t, *orgList, 3)

	owners, err := orgsAPI.GetOwners(ctx, org2)
	assert.Nil(t, err)
	assert.NotNil(t, owners)
	assert.Len(t, *owners, 1)

	//get owners with invalid id
	owners, err = orgsAPI.GetOwnersWithID(ctx, invalidID)
	assert.NotNil(t, err)
	assert.Nil(t, owners)

	owner, err := orgsAPI.AddOwner(ctx, org2, user)
	require.Nil(t, err)
	require.NotNil(t, owner)

	// add owner with invalid ID
	owner, err = orgsAPI.AddOwnerWithID(ctx, *org2.Id, invalidID)
	assert.NotNil(t, err)
	assert.Nil(t, owner)

	owners, err = orgsAPI.GetOwners(ctx, org2)
	require.Nil(t, err)
	require.NotNil(t, owners)
	assert.Len(t, *owners, 2)

	u, err := usersAPI.FindUserByName(ctx, "my-user")
	require.Nil(t, err)
	require.NotNil(t, u)

	err = orgsAPI.RemoveOwner(ctx, org2, u)
	require.Nil(t, err)

	// remove owner with invalid ID
	err = orgsAPI.RemoveOwnerWithID(ctx, invalidID, invalidID)
	assert.NotNil(t, err)

	owners, err = orgsAPI.GetOwners(ctx, org2)
	require.Nil(t, err)
	require.NotNil(t, owners)
	assert.Len(t, *owners, 1)

	orgs, err = orgsAPI.FindOrganizationsByUserID(ctx, *user.Id)
	require.Nil(t, err)
	require.NotNil(t, orgs)
	require.Len(t, *orgs, 2)

	err = orgsAPI.RemoveMember(ctx, org, user)
	require.Nil(t, err)

	// remove invalid memberID
	err = orgsAPI.RemoveMemberWithID(ctx, invalidID, invalidID)
	assert.NotNil(t, err)

	members, err = orgsAPI.GetMembers(ctx, org)
	require.Nil(t, err)
	require.NotNil(t, members)
	require.Len(t, *members, 0)

	err = orgsAPI.DeleteOrganization(ctx, org)
	require.Nil(t, err)

	err = orgsAPI.DeleteOrganization(ctx, org2)
	assert.Nil(t, err)

	// delete invalid org
	err = orgsAPI.DeleteOrganizationWithID(ctx, invalidID)
	assert.NotNil(t, err)

	orgList, err = orgsAPI.GetOrganizations(ctx)
	require.Nil(t, err)
	require.NotNil(t, orgList)
	assert.Len(t, *orgList, 1)

	err = usersAPI.DeleteUser(ctx, user)
	require.Nil(t, err)

}
