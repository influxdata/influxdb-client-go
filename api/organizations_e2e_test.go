// +build e2e

// Copyright 2020-2021 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package api_test

import (
	"context"
	"fmt"
	"testing"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2/domain"
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

	members, err = orgsAPI.GetMembers(ctx, org)
	require.Nil(t, err)
	require.NotNil(t, members)
	require.Len(t, *members, 1)

	org2, err = orgsAPI.FindOrganizationByID(ctx, *org.Id)
	require.Nil(t, err)
	require.NotNil(t, org2)
	assert.Equal(t, org.Name, org2.Name)

	orgs, err := orgsAPI.FindOrganizationsByUserID(ctx, *user.Id)
	require.Nil(t, err)
	require.NotNil(t, orgs)
	require.Len(t, *orgs, 1)
	assert.Equal(t, org.Name, (*orgs)[0].Name)

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

	owner, err := orgsAPI.AddOwner(ctx, org2, user)
	require.Nil(t, err)
	require.NotNil(t, owner)

	owners, err = orgsAPI.GetOwners(ctx, org2)
	require.Nil(t, err)
	require.NotNil(t, owners)
	assert.Len(t, *owners, 2)

	u, err := usersAPI.FindUserByName(ctx, "my-user")
	require.Nil(t, err)
	require.NotNil(t, u)

	err = orgsAPI.RemoveOwnerWithID(ctx, *org2.Id, *u.Id)
	require.Nil(t, err)

	owners, err = orgsAPI.GetOwners(ctx, org2)
	require.Nil(t, err)
	require.NotNil(t, owners)
	assert.Len(t, *owners, 1)

	orgs, err = orgsAPI.FindOrganizationsByUserID(ctx, *user.Id)
	require.Nil(t, err)
	require.NotNil(t, orgs)
	require.Len(t, *orgs, 2)

	err = orgsAPI.RemoveMemberWithID(ctx, *org.Id, *user.Id)
	require.Nil(t, err)

	members, err = orgsAPI.GetMembers(ctx, org)
	require.Nil(t, err)
	require.NotNil(t, members)
	require.Len(t, *members, 0)

	err = orgsAPI.DeleteOrganization(ctx, org)
	require.Nil(t, err)

	err = orgsAPI.DeleteOrganization(ctx, org2)
	assert.Nil(t, err)

	orgList, err = orgsAPI.GetOrganizations(ctx)
	require.Nil(t, err)
	require.NotNil(t, orgList)
	assert.Len(t, *orgList, 1)

	err = usersAPI.DeleteUser(ctx, user)
	require.Nil(t, err)

}

func TestOrganizationAPI_pagination(t *testing.T) {
	ctx := context.Background()
	client := influxdb2.NewClient(serverURL, authToken)

	orgsAPI := client.OrganizationsAPI()

	for i := 0; i < 50; i++ {
		org, err := orgsAPI.CreateOrganizationWithName(ctx, fmt.Sprintf("org-%02d", i+1))
		require.Nil(t, err)
		require.NotNil(t, org)
	}

	orgList, err := orgsAPI.GetOrganizations(ctx)
	require.Nil(t, err)
	require.NotNil(t, orgList)
	require.Len(t, *orgList, 20)

	orgList, err = orgsAPI.GetOrganizations(ctx, api.PagingWithOffset(20))
	require.Nil(t, err)
	require.NotNil(t, orgList)
	require.Len(t, *orgList, 20)

	orgList, err = orgsAPI.GetOrganizations(ctx, api.PagingWithOffset(40))
	require.Nil(t, err)
	require.NotNil(t, orgList)
	require.Len(t, *orgList, 11)

	orgList, err = orgsAPI.GetOrganizations(ctx, api.PagingWithLimit(100))
	require.Nil(t, err)
	require.NotNil(t, orgList)
	require.Len(t, *orgList, 51)

	for _, org := range *orgList {
		if org.Name == "my-org" {
			continue
		}
		err = orgsAPI.DeleteOrganization(ctx, &org)
		assert.Nil(t, err)
	}

}

func TestOrganizationAPI_failing(t *testing.T) {
	client := influxdb2.NewClient(serverURL, authToken)
	orgsAPI := client.OrganizationsAPI()
	ctx := context.Background()
	notExistingID := "aaaaaaaaaaaaaaaa"
	invalidID := "aaaaaa"

	org, err := orgsAPI.FindOrganizationByName(ctx, "my-org")
	require.Nil(t, err, err)
	require.NotNil(t, org)

	// Add member with invalid id
	member, err := orgsAPI.AddMemberWithID(ctx, *org.Id, notExistingID)
	assert.NotNil(t, err)
	assert.Nil(t, member)

	// get member with invalid id
	members, err := orgsAPI.GetMembersWithID(ctx, invalidID)
	assert.NotNil(t, err)
	assert.Nil(t, members)

	// get member with not existing id
	members, err = orgsAPI.GetMembersWithID(ctx, notExistingID)
	assert.NotNil(t, err)
	assert.Nil(t, members)

	//get owners with invalid id
	owners, err := orgsAPI.GetOwnersWithID(ctx, invalidID)
	assert.NotNil(t, err)
	assert.Nil(t, owners)

	//get owners with not existing id
	owners, err = orgsAPI.GetOwnersWithID(ctx, notExistingID)
	assert.NotNil(t, err)
	assert.Nil(t, owners)

	// look for not existent
	orgs, err := orgsAPI.FindOrganizationsByUserID(ctx, notExistingID)
	assert.Nil(t, err)
	assert.NotNil(t, orgs)
	assert.Len(t, *orgs, 0)

	// look for not existent - bug returns current org: https://github.com/influxdata/influxdb/issues/19110
	//orgs, err = orgsAPI.FindOrganizationsByUserID(ctx, invalidID)
	//assert.NotNil(t, err)
	//assert.Nil(t, orgs)

	// add owner with invalid ID
	owner, err := orgsAPI.AddOwnerWithID(ctx, *org.Id, notExistingID)
	assert.NotNil(t, err)
	assert.Nil(t, owner)

	// update with not existing id
	org.Id = &notExistingID
	org, err = orgsAPI.UpdateOrganization(ctx, org)
	assert.NotNil(t, err)
	assert.Nil(t, org)

	// remove owner with invalid ID
	err = orgsAPI.RemoveOwnerWithID(ctx, notExistingID, notExistingID)
	assert.NotNil(t, err)

	// find invalid id
	org, err = orgsAPI.FindOrganizationByID(ctx, notExistingID)
	assert.NotNil(t, err)
	assert.Nil(t, org)

	// remove invalid memberID
	err = orgsAPI.RemoveMemberWithID(ctx, notExistingID, notExistingID)
	assert.NotNil(t, err)

	// delete not existent org
	err = orgsAPI.DeleteOrganizationWithID(ctx, notExistingID)
	assert.NotNil(t, err)

	// delete invalid org
	err = orgsAPI.DeleteOrganizationWithID(ctx, invalidID)
	assert.NotNil(t, err)
}

func TestOrganizationAPI_requestFailing(t *testing.T) {
	client := influxdb2.NewClient("serverURL", authToken)
	orgsAPI := client.OrganizationsAPI()
	ctx := context.Background()

	anID := "aaaaaaaaaaaaaaaa"
	invalidID := "aaaaaa"

	org := &domain.Organization{Id: &anID}

	_, err := orgsAPI.GetOrganizations(ctx)
	assert.NotNil(t, err)

	_, err = orgsAPI.FindOrganizationByName(ctx, "my-org")
	assert.NotNil(t, err)

	_, err = orgsAPI.FindOrganizationByID(ctx, anID)
	assert.NotNil(t, err)

	_, err = orgsAPI.FindOrganizationsByUserID(ctx, anID)
	assert.NotNil(t, err)

	_, err = orgsAPI.CreateOrganizationWithName(ctx, "name")
	assert.NotNil(t, err)

	err = orgsAPI.DeleteOrganizationWithID(ctx, anID)
	assert.NotNil(t, err)

	_, err = orgsAPI.UpdateOrganization(ctx, org)
	assert.NotNil(t, err)

	_, err = orgsAPI.GetMembersWithID(ctx, invalidID)
	assert.NotNil(t, err)

	_, err = orgsAPI.AddMemberWithID(ctx, *org.Id, anID)
	assert.NotNil(t, err)

	// remove invalid memberID
	err = orgsAPI.RemoveMemberWithID(ctx, anID, anID)
	assert.NotNil(t, err)

	_, err = orgsAPI.GetOwnersWithID(ctx, invalidID)
	assert.NotNil(t, err)

	_, err = orgsAPI.AddOwnerWithID(ctx, *org.Id, anID)
	assert.NotNil(t, err)

	// remove owner with invalid ID
	err = orgsAPI.RemoveOwnerWithID(ctx, anID, anID)
	assert.NotNil(t, err)

	_, err = orgsAPI.GetOwnersWithID(ctx, invalidID)
	assert.NotNil(t, err)
}
